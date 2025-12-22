package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/file/application"
	"github.com/wyfcoding/ecommerce/internal/file/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了File模块的HTTP处理层。
type Handler struct {
	app    *application.FileService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 File HTTP Handler 实例。
func NewHandler(app *application.FileService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// UploadFile 处理文件上传的HTTP请求。
func (h *Handler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "No file uploaded", err.Error())
		return
	}

	fileType := domain.FileTypeOther

	metadata, err := h.app.UploadFile(c.Request.Context(), fileHeader.Filename, fileHeader.Size, fileType, nil)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to upload file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to upload file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "File uploaded successfully", metadata)
}

// GetFile 处理获取文件元数据信息的HTTP请求。
func (h *Handler) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	file, err := h.app.GetFile(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "File retrieved successfully", file)
}

// DeleteFile 处理删除文件的HTTP请求。
func (h *Handler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.DeleteFile(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete file", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "File deleted successfully", nil)
}

// ListFiles 处理获取文件列表的HTTP请求。
func (h *Handler) ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListFiles(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list files", "error", err)
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

// RegisterRoutes 在给定的Gin路由组中注册File模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/files")
	{
		group.POST("/upload", h.UploadFile)
		group.GET("/:id", h.GetFile)
		group.DELETE("/:id", h.DeleteFile)
		group.GET("", h.ListFiles)
	}
}
