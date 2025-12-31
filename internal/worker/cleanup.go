package worker

import (
	"context"
	"time"

	"privacy-social-backend/internal/repository"

	"github.com/rs/zerolog/log"
)

type CleanupWorker struct {
	store repository.Store
}

func NewCleanupWorker(store repository.Store) *CleanupWorker {
	return &CleanupWorker{
		store: store,
	}
}

func (worker *CleanupWorker) Start() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			<-ticker.C
			log.Info().Msg("Running cleanup worker...")
			worker.cleanup()
		}
	}()
}

func (worker *CleanupWorker) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err := worker.store.DeleteExpiredLocations(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete expired locations")
	} else {
		log.Info().Msg("Expired locations deleted")
	}

	// Cleanup expired stories
	err = worker.store.DeleteExpiredStories(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete expired stories")
	} else {
		log.Info().Msg("Expired stories deleted")
	}

	// Cleanup old messages (30+ days)
	err = worker.store.DeleteOldMessages(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete old messages")
	} else {
		log.Info().Msg("Old messages deleted")
	}

	// Cleanup expired messages (Secret Mode disappearing messages)
	err = worker.store.DeleteExpiredMessages(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete expired messages")
	} else {
		log.Info().Msg("Expired messages deleted")
	}

	// Cleanup old notifications (30+ days)
	err = worker.store.DeleteOldNotifications(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete old notifications")
	} else {
		log.Info().Msg("Old notifications deleted")
	}
}
