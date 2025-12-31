package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"privacy-social-backend/internal/repository/db"
)

type shareStoryRequest struct {
	StoryID string   `json:"story_id" binding:"required,uuid"`
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
}

// shareStory shares a story to connections via DM
func (server *Server) shareStory(ctx *gin.Context) {
	var req shareStoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)
	storyID, ok := parseUUIDParam(ctx, req.StoryID, "story_id")
	if !ok {
		return
	}

	// Get story to create share message
	story, err := server.store.GetStoryByID(ctx, storyID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Create message with story link in content
	// Use relative path for internal deep linking in frontend
	shareText := fmt.Sprintf("📸 Shared a story with you: /view-story/%s", story.ID)
	successCount := 0

	for _, userIDStr := range req.UserIDs {
		targetUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			continue
		}

		// Check if users are connected before allowing share
		if err := server.checkConnection(ctx, authPayload.UserID, targetUserID); err != nil {
			continue // Skip non-connected users
		}

		// Create message with story link in content
		_, err = server.store.CreateMessage(ctx, db.CreateMessageParams{
			SenderID:   authPayload.UserID,
			ReceiverID: uuid.NullUUID{UUID: targetUserID, Valid: true},
			GroupID:    uuid.NullUUID{},
			Content:    shareText,
		})
		if err != nil {
			continue
		}

		successCount++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "story shared",
		"shared_to": successCount,
	})
}

// parseMentions extracts @username mentions from text
func parseMentions(text string) []string {
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	mentions := make([]string, 0)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			username := strings.ToLower(match[1])
			if !seen[username] {
				mentions = append(mentions, username)
				seen[username] = true
			}
		}
	}

	return mentions
}

// createStoryMentions creates mention records for a story
func (server *Server) createStoryMentions(ctx *gin.Context, storyID uuid.UUID, caption string) error {
	if caption == "" {
		return nil
	}

	mentions := parseMentions(caption)
	if len(mentions) == 0 {
		return nil
	}

	for _, username := range mentions {
		// Get user by username
		user, err := server.store.GetUserByUsername(ctx, username)
		if err != nil {
			continue // Skip if user not found
		}

		// Create mention
		_, err = server.store.CreateStoryMention(ctx, db.CreateStoryMentionParams{
			StoryID:         storyID,
			MentionedUserID: user.ID,
		})
		if err != nil {
			continue // Skip on error
		}

		// Send notification to mentioned user
		_, err = server.store.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:         user.ID,
			Type:           "story_mention",
			Title:          "You were mentioned!",
			Message:        "You were mentioned in a story",
			RelatedStoryID: uuid.NullUUID{UUID: storyID, Valid: true},
		})
		if err != nil {
			// Log error but don't fail the whole operation
			continue
		}
	}

	return nil
}
