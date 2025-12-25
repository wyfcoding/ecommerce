package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/cart/application" // 导入购物车模块的应用服务。
	"github.com/wyfcoding/pkg/response"                        // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Cart模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	app    *application.Cart // 依赖Cart应用服务，处理核心业务逻辑。
	logger *slog.Logger             // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Cart HTTP Handler 实例。
func NewHandler(app *application.Cart, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// GetCart 处理获取用户购物车信息的HTTP请求。
// HTTP 方法: GET
// 请求路径: /cart
func (h *Handler) GetCart(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if userID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid user ID", "user_id is required")
		return
	}

	// 调用应用服务层获取购物车。
	cart, err := h.app.GetCart(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get cart", err.Error())
		return
	}

	// 返回成功的响应，包含购物车信息。
	response.SuccessWithStatus(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddItem 处理添加商品到购物车的HTTP请求。
// HTTP 方法: POST
// 请求路径: /cart/items
func (h *Handler) AddItem(c *gin.Context) {
	// 定义请求体结构，用于接收要添加到购物车的商品信息。
	var req struct {
		UserID          uint64  `json:"user_id" binding:"required"`      // 用户ID，必填。
		ProductID       uint64  `json:"product_id" binding:"required"`   // 商品ID，必填。
		SkuID           uint64  `json:"sku_id" binding:"required"`       // SKU ID，必填。
		ProductName     string  `json:"product_name" binding:"required"` // 商品名称，必填。
		SkuName         string  `json:"sku_name" binding:"required"`     // SKU名称，必填。
		Price           float64 `json:"price" binding:"required"`        // 商品单价，必填。
		Quantity        int32   `json:"quantity" binding:"required"`     // 商品数量，必填。
		ProductImageURL string  `json:"product_image_url"`               // 商品图片URL，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加商品到购物车。
	err := h.app.AddItem(c.Request.Context(), req.UserID, req.ProductID, req.SkuID, req.ProductName, req.SkuName, req.Price, req.Quantity, req.ProductImageURL)
	if err != nil {
		h.logger.Error("Failed to add item to cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add item to cart", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Item added to cart successfully", nil)
}

// UpdateItemQuantity 处理更新购物车中商品数量的HTTP请求。
// HTTP 方法: PUT
// 请求路径: /cart/items
func (h *Handler) UpdateItemQuantity(c *gin.Context) {
	// 定义请求体结构，用于接收要更新的商品信息。
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`  // 用户ID，必填。
		SkuID    uint64 `json:"sku_id" binding:"required"`   // SKU ID，必填。
		Quantity int32  `json:"quantity" binding:"required"` // 更新后的数量，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层更新购物车商品数量。
	err := h.app.UpdateItemQuantity(c.Request.Context(), req.UserID, req.SkuID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to update item quantity", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update item quantity", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Item quantity updated successfully", nil)
}

// RemoveItem 处理从购物车中移除商品的HTTP请求。
// HTTP 方法: DELETE
// 请求路径: /cart/items
func (h *Handler) RemoveItem(c *gin.Context) {
	// 定义请求体结构，用于接收要移除的商品信息。
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
		SkuID  uint64 `json:"sku_id" binding:"required"`  // SKU ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层移除购物车商品。
	err := h.app.RemoveItem(c.Request.Context(), req.UserID, req.SkuID)
	if err != nil {
		h.logger.Error("Failed to remove item from cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove item from cart", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Item removed from cart successfully", nil)
}

// ClearCart 处理清空购物车所有商品的HTTP请求。
// HTTP 方法: DELETE
// 请求路径: /cart
func (h *Handler) ClearCart(c *gin.Context) {
	// 定义请求体结构，用于接收用户ID。
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层清空购物车。
	err := h.app.ClearCart(c.Request.Context(), req.UserID)
	if err != nil {
		h.logger.Error("Failed to clear cart", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Cart cleared successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Cart模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /cart 路由组，用于所有购物车相关接口。
	group := r.Group("/cart")
	{
		group.GET("", h.GetCart)                  // 获取购物车。
		group.POST("/items", h.AddItem)           // 添加商品到购物车。
		group.PUT("/items", h.UpdateItemQuantity) // 更新购物车商品数量。
		group.DELETE("/items", h.RemoveItem)      // 移除购物车商品。
		group.DELETE("", h.ClearCart)             // 清空购物车。
	}
}
