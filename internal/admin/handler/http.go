package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/service"
	// 伪代码: 模拟认证中间件
	// auth "ecommerce/internal/auth/handler"
)

// AdminHandler 负责处理管理后台的 HTTP 请求。
type AdminHandler struct {
	svc    service.AdminService // 业务逻辑服务接口
	logger *zap.Logger
}

// NewAdminHandler 创建一个新的 AdminHandler 实例。
func NewAdminHandler(svc service.AdminService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有管理后台相关的路由。
func (h *AdminHandler) RegisterRoutes(r *gin.Engine) {
	// 所有管理后台接口都需要管理员权限
	group := r.Group("/api/v1/admin")
	// group.Use(auth.AuthMiddleware(...), auth.AdminMiddleware(...)) // 认证和权限中间件
	{
		// 仪表盘统计
		group.GET("/dashboard/statistics", h.GetDashboardStatistics)

		// 用户管理
		group.GET("/users", h.ListUsers)
		group.GET("/users/:id", h.GetUserDetails)
		group.PUT("/users/:id/status", h.UpdateUserStatus)

		// 商品管理
		group.POST("/products", h.CreateProduct)
		group.GET("/products", h.ListProducts)
		group.PUT("/products/:id", h.UpdateProduct)
		group.DELETE("/products/:id", h.DeleteProduct)

		// 订单管理
		group.GET("/orders", h.ListOrders)
		group.GET("/orders/:id", h.GetOrderDetail)
		group.POST("/orders/:id/ship", h.ShipOrder)
		group.PUT("/orders/:id/status", h.UpdateOrderStatus)

		// 评论管理
		group.GET("/reviews", h.ListReviews)
		group.PUT("/reviews/:id/moderate", h.ModerateReview)

		// 优惠券管理
		group.POST("/coupons", h.CreateCoupon)
		group.GET("/coupons", h.ListCoupons)
		group.PUT("/coupons/:id", h.UpdateCoupon)
		group.DELETE("/coupons/:id", h.DeleteCoupon)

		// 管理员信息
		group.GET("/me", h.GetAdminUserInfo)
	}
}

// GetDashboardStatistics 处理获取仪表盘统计数据的请求。
// 它调用 AdminService 获取聚合统计数据并返回 JSON 响应。
func (h *AdminHandler) GetDashboardStatistics(c *gin.Context) {
	stats, err := h.svc.GetDashboardStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ListUsers 处理列出用户列表的请求。
// 它解析分页和过滤参数，调用 AdminService 获取用户列表并返回 JSON 响应。
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req v1.ListUsersRequest
	// 从查询参数绑定请求
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("ListUsers: invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.svc.ListUsers(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserDetails 处理获取用户详情的请求。
// 它从 URL 参数中获取用户ID，调用 AdminService 获取用户详情并返回 JSON 响应。
func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.logger.Warn("GetUserDetails: missing user ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	resp, err := h.svc.GetUserDetails(c.Request.Context(), &v1.GetUserDetailsRequest{UserId: userID})
	if err != nil {
		h.logger.Error("Failed to get user details", zap.Error(err), zap.String("user_id", userID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUserStatus 处理更新用户状态的请求。
// 它从 URL 参数中获取用户ID，从请求体中获取状态，调用 AdminService 更新用户状态并返回 JSON 响应。
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.logger.Warn("UpdateUserStatus: missing user ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	var reqBody struct {
		Status int32 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Warn("UpdateUserStatus: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.UpdateUserStatus(c.Request.Context(), &v1.UpdateUserStatusRequest{UserId: userID, Status: reqBody.Status})
	if err != nil {
		h.logger.Error("Failed to update user status", zap.Error(err), zap.String("user_id", userID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateProduct 处理创建商品的请求。
// 它从请求体中获取商品信息，调用 AdminService 创建商品并返回 JSON 响应。
func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req v1.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("CreateProduct: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListProducts 处理列出商品列表的请求。
// 它解析分页和过滤参数，调用 AdminService 获取商品列表并返回 JSON 响应。
func (h *AdminHandler) ListProducts(c *gin.Context) {
	var req v1.ListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("ListProducts: invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.svc.ListProducts(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateProduct 处理更新商品的请求。
// 它从 URL 参数中获取商品ID，从请求体中获取商品信息，调用 AdminService 更新商品并返回 JSON 响应。
func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.logger.Warn("UpdateProduct: missing product ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing product ID"})
		return
	}

	var req v1.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("UpdateProduct: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	req.ProductId = productID // 从路径参数设置 ProductId

	resp, err := h.svc.UpdateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Error(err), zap.String("product_id", productID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteProduct 处理删除商品的请求。
// 它从 URL 参数中获取商品ID，调用 AdminService 删除商品并返回 JSON 响应。
func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		h.logger.Warn("DeleteProduct: missing product ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing product ID"})
		return
	}

	resp, err := h.svc.DeleteProduct(c.Request.Context(), &v1.DeleteProductRequest{ProductId: productID})
	if err != nil {
		h.logger.Error("Failed to delete product", zap.Error(err), zap.String("product_id", productID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListOrders 处理列出订单列表的请求。
// 它解析分页和过滤参数，调用 AdminService 获取订单列表并返回 JSON 响应。
func (h *AdminHandler) ListOrders(c *gin.Context) {
	var req v1.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("ListOrders: invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.svc.ListOrders(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrderDetail 处理获取订单详情的请求。
// 它从 URL 参数中获取订单ID，调用 AdminService 获取订单详情并返回 JSON 响应。
func (h *AdminHandler) GetOrderDetail(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		h.logger.Warn("GetOrderDetail: missing order ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing order ID"})
		return
	}

	orderIDUint64, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		h.logger.Warn("GetOrderDetail: invalid order ID format", zap.Error(err), zap.String("order_id_str", orderID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	resp, err := h.svc.GetOrderDetail(c.Request.Context(), &v1.GetOrderDetailRequest{OrderId: orderIDUint64})
	if err != nil {
		h.logger.Error("Failed to get order detail", zap.Error(err), zap.String("order_id", orderID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order detail: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ShipOrder 处理发货订单的请求。
// 它从 URL 参数中获取订单ID，从请求体中获取物流信息，调用 AdminService 发货订单并返回 JSON 响应。
func (h *AdminHandler) ShipOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		h.logger.Warn("ShipOrder: missing order ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing order ID"})
		return
	}

	orderIDUint64, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		h.logger.Warn("ShipOrder: invalid order ID format", zap.Error(err), zap.String("order_id_str", orderID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var reqBody struct {
		TrackingCompany string `json:"tracking_company" binding:"required"`
		TrackingNumber  string `json:"tracking_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Warn("ShipOrder: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.ShipOrder(c.Request.Context(), &v1.ShipOrderRequest{OrderId: orderIDUint64, TrackingCompany: reqBody.TrackingCompany, TrackingNumber: reqBody.TrackingNumber})
	if err != nil {
		h.logger.Error("Failed to ship order", zap.Error(err), zap.String("order_id", orderID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ship order: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus 处理更新订单状态的请求。
// 它从 URL 参数中获取订单ID，从请求体中获取状态，调用 AdminService 更新订单状态并返回 JSON 响应。
func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		h.logger.Warn("UpdateOrderStatus: missing order ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing order ID"})
		return
	}

	orderIDUint64, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		h.logger.Warn("UpdateOrderStatus: invalid order ID format", zap.Error(err), zap.String("order_id_str", orderID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var reqBody struct {
		Status int32 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Warn("UpdateOrderStatus: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.UpdateOrderStatus(c.Request.Context(), &v1.UpdateOrderStatusRequest{OrderId: orderIDUint64, Status: reqBody.Status})
	if err != nil {
		h.logger.Error("Failed to update order status", zap.Error(err), zap.String("order_id", orderID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListReviews 处理列出评论列表的请求。
// 它解析分页和过滤参数，调用 AdminService 获取评论列表并返回 JSON 响应。
func (h *AdminHandler) ListReviews(c *gin.Context) {
	var req v1.ListReviewsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("ListReviews: invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.svc.ListReviews(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list reviews", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reviews: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ModerateReview 处理审核评论的请求。
// 它从 URL 参数中获取评论ID，从请求体中获取审核信息，调用 AdminService 审核评论并返回 JSON 响应。
func (h *AdminHandler) ModerateReview(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		h.logger.Warn("ModerateReview: missing review ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing review ID"})
		return
	}

	reviewIDUint64, err := strconv.ParseUint(reviewID, 10, 64)
	if err != nil {
		h.logger.Warn("ModerateReview: invalid review ID format", zap.Error(err), zap.String("review_id_str", reviewID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID format"})
		return
	}

	var reqBody struct {
		Status          int32  `json:"status" binding:"required"`
		ModerationNotes string `json:"moderation_notes"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		h.logger.Warn("ModerateReview: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.ModerateReview(c.Request.Context(), &v1.ModerateReviewRequest{ReviewId: reviewIDUint64, Status: reqBody.Status, ModerationNotes: reqBody.ModerationNotes})
	if err != nil {
		h.logger.Error("Failed to moderate review", zap.Error(err), zap.String("review_id", reviewID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to moderate review: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateCoupon 处理创建优惠券的请求。
// 它从请求体中获取优惠券信息，调用 AdminService 创建优惠券并返回 JSON 响应。
func (h *AdminHandler) CreateCoupon(c *gin.Context) {
	var req v1.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("CreateCoupon: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.svc.CreateCoupon(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create coupon", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListCoupons 处理列出优惠券列表的请求。
// 它解析分页和过滤参数，调用 AdminService 获取优惠券列表并返回 JSON 响应。
func (h *AdminHandler) ListCoupons(c *gin.Context) {
	var req v1.ListCouponsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("ListCoupons: invalid query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.svc.ListCoupons(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list coupons", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list coupons: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateCoupon 处理更新优惠券的请求。
// 它从 URL 参数中获取优惠券ID，从请求体中获取优惠券信息，调用 AdminService 更新优惠券并返回 JSON 响应。
func (h *AdminHandler) UpdateCoupon(c *gin.Context) {
	couponID := c.Param("id")
	if couponID == "" {
		h.logger.Warn("UpdateCoupon: missing coupon ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing coupon ID"})
		return
	}

	couponIDUint64, err := strconv.ParseUint(couponID, 10, 64)
	if err != nil {
		h.logger.Warn("UpdateCoupon: invalid coupon ID format", zap.Error(err), zap.String("coupon_id_str", couponID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon ID format"})
		return
	}

	var req v1.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("UpdateCoupon: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	req.CouponId = couponIDUint64 // 从路径参数设置 CouponId

	resp, err := h.svc.UpdateCoupon(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update coupon", zap.Error(err), zap.String("coupon_id", couponID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update coupon: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteCoupon 处理删除优惠券的请求。
// 它从 URL 参数中获取优惠券ID，调用 AdminService 删除优惠券并返回 JSON 响应。
func (h *AdminHandler) DeleteCoupon(c *gin.Context) {
	couponID := c.Param("id")
	if couponID == "" {
		h.logger.Warn("DeleteCoupon: missing coupon ID in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing coupon ID"})
		return
	}

	couponIDUint64, err := strconv.ParseUint(couponID, 10, 64)
	if err != nil {
		h.logger.Warn("DeleteCoupon: invalid coupon ID format", zap.Error(err), zap.String("coupon_id_str", couponID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon ID format"})
		return
	}

	resp, err := h.svc.DeleteCoupon(c.Request.Context(), &v1.DeleteCouponRequest{CouponId: couponIDUint64})
	if err != nil {
		h.logger.Error("Failed to delete coupon", zap.Error(err), zap.String("coupon_id", couponID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coupon: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAdminUserInfo 处理获取管理员自身信息的请求。
// 它调用 AdminService 获取管理员信息并返回 JSON 响应。
func (h *AdminHandler) GetAdminUserInfo(c *gin.Context) {
	// 从 JWT claims 中获取管理员用户ID，这里简化为从路径参数获取
	adminUserID := c.Param("id") // 假设 /api/v1/admin/me/:id
	if adminUserID == "" {
		// 如果没有从路径参数获取，可以尝试从 context 中获取 (JWT 拦截器设置的)
		// adminUserID = fmt.Sprintf("%d", c.MustGet("adminUserID").(uint64))
		// 这里为了简化，如果路径参数没有，就返回错误
		h.logger.Warn("GetAdminUserInfo: missing admin user ID in path or context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing admin user ID"})
		return
	}

	resp, err := h.svc.GetAdminUserInfo(c.Request.Context(), &v1.GetAdminUserInfoRequest{Id: adminUserID})
	if err != nil {
		h.logger.Error("Failed to get admin user info", zap.Error(err), zap.String("admin_user_id", adminUserID))
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin user info: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
