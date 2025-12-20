package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了ContentModeration模块的HTTP处理层。
type Handler struct {
	app    *application.ModerationService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 ContentModeration HTTP Handler 实例。
func NewHandler(app *application.ModerationService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// SubmitContent 处理提交内容审核的HTTP请求。
func (h *Handler) SubmitContent(c *gin.Context) {
	var req struct {
		ContentType string `json:"content_type" binding:"required"`
		ContentID   uint64 `json:"content_id" binding:"required"`
		Content     string `json:"content" binding:"required"`
		UserID      uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	record, err := h.app.SubmitContent(c.Request.Context(), domain.ContentType(req.ContentType), req.ContentID, req.Content, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to submit content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to submit content", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Content submitted successfully", record)
}

// ReviewContent 处理人工审核内容的HTTP请求。
func (h *Handler) ReviewContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		ModeratorID uint64 `json:"moderator_id" binding:"required"`
		Approved    bool   `json:"approved"`
		Reason      string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	err = h.app.ReviewContent(c.Request.Context(), id, req.ModeratorID, req.Approved, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to review content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to review content", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Content reviewed successfully", nil)
}

// ListPendingRecords 处理获取待审核内容列表的HTTP请求。
func (h *Handler) ListPendingRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListPendingRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list pending records", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list pending records", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Pending records listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddSensitiveWord 处理添加敏感词的HTTP请求。
func (h *Handler) AddSensitiveWord(c *gin.Context) {
	var req struct {
		Word     string `json:"word" binding:"required"`
		Category string `json:"category" binding:"required"`
		Level    int8   `json:"level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	word, err := h.app.AddSensitiveWord(c.Request.Context(), req.Word, req.Category, req.Level)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to add sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add sensitive word", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Sensitive word added successfully", word)
}

// ListSensitiveWords 处理获取敏感词列表的HTTP请求。
func (h *Handler) ListSensitiveWords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.app.ListSensitiveWords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list sensitive words", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list sensitive words", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Sensitive words listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteSensitiveWord 处理删除敏感词的HTTP请求。
func (h *Handler) DeleteSensitiveWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.app.DeleteSensitiveWord(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete sensitive word", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Sensitive word deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册ContentModeration模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/moderation")
	{
		group.POST("/submit", h.SubmitContent)
		group.POST("/review/:id", h.ReviewContent)
		group.GET("/pending", h.ListPendingRecords)
		group.POST("/sensitive-words", h.AddSensitiveWord)
		group.GET("/sensitive-words", h.ListSensitiveWords)
		group.DELETE("/sensitive-words/:id", h.DeleteSensitiveWord)
	}
}
