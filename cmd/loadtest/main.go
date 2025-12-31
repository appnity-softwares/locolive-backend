package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Constants
const (
	BaseURL = "https://localhost:8080"
)

// Configuration flags
var (
	numUsers    = flag.Int("users", 100, "Number of concurrent users")
	duration    = flag.Duration("duration", 30*time.Second, "Test duration")
	requestRate = flag.Int("rate", 10, "Requests per second per user (approx)")
)

// Global Stats
var (
	totalRequests  int64
	successfulReqs int64
	failedReqs     int64
	totalLatency   int64 // Microseconds
)

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	} `json:"user"`
}

type FeedResponse struct {
	Count int `json:"count"`
}

func main() {
	flag.Parse()
	fmt.Printf("🚀 Starting Load Test with %d users for %v...\n", *numUsers, *duration)

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
	start := time.Now()

	// Rate limiter channel
	// A simple preventive measure, though strictly speaking each user loops independently.

	// Create N users
	for i := 0; i < *numUsers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runUser(id)
		}(i)
		time.Sleep(200 * time.Millisecond) // Stagger login to avoid 429
	}

	wg.Wait()
	elapsed := time.Since(start)

	printStats(elapsed)
}

func runUser(id int) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}

	// 1. Register/Login (Simplified: Create a random user)
	// For load testing reliability without polluting DB forever,
	// ideally we use a set of existing test users or clean them up.
	// Here we just create one random user per thread.
	username := fmt.Sprintf("loaduser%d%d", id, rand.Intn(100000))
	password := "password123"
	phone := fmt.Sprintf("+1555%06d", rand.Intn(1000000))

	// Register
	regPayload := map[string]string{
		"username":  username,
		"password":  password,
		"full_name": "Load Test User",
		"email":     fmt.Sprintf("%s@example.com", username),
		"phone":     phone,
	}
	// We might fail if user exists, so ignore error and try login
	_, rCode, _ := postJSON(client, "/users", regPayload, "")
	if rCode >= 400 {
		// Ignore
	}

	// Login
	loginPayload := map[string]string{
		"phone":    phone,
		"password": password,
	}
	respBody, code, err := postJSON(client, "/users/login", loginPayload, "")
	if err != nil || code != 200 {
		fmt.Printf("User %d failed to login: %v (Code: %d)\n", id, err, code)
		return
	}

	var loginResp LoginResponse
	json.Unmarshal(respBody, &loginResp)
	token := loginResp.AccessToken

	// Main Loop
	endTime := time.Now().Add(*duration)
	for time.Now().Before(endTime) {
		// Action: Get Feed
		// Simulate random location around a center point (e.g. San Francisco)
		// Lat: 37.7749, Lng: -122.4194
		lat := 37.7749 + (rand.Float64()-0.5)*0.1 // +/- 0.05 degrees (~5km)
		lng := -122.4194 + (rand.Float64()-0.5)*0.1

		url := fmt.Sprintf("/feed?latitude=%f&longitude=%f", lat, lng)

		reqStart := time.Now()
		_, code, err := get(client, url, token)
		latency := time.Since(reqStart).Microseconds()

		atomic.AddInt64(&totalRequests, 1)
		atomic.AddInt64(&totalLatency, latency)

		if err == nil && code == 200 {
			atomic.AddInt64(&successfulReqs, 1)
		} else {
			if atomic.LoadInt64(&failedReqs) == 0 {
				fmt.Printf("First failure: Code=%d, Err=%v\n", code, err)
			}
			atomic.AddInt64(&failedReqs, 1)
		}

		// Sleep a bit to match rate
		time.Sleep(time.Duration(1000 / *requestRate) * time.Millisecond)
	}
}

func postJSON(client *http.Client, path string, data interface{}, token string) ([]byte, int, error) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", BaseURL+path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func get(client *http.Client, path string, token string) ([]byte, int, error) {
	req, _ := http.NewRequest("GET", BaseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func printStats(elapsed time.Duration) {
	total := atomic.LoadInt64(&totalRequests)
	success := atomic.LoadInt64(&successfulReqs)
	failed := atomic.LoadInt64(&failedReqs)
	totalLat := atomic.LoadInt64(&totalLatency)

	fmt.Println("\n📊 Load Test Results")
	fmt.Println("====================")
	fmt.Printf("Duration:    %v\n", elapsed)
	fmt.Printf("Total Reqs:  %d\n", total)
	fmt.Printf("Success:     %d\n", success)
	fmt.Printf("Failed:      %d\n", failed)
	if total > 0 {
		fmt.Printf("Avg Latency: %.2f ms\n", float64(totalLat)/float64(total)/1000.0)
		fmt.Printf("RPS:         %.2f\n", float64(total)/elapsed.Seconds())
	}
}
