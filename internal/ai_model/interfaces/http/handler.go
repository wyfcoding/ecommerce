package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/ai_model/application"
	"github.com/wyfcoding/ecommerce/internal/ai_model/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了AI模型模块的HTTP处理层。
type Handler struct {
	app    *application.AIModelService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 AIModel HTTP Handler 实例。
func NewHandler(app *application.AIModelService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateModel 处理创建AI模型的HTTP请求。
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

	model, err := h.app.CreateModel(c.Request.Context(), req.Name, req.Description, req.Type, req.Algorithm, req.CreatorID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create model", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create model", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Model created successfully", model)
}

// ListModels 处理获取AI模型列表的HTTP请求。
func (h *Handler) ListModels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.DefaultQuery("status", "")
	modelType := c.DefaultQuery("type", "")
	algorithm := c.DefaultQuery("algorithm", "")
	creatorID, _ := strconv.ParseUint(c.DefaultQuery("creator_id", "0"), 10, 64)

	query := &domain.ModelQuery{
		Status:    domain.ModelStatus(status),
		Type:      modelType,
		Algorithm: algorithm,
		CreatorID: creatorID,
		Page:      page,
		PageSize:  pageSize,
	}

	list, total, err := h.app.ListModels(c.Request.Context(), query)
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

// GetModelDetails 处理获取AI模型详情的HTTP请求。
func (h *Handler) GetModelDetails(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	model, err := h.app.GetModelDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get model details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get model details", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Model details retrieved successfully", model)
}

// StartTraining 处理启动AI模型训练的HTTP请求。
func (h *Handler) StartTraining(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.StartTraining(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to start training", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to start training", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Training started successfully", nil)
}

// Predict 处理AI模型预测的HTTP请求。
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

	output, confidence, err := h.app.Predict(c.Request.Context(), id, req.Input, req.UserID)
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

// RegisterRoutes 在给定的Gin路由组中注册AI模型模块的HTTP路由。
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
