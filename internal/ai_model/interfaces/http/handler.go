package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/ai_model/application"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.AIModelService
	logger  *slog.Logger
}

func NewHandler(service *application.AIModelService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateModel 创建模型
func (h *Handler) CreateModel(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Algorithm   string `json:"algorithm" binding:"required"`
		CreatorID   uint64 `json:"creator_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	model, err := h.service.CreateModel(c.Request.Context(), req.Name, req.Description, req.Type, req.Algorithm, req.CreatorID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create model", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create model", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Model created successfully", model)
}

// ListModels 获取模型列表
func (h *Handler) ListModels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.DefaultQuery("status", "")
	modelType := c.DefaultQuery("type", "")
	algorithm := c.DefaultQuery("algorithm", "")
	creatorID, _ := strconv.ParseUint(c.DefaultQuery("creator_id", "0"), 10, 64)

	query := &repository.ModelQuery{
		Status:    entity.ModelStatus(status),
		Type:      modelType,
		Algorithm: algorithm,
		CreatorID: creatorID,
		Page:      page,
		PageSize:  pageSize,
	}

	list, total, err := h.service.ListModels(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list models", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list models", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Models listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetModelDetails 获取模型详情
func (h *Handler) GetModelDetails(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	model, err := h.service.GetModelDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get model details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get model details", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Model details retrieved successfully", model)
}

// StartTraining 开始训练
func (h *Handler) StartTraining(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.StartTraining(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to start training", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to start training", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Training started successfully", nil)
}

// Predict 预测
func (h *Handler) Predict(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Input  string `json:"input" binding:"required"`
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	output, confidence, err := h.service.Predict(c.Request.Context(), id, req.Input, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to predict", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to predict", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Prediction successful", gin.H{
		"output":     output,
		"confidence": confidence,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/ai-models")
	{
		group.POST("", h.CreateModel)
		group.GET("", h.ListModels)
		group.GET("/:id", h.GetModelDetails)
		group.POST("/:id/train", h.StartTraining)
		group.POST("/:id/predict", h.Predict)
	}
}
