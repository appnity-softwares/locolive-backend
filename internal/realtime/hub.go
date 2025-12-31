package realtime

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	streamKey = "locolive:stream:routing"
)

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.RWMutex
	redis      *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[uuid.UUID]map[*Client]bool),
		redis:      rdb,
	}
}

func (h *Hub) Run() {
	// Start consuming Redis Stream messages
	go h.listenRedisStream()

	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mutex.Unlock()
			log.Info().Str("username", client.Username).Msg("Client registered")

		case client := <-h.Unregister:
			h.mutex.Lock()
			if userClients, ok := h.clients[client.UserID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					close(client.Send)
					if len(userClients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mutex.Unlock()
			log.Info().Str("username", client.Username).Msg("Client unregistered")
		}
	}
}

// listenRedisStream pumps messages from Redis Stream to local clients
func (h *Hub) listenRedisStream() {
	// Start reading from the end of the stream ($)
	lastID := "$"

	for {
		// Block for up to 2 seconds waiting for new messages
		streams, err := h.redis.XRead(context.Background(), &redis.XReadArgs{
			Streams: []string{streamKey, lastID},
			Count:   10,
			Block:   2000 * time.Millisecond,
		}).Result()

		if err == redis.Nil {
			continue // No new messages
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to read from Redis Stream")
			time.Sleep(5 * time.Second) // Backoff on error
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				lastID = msg.ID

				targetUserIDStr, ok := msg.Values["target_user_id"].(string)
				if !ok {
					continue
				}
				payload, ok := msg.Values["payload"].(string)
				if !ok {
					continue
				}

				userID, err := uuid.Parse(targetUserIDStr)
				if err != nil {
					continue
				}

				h.broadcastToLocal(userID, []byte(payload))
			}
		}
	}
}

// broadcastToLocal sends a message ONLY to locally connected clients
func (h *Hub) broadcastToLocal(userID uuid.UUID, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

// SendToUser writes a message to the Redis Stream.
// This ensures that ANY server instance holding the user's connection receives it.
func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	// Add message to the stream
	// We use "*" to let Redis generate the ID
	// We define fields "target_user_id" and "payload"
	err := h.redis.XAdd(context.Background(), &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"target_user_id": userID.String(),
			"payload":        string(message),
		},
		// Optional: Cap the stream approx length to prevent infinite growth
		MaxLen: 100000,
		Approx: true,
	}).Err()

	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to publish message to Redis Stream")
	}
}

// IsUserOnline checks if a user has any active connections (Local check only for now)
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}
