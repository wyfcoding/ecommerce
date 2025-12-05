package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/wishlist/application" // 导入收藏夹模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                  // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Wishlist模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.WishlistService // 依赖Wishlist应用服务，处理核心业务逻辑。
	logger  *slog.Logger                 // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Wishlist HTTP Handler 实例。
func NewHandler(service *application.WishlistService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Add 处理将商品添加到收藏夹的HTTP请求。
// Method: POST
// Path: /wishlist
func (h *Handler) Add(c *gin.Context) {
	// 定义请求体结构，用于接收待添加到收藏夹的商品信息。
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`      // 用户ID，必填。
		ProductID   uint64 `json:"product_id" binding:"required"`   // 商品ID，必填。
		SkuID       uint64 `json:"sku_id" binding:"required"`       // SKU ID，必填。
		ProductName string `json:"product_name" binding:"required"` // 商品名称，必填。
		SkuName     string `json:"sku_name" binding:"required"`     // SKU名称，必填。
		Price       uint64 `json:"price" binding:"required"`        // 价格，必填。
		ImageURL    string `json:"image_url"`                       // 图片URL，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加商品到收藏夹。
	wishlist, err := h.service.Add(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.ImageURL)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add to wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to wishlist", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Added to wishlist successfully", wishlist)
}

// Remove 处理从收藏夹移除商品的HTTP请求。
// Method: DELETE
// Path: /wishlist/:id
func (h *Handler) Remove(c *gin.Context) {
	// 从URL路径中解析收藏夹条目ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 备注：用户ID通常应从认证token或上下文获取，而不是从查询参数。
	// 这里为简化，从查询参数中获取user_id。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	// 调用应用服务层从收藏夹移除商品。
	if err := h.service.Remove(c.Request.Context(), userID, id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to remove from wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from wishlist", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Removed from wishlist successfully", nil)
}

// List 处理获取用户收藏夹列表的HTTP请求。
// Method: GET
// Path: /wishlist
func (h *Handler) List(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取收藏夹列表。
	list, total, err := h.service.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list wishlist", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Wishlist listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CheckStatus 处理检查商品是否在用户收藏夹中的HTTP请求。
// Method: GET
// Path: /wishlist/status
func (h *Handler) CheckStatus(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}
	// 从查询参数中获取SKU ID。
	skuID, err := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "SKU ID required", err.Error())
		return
	}

	// 调用应用服务层检查收藏状态。
	exists, err := h.service.CheckStatus(c.Request.Context(), userID, skuID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to check status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to check status", err.Error())
		return
	}

	// 返回成功的响应，包含存在状态。
	response.SuccessWithStatus(c, http.StatusOK, "Status checked successfully", gin.H{"exists": exists})
}

// RegisterRoutes 在给定的Gin路由组中注册Wishlist模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /wishlist 路由组，用于所有收藏夹相关接口。
	group := r.Group("/wishlist")
	{
		group.POST("", h.Add)                               // 添加商品到收藏夹。
		group.DELETE("/:id", h.Remove)                      // 从收藏夹移除商品。
		group.GET("", h.List)                               // 获取收藏夹列表。
		group.GET("/status", h.CheckStatus)                 // 检查商品收藏状态。
		group.DELETE("", h.Clear)                           // 清空收藏夹。
		group.DELETE("/product/:sku_id", h.RemoveByProduct) // 按商品移除。
	}
}

// Clear 处理清空收藏夹的HTTP请求。
// Method: DELETE
// Path: /wishlist
func (h *Handler) Clear(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	// 调用应用服务层清空收藏夹。
	if err := h.service.Clear(c.Request.Context(), userID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to clear wishlist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear wishlist", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Wishlist cleared successfully", nil)
}

// RemoveByProduct 处理按商品移除收藏夹条目的HTTP请求。
// Method: DELETE
// Path: /wishlist/product/:sku_id
func (h *Handler) RemoveByProduct(c *gin.Context) {
	// 从URL路径中解析SKU ID。
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "User ID required", err.Error())
		return
	}

	// 调用应用服务层按商品移除。
	if err := h.service.RemoveByProduct(c.Request.Context(), userID, skuID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to remove from wishlist by product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from wishlist by product", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Removed from wishlist by product successfully", nil)
}
