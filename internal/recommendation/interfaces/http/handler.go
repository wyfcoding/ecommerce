package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/recommendation/application"
	"ecommerce/internal/recommendation/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.RecommendationService
	logger  *slog.Logger
}

func NewHandler(service *application.RecommendationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetRecommendations 获取推荐
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	recType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	recs, err := h.service.GetRecommendations(c.Request.Context(), userID, recType, limit)
	if err != nil {
		h.logger.Error("Failed to get recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get recommendations", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Recommendations retrieved successfully", recs)
}

// TrackBehavior 记录行为
func (h *Handler) TrackBehavior(c *gin.Context) {
	var req struct {
		UserID    uint64 `json:"user_id" binding:"required"`
		ProductID uint64 `json:"product_id" binding:"required"`
		Action    string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.TrackBehavior(c.Request.Context(), req.UserID, req.ProductID, req.Action); err != nil {
		h.logger.Error("Failed to track behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to track behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Behavior tracked successfully", nil)
}

// UpdatePreference 更新偏好
func (h *Handler) UpdatePreference(c *gin.Context) {
	var req entity.UserPreference
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.UpdateUserPreference(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to update preference", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update preference", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Preference updated successfully", nil)
}

// GetSimilarProducts 获取相似商品
func (h *Handler) GetSimilarProducts(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Product ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sims, err := h.service.GetSimilarProducts(c.Request.Context(), productID, limit)
	if err != nil {
		h.logger.Error("Failed to get similar products", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get similar products", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Similar products retrieved successfully", sims)
}

// GenerateRecommendations 触发生成推荐
func (h *Handler) GenerateRecommendations(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.GenerateRecommendations(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to generate recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate recommendations", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Recommendations generated successfully", nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/recommendation")
	{
		group.GET("/list", h.GetRecommendations)
		group.POST("/track", h.TrackBehavior)
		group.POST("/preference", h.UpdatePreference)
		group.GET("/similar", h.GetSimilarProducts)
		group.POST("/generate", h.GenerateRecommendations)
	}
}
