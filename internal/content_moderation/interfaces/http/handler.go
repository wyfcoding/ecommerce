package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/application"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.ModerationService
	logger  *slog.Logger
}

func NewHandler(service *application.ModerationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// SubmitContent 提交内容审核
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

	record, err := h.service.SubmitContent(c.Request.Context(), entity.ContentType(req.ContentType), req.ContentID, req.Content, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to submit content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to submit content", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Content submitted successfully", record)
}

// ReviewContent 人工审核
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

	err = h.service.ReviewContent(c.Request.Context(), id, req.ModeratorID, req.Approved, req.Reason)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to review content", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to review content", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Content reviewed successfully", nil)
}

// ListPendingRecords 获取待审核列表
func (h *Handler) ListPendingRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListPendingRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list pending records", "error", err)
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

// AddSensitiveWord 添加敏感词
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

	word, err := h.service.AddSensitiveWord(c.Request.Context(), req.Word, req.Category, req.Level)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add sensitive word", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Sensitive word added successfully", word)
}

// ListSensitiveWords 获取敏感词列表
func (h *Handler) ListSensitiveWords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.service.ListSensitiveWords(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list sensitive words", "error", err)
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

// DeleteSensitiveWord 删除敏感词
func (h *Handler) DeleteSensitiveWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	err = h.service.DeleteSensitiveWord(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete sensitive word", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete sensitive word", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Sensitive word deleted successfully", nil)
}

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
