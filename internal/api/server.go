package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"privacy-social-backend/internal/config"
	"privacy-social-backend/internal/realtime"
	"privacy-social-backend/internal/repository"
	"privacy-social-backend/internal/service/admin"
	"privacy-social-backend/internal/service/location"
	"privacy-social-backend/internal/service/safety"
	"privacy-social-backend/internal/service/storage"
	"privacy-social-backend/internal/service/story"
	"privacy-social-backend/internal/service/user"
	"privacy-social-backend/internal/token"
)

// Server serves HTTP requests for our privacy social service
type Server struct {
	config     config.Config
	store      repository.Store
	tokenMaker token.Maker
	redis      *redis.Client
	router     *gin.Engine
	hub        *realtime.Hub
	safety     *safety.Monitor
	location   *location.RedisLocationService
	story      story.Service
	user       user.Service
	admin      admin.Service
	storage    storage.Service
}

// NewServer creates a new HTTP server and setup routing
func NewServer(
	config config.Config,
	store repository.Store,
	storageService storage.Service,
) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	opt, err := redis.ParseURL(config.RedisAddress)
	if err != nil {
		// Fallback for simple address
		opt = &redis.Options{Addr: config.RedisAddress}
	}

	rdb := redis.NewClient(opt)
	hub := realtime.NewHub(rdb)
	go hub.Run() // Start the hub in a goroutine

	safetyMonitor := safety.NewMonitor(rdb)
	locationService := location.NewRedisLocationService(rdb, store)
	storyService := story.NewService(store, rdb, safetyMonitor)
	userService := user.NewService(store, tokenMaker, user.TokenConfig{
		AccessTokenDuration:  config.AccessTokenDuration,
		RefreshTokenDuration: config.RefreshTokenDuration,
	})
	adminService := admin.NewService(store, rdb)

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		redis:      rdb,
		safety:     safetyMonitor,
		hub:        hub,
		location:   locationService,
		story:      storyService,
		user:       userService,
		admin:      adminService,
		storage:    storageService,
	}

	server.setupRouter()
	return server, nil
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	// Force HTTP for localtunnel compatibility
	fmt.Printf("Starting HTTP server on %s\n", address)
	return server.router.Run(address)
}
