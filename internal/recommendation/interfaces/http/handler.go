package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/recommendation/application"   // 导入推荐模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/recommendation/domain/entity" // 导入推荐模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                          // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Recommendation模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.RecommendationService // 依赖Recommendation应用服务，处理核心业务逻辑。
	logger  *slog.Logger                       // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Recommendation HTTP Handler 实例。
func NewHandler(service *application.RecommendationService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetRecommendations 处理获取推荐列表的HTTP请求。
// Method: GET
// Path: /recommendation/list
func (h *Handler) GetRecommendations(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取推荐类型和数量限制，并设置默认值。
	recType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用应用服务层获取推荐列表。
	recs, err := h.service.GetRecommendations(c.Request.Context(), userID, recType, limit)
	if err != nil {
		h.logger.Error("Failed to get recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get recommendations", err.Error())
		return
	}

	// 返回成功的响应，包含推荐列表。
	response.SuccessWithStatus(c, http.StatusOK, "Recommendations retrieved successfully", recs)
}

// TrackBehavior 处理记录用户行为的HTTP请求。
// Method: POST
// Path: /recommendation/track
func (h *Handler) TrackBehavior(c *gin.Context) {
	// 定义请求体结构，用于接收用户行为数据。
	var req struct {
		UserID    uint64 `json:"user_id" binding:"required"`    // 用户ID，必填。
		ProductID uint64 `json:"product_id" binding:"required"` // 商品ID，必填。
		Action    string `json:"action" binding:"required"`     // 行为类型，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层记录用户行为。
	if err := h.service.TrackBehavior(c.Request.Context(), req.UserID, req.ProductID, req.Action); err != nil {
		h.logger.Error("Failed to track behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to track behavior", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Behavior tracked successfully", nil)
}

// UpdatePreference 处理更新用户偏好设置的HTTP请求。
// Method: POST
// Path: /recommendation/preference
func (h *Handler) UpdatePreference(c *gin.Context) {
	// 定义请求体结构，使用 entity.UserPreference 结构体直接绑定。
	var req entity.UserPreference
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层更新用户偏好。
	if err := h.service.UpdateUserPreference(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to update preference", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update preference", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Preference updated successfully", nil)
}

// GetSimilarProducts 处理获取相似商品列表的HTTP请求。
// Method: GET
// Path: /recommendation/similar
func (h *Handler) GetSimilarProducts(c *gin.Context) {
	// 从查询参数中获取商品ID。
	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Product ID", err.Error())
		return
	}

	// 从查询参数中获取数量限制，并设置默认值。
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用应用服务层获取相似商品列表。
	sims, err := h.service.GetSimilarProducts(c.Request.Context(), productID, limit)
	if err != nil {
		h.logger.Error("Failed to get similar products", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get similar products", err.Error())
		return
	}

	// 返回成功的响应，包含相似商品列表。
	response.SuccessWithStatus(c, http.StatusOK, "Similar products retrieved successfully", sims)
}

// GenerateRecommendations 处理触发生成推荐的HTTP请求。
// Method: POST
// Path: /recommendation/generate
func (h *Handler) GenerateRecommendations(c *gin.Context) {
	// 定义请求体结构，用于接收用户ID。
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层生成推荐。
	if err := h.service.GenerateRecommendations(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to generate recommendations", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to generate recommendations", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Recommendations generated successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Recommendation模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /recommendation 路由组，用于所有推荐相关接口。
	group := r.Group("/recommendation")
	{
		group.GET("/list", h.GetRecommendations)           // 获取推荐列表。
		group.POST("/track", h.TrackBehavior)              // 记录用户行为。
		group.POST("/preference", h.UpdatePreference)      // 更新用户偏好。
		group.GET("/similar", h.GetSimilarProducts)        // 获取相似商品。
		group.POST("/generate", h.GenerateRecommendations) // 触发生成推荐。
		// TODO: 补充获取用户偏好、用户行为列表等接口。
	}
}
