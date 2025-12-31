package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8081"

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type UploadResponse struct {
	URL string `json:"url"`
}

type Message struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	MediaUrl  string `json:"media_url"`
	MediaType string `json:"media_type"`
}

func main() {
	// 1. Register User A
	// 1. Register User A
	ts := time.Now().Unix()
	userA := fmt.Sprintf("mediauserA%d", ts)
	userB := fmt.Sprintf("mediauserB%d", ts)
	phoneA := fmt.Sprintf("555%07d", ts%10000000)
	phoneB := fmt.Sprintf("556%07d", ts%10000000)

	tokenA, userA_ID := registerUser(userA, fmt.Sprintf("%s@example.com", userA), phoneA)
	if tokenA == "" {
		panic("User A registration failed")
	}
	fmt.Println("User A registered:", userA)

	// 2. Register User B
	tokenB, userB_ID := registerUser(userB, fmt.Sprintf("%s@example.com", userB), phoneB)
	if tokenB == "" {
		panic("User B registration failed")
	}
	fmt.Println("User B registered:", userB)

	// 3. Connect A and B
	sendConnectionRequest(tokenA, userB_ID)
	acceptConnectionRequest(tokenB, userA) // Need A's Username for comparison logic
	fmt.Println("Users connected")

	// Unused ID check removal
	_ = userA_ID

	// 4. Upload File
	fileUrl := uploadFile(tokenA)
	if fileUrl == "" {
		panic("File upload failed")
	}
	fmt.Println("File uploaded:", fileUrl)

	// 5. Send Message with Media
	sendMediaMessage(tokenA, userB_ID, fileUrl)
	fmt.Println("Message sent")

	// 6. Verify Message Persistence
	if verifyMessage(tokenA, userB_ID, fileUrl) {
		fmt.Println("✅ SUCCESS: Media message verified")
	} else {
		fmt.Println("❌ FAILURE: Media message not verified")
	}
}

func registerUser(username, email, phone string) (string, string) {
	payload := map[string]string{
		"username": username,
		"password": "password123",
	}
	// Try login first (in case exists)
	body, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/users/login", "application/json", bytes.NewBuffer(body))

	// Login Response structure in script needs to match backend
	type LoginUserResponse struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID string `json:"id"`
		} `json:"user"`
	}

	if err == nil && resp.StatusCode == 200 {
		var res LoginUserResponse
		json.NewDecoder(resp.Body).Decode(&res)
		return res.AccessToken, res.User.ID
	}

	payload["full_name"] = username
	payload["email"] = email
	payload["phone"] = phone
	body, _ = json.Marshal(payload)
	resp, _ = http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(body))

	// Login logic again...
	body, _ = json.Marshal(map[string]string{"username": username, "password": "password123", "phone": phone})
	resp, _ = http.Post(baseURL+"/users/login", "application/json", bytes.NewBuffer(body))

	var res LoginUserResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.AccessToken, res.User.ID
}

func searchUser(token, query string) string {
	req, _ := http.NewRequest("GET", baseURL+"/users/search?q="+query, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)
	// Assuming returns list, take first id
	type User struct {
		ID string `json:"id"`
	}
	var users []User
	json.NewDecoder(resp.Body).Decode(&users)
	if len(users) > 0 {
		return users[0].ID
	}
	return ""
}

func sendConnectionRequest(token, targetID string) {
	body, _ := json.Marshal(map[string]string{"target_user_id": targetID})
	req, _ := http.NewRequest("POST", baseURL+"/connections/request", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)
}

func acceptConnectionRequest(token, requesterUsername string) {
	// Get requests
	req, _ := http.NewRequest("GET", baseURL+"/connections/requests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)

	// Flat struct matching DB row
	type ConnRequest struct {
		RequesterID string `json:"requester_id"`
		Username    string `json:"username"`
	}
	var requests []ConnRequest
	json.NewDecoder(resp.Body).Decode(&requests)

	fmt.Printf("DEBUG: Found %d requests. Looking for %s\n", len(requests), requesterUsername)
	for _, r := range requests {
		fmt.Printf("DEBUG: Request from %s\n", r.Username)
		if r.Username == requesterUsername {
			payload := map[string]string{
				"requester_id": r.RequesterID,
				"status":       "accepted",
			}
			body, _ := json.Marshal(payload)
			req2, _ := http.NewRequest("POST", baseURL+"/connections/update", bytes.NewBuffer(body))
			req2.Header.Set("Authorization", "Bearer "+token)
			http.DefaultClient.Do(req2)
			return
		}
	}
}

func uploadFile(token string) string {
	// Create temp file
	tmpfile, _ := ioutil.TempFile("", "test.png")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte("fake image content"))
	tmpfile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.png")

	f, _ := os.Open(tmpfile.Name())
	defer f.Close()
	io.Copy(part, f)
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/upload", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, _ := http.DefaultClient.Do(req)
	var res UploadResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.URL
}

func sendMediaMessage(token, receiverID, url string) {
	payload := map[string]string{
		"receiver_id": receiverID,
		"media_url":   url,
		"media_type":  "image",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/messages", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)
	time.Sleep(1 * time.Second)
}

func verifyMessage(token, partnerID, expectedUrl string) bool {
	req, _ := http.NewRequest("GET", baseURL+"/messages?user_id="+partnerID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)

	var msgs []Message
	json.NewDecoder(resp.Body).Decode(&msgs)

	fmt.Printf("DEBUG: Expected URL: %s\n", expectedUrl)
	fmt.Printf("DEBUG: Retrieved %d messages\n", len(msgs))

	for _, m := range msgs {
		fmt.Printf("MSG ID: %s, MediaUrl: %s, MediaType: %s, Content: %s\n", m.ID, m.MediaUrl, m.MediaType, m.Content)
		if m.MediaUrl == expectedUrl {
			return true
		}
	}
	return false
}
