package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"privacy-social-backend/internal/realtime"
	"privacy-social-backend/internal/token"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In debug mode, allow all origins for development
		if gin.Mode() == gin.DebugMode {
			return true
		}

		// In production, check against allowed origins
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:3000",

			"https://yourdomain.com", // Add your production domain
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// chatWebSocket handles WebSocket connections for real-time chat
func (server *Server) chatWebSocket(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Upgrade HTTP to WS
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set websocket upgrade")
		return
	}

	client := &realtime.Client{
		Hub:      server.hub,
		UserID:   authPayload.UserID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Username: authPayload.Username,
	}

	server.hub.Register <- client

	// Start pumps in goroutines
	go client.WritePump()
	go client.ReadPump()
}
