package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/file/application"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.FileService
	logger  *slog.Logger
}

func NewHandler(service *application.FileService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// UploadFile 上传文件
func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "No file uploaded", err.Error())
		return
	}

	// For simulation, we just read the size and name
	// In production, we would read the content and stream it to storage

	fileType := entity.FileTypeOther // Simplified type detection

	metadata, err := h.service.UploadFile(c.Request.Context(), file.Filename, file.Size, fileType, nil)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to upload file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "File uploaded successfully", metadata)
}

// GetFile 获取文件信息
func (h *Handler) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	file, err := h.service.GetFile(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "File retrieved successfully", file)
}

// DeleteFile 删除文件
func (h *Handler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteFile(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "File deleted successfully", nil)
}

// ListFiles 获取文件列表
func (h *Handler) ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListFiles(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list files", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list files", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Files listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/files")
	{
		group.POST("/upload", h.UploadFile)
		group.GET("/:id", h.GetFile)
		group.DELETE("/:id", h.DeleteFile)
		group.GET("", h.ListFiles)
	}
}
