package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/aimodel/application"
	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	app    *application.AIModelService
	logger *slog.Logger
}

func NewHandler(app *application.AIModelService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) CreateModel(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Algorithm   string `json:"algorithm" binding:"required"`
		CreatorID   uint64 `json:"creator_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	model, err := h.app.CreateModel(c.Request.Context(), req.Name, req.Description, req.Type, req.Algorithm, req.CreatorID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create model", "name", req.Name, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Created", model)
}

func (h *Handler) ListModels(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page", "")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page_size", "")
		return
	}
	status := c.DefaultQuery("status", "")
	modelType := c.DefaultQuery("type", "")
	algorithm := c.DefaultQuery("algorithm", "")
	var creatorID uint64
	if val := c.Query("creator_id"); val != "" {
		creatorID, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid creator_id", err.Error())
			return
		}
	}

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
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) GetModelDetails(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid model ID", "")
		return
	}

	model, err := h.app.GetModelDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get model details", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, model)
}

func (h *Handler) StartTraining(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid model ID", "")
		return
	}

	if err := h.app.StartTraining(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to start training", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) Predict(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid model ID", "")
		return
	}

	var req struct {
		Input  string `json:"input" binding:"required"`
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	output, confidence, err := h.app.Predict(c.Request.Context(), id, req.Input, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to predict", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
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
