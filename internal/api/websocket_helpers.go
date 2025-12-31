package api

import (
	"encoding/json"
	"privacy-social-backend/internal/realtime"

	"github.com/google/uuid"
)

// sendWSNotification sends a WebSocket notification to a user
func (server *Server) sendWSNotification(userID uuid.UUID, msgType string, payload interface{}) {
	wsMsg := realtime.WSMessage{
		Type:    msgType,
		Payload: payload,
	}
	wsMsgBytes, _ := json.Marshal(wsMsg)
	server.hub.SendToUser(userID, wsMsgBytes)
}
