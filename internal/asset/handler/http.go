package handler

import (
	"net/http"
	"strconv"

	"ecommerce/internal/asset/service"
	"github.com/gin-gonic/gin"
)

// AssetHandler handles HTTP requests for the asset service.
type AssetHandler struct {
	service *service.AssetService
}

// NewAssetHandler creates a new AssetHandler.
func NewAssetHandler(s *service.AssetService) *AssetHandler {
	return &AssetHandler{service: s}
}

// UploadAsset is a Gin handler for uploading an asset.
func (h *AssetHandler) UploadAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// You might want to get bucket name and uploadedBy from the request context (e.g., from a JWT token)
	bucketName := c.PostForm("bucket_name")
	if bucketName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bucket_name is required"})
		return
	}

	uploadedBy, err := strconv.ParseUint(c.PostForm("uploaded_by"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uploaded_by"})
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer fileContent.Close()

	uploadedFile, err := h.service.UploadFile(
		c.Request.Context(),
		file.Filename,
		file.Header.Get("Content-Type"),
		bucketName,
		file.Size,
		uploadedBy,
		fileContent,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"file_path": uploadedFile.FilePath})
}