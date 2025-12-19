package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/recommendation/application" // 导入推荐模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"      // 导入推荐模块的领域层。
	"github.com/wyfcoding/pkg/response"                                  // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Recommendation模块的HTTP处理层。
type Handler struct {
	app    *application.RecommendationService // 依赖Recommendation应用服务 (Facade)。
	logger *slog.Logger                       // 日志记录器。
}

// NewHandler 创建并返回一个新的 Recommendation HTTP Handler 实例。
func NewHandler(app *application.RecommendationService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// GetRecommendations 处理获取推荐列表的HTTP请求。
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	recType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	recs, err := h.app.GetRecommendations(c.Request.Context(), userID, recType, limit)
	if err != nil {
		h.logger.Error("Failed to get recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get recommendations", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Recommendations retrieved successfully", recs)
}

// TrackBehavior 处理记录用户行为的HTTP请求。
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

	if err := h.app.TrackBehavior(c.Request.Context(), req.UserID, req.ProductID, req.Action); err != nil {
		h.logger.Error("Failed to track behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to track behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Behavior tracked successfully", nil)
}

// UpdatePreference 处理更新用户偏好设置的HTTP请求。
func (h *Handler) UpdatePreference(c *gin.Context) {
	var req domain.UserPreference
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.UpdateUserPreference(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to update preference", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update preference", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Preference updated successfully", nil)
}

// GetSimilarProducts 处理获取相似商品列表的HTTP请求。
func (h *Handler) GetSimilarProducts(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Product ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sims, err := h.app.GetSimilarProducts(c.Request.Context(), productID, limit)
	if err != nil {
		h.logger.Error("Failed to get similar products", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get similar products", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Similar products retrieved successfully", sims)
}

// GenerateRecommendations 处理触发生成推荐的HTTP请求。
func (h *Handler) GenerateRecommendations(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.GenerateRecommendations(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to generate recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate recommendations", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Recommendations generated successfully", nil)
}

// RegisterRoutes 注册推荐模块的HTTP路由。
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
