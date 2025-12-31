package api

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
)

// conversationCacheKey generates a consistent cache key for a conversation between two users
func conversationCacheKey(userID1, userID2 uuid.UUID) string {
	ids := []string{userID1.String(), userID2.String()}
	sort.Strings(ids)
	return "messages:" + ids[0] + ":" + ids[1]
}

// invalidateConversationCache removes the cached conversation between two users
func (server *Server) invalidateConversationCache(userID1, userID2 uuid.UUID) {
	cacheKey := conversationCacheKey(userID1, userID2)
	server.redis.Del(context.Background(), cacheKey)
}

// invalidateProfileCache removes the cached profile for a user
func (server *Server) invalidateProfileCache(userID uuid.UUID) {
	cacheKey := "profile:" + userID.String()
	server.redis.Del(context.Background(), cacheKey)
}

// invalidateFeedCache removes the cached feed for a geohash
func (server *Server) invalidateFeedCache(geohash string) {
	cacheKey := "feed:" + geohash
	server.redis.Del(context.Background(), cacheKey)
}

// invalidateUnreadCountCache removes the cached unread count for a user
func (server *Server) invalidateUnreadCountCache(userID uuid.UUID) {
	unreadKey := "unread_count:" + userID.String()
	server.redis.Del(context.Background(), unreadKey)
}

// incrementUnreadCount increments the unread message count for a user
func (server *Server) incrementUnreadCount(userID uuid.UUID) {
	unreadKey := "unread_count:" + userID.String()
	server.redis.Incr(context.Background(), unreadKey)
}

// setCache stores data in Redis with the given key and TTL
func (server *Server) setCache(key string, data []byte, ttl time.Duration) {
	server.redis.Set(context.Background(), key, data, ttl)
}

// getCache retrieves data from Redis for the given key
func (server *Server) getCache(key string) (string, error) {
	return server.redis.Get(context.Background(), key).Result()
}

// invalidateCrossingsCache removes the cached crossings for a user
func (server *Server) invalidateCrossingsCache(userID uuid.UUID) {
	cacheKey := "crossings:v3:" + userID.String()
	server.redis.Del(context.Background(), cacheKey)
}
