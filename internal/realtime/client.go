package realtime

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Client represents a connected user
type Client struct {
	Hub      *Hub
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
}

// WSMessage defines the structure of WebSocket messages
type WSMessage struct {
	Type      string      `json:"type"` // "new_message", "typing", etc.
	Payload   interface{} `json:"payload"`
	SenderID  uuid.UUID   `json:"sender_id,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping period
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Write wait
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(4096)                                // Max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Pong wait
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket unexpected close error")
			}
			break
		}

		// Handle typing indicator messages
		var wsMsg struct {
			Type       string    `json:"type"`
			ReceiverID uuid.UUID `json:"receiver_id"`
		}
		if err := json.Unmarshal(message, &wsMsg); err == nil {
			if wsMsg.Type == "typing" {
				// Forward typing indicator to the receiver
				typingMsg := WSMessage{
					Type: "typing",
					Payload: map[string]interface{}{
						"user_id":  c.UserID,
						"username": c.Username,
					},
				}
				typingBytes, _ := json.Marshal(typingMsg)
				c.Hub.SendToUser(wsMsg.ReceiverID, typingBytes)
			}
		}
	}
}
