package api

import (
	"fmt"
	"net/http"
	"privacy-social-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type uploadResponse struct {
	URL string `json:"url"`
}

func (server *Server) uploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("no file uploaded")))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to open file: %w", err)))
		return
	}
	defer file.Close()

	// Save locally to ./uploads
	filename := util.RandomString(32) + "_" + fileHeader.Filename
	dst := "./uploads/" + filename

	if err := ctx.SaveUploadedFile(fileHeader, dst); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to save file locally: %w", err)))
		return
	}

	// Construct public URL (Assuming backend serves ./uploads at /uploads)
	// We need the server address or just relative path?
	// The mobile app config API_URL usually ends with a slash or not?
	// Let's return a relative path that the frontend can append to API_URL or a full URL if we can construct it.
	// For simplicity and "make it work", let's return the relative path "/uploads/<filename>"
	publicURL := "/uploads/" + filename

	ctx.JSON(http.StatusOK, uploadResponse{
		URL: publicURL,
	})
}
