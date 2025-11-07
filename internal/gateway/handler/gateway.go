package handler

import (
	"net/http"
	"time"

	v1_order "ecommerce/api/order/v1"
	v1_product "ecommerce/api/product/v1"
	v1_user "ecommerce/api/user/v1"
	"ecommerce/internal/gateway/proxy"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GatewayHandler 网关处理器
type GatewayHandler struct {
	grpcProxy *proxy.GRPCProxy
}

// NewGatewayHandler 创建网关处理器
func NewGatewayHandler(grpcProxy *proxy.GRPCProxy) *GatewayHandler {
	return &GatewayHandler{
		grpcProxy: grpcProxy,
	}
}

// RegisterRoutes 注册路由
func (h *GatewayHandler) RegisterRoutes(r *gin.Engine, authMiddleware, optionalAuthMiddleware gin.HandlerFunc) {
	// 健康检查
	r.GET("/health", h.Health)
	r.GET("/health/services", h.ServicesHealth)

	// API 版本组
	v1 := r.Group("/api/v1")
	{
		// 用户相关路由 (公开)
		userPublic := v1.Group("/users")
		{
			userPublic.POST("/register", h.Register)
			userPublic.POST("/login", h.Login)
		}

		// 用户相关路由 (需要认证)
		userAuth := v1.Group("/users", authMiddleware)
		{
			userAuth.GET("/profile", h.GetUserProfile)
			userAuth.PUT("/profile", h.UpdateUserProfile)
			userAuth.GET("/addresses", h.ListAddresses)
			userAuth.POST("/addresses", h.AddAddress)
			userAuth.PUT("/addresses/:id", h.UpdateAddress)
			userAuth.DELETE("/addresses/:id", h.DeleteAddress)
		}

		// 商品相关路由 (公开)
		products := v1.Group("/products", optionalAuthMiddleware)
		{
			products.GET("", h.ListProducts)
			products.GET("/:id", h.GetProduct)
			products.GET("/categories", h.ListCategories)
			products.GET("/brands", h.ListBrands)
		}

		// 订单相关路由 (需要认证)
		orders := v1.Group("/orders", authMiddleware)
		{
			orders.POST("", h.CreateOrder)
			orders.GET("", h.ListOrders)
			orders.GET("/:id", h.GetOrder)
			orders.POST("/:id/cancel", h.CancelOrder)
			orders.POST("/:id/pay", h.PayOrder)
		}

		// 购物车相关路由 (需要认证)
		cart := v1.Group("/cart", authMiddleware)
		{
			cart.GET("", h.GetCart)
			cart.POST("/items", h.AddCartItem)
			cart.PUT("/items/:id", h.UpdateCartItem)
			cart.DELETE("/items/:id", h.RemoveCartItem)
			cart.DELETE("", h.ClearCart)
		}

		// 搜索相关路由 (公开)
		search := v1.Group("/search", optionalAuthMiddleware)
		{
			search.GET("/products", h.SearchProducts)
			search.GET("/suggestions", h.SearchSuggestions)
		}

		// 推荐相关路由 (可选认证)
		recommend := v1.Group("/recommendations", optionalAuthMiddleware)
		{
			recommend.GET("/products", h.GetRecommendations)
			recommend.GET("/similar/:id", h.GetSimilarProducts)
		}
	}
}

// --- 健康检查 ---

func (h *GatewayHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

func (h *GatewayHandler) ServicesHealth(c *gin.Context) {
	health := h.grpcProxy.HealthCheck()
	allHealthy := true
	for _, healthy := range health {
		if !healthy {
			allHealthy = false
			break
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"services": health,
		"healthy":  allHealthy,
	})
}

// --- 用户服务代理 ---

func (h *GatewayHandler) Register(c *gin.Context) {
	var req v1_user.RegisterByPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	var resp v1_user.RegisterResponse
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/RegisterByPassword", &req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "registration failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) Login(c *gin.Context) {
	var req v1_user.LoginByPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	var resp v1_user.LoginByPasswordResponse
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/LoginByPassword", &req, &resp); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "login failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) GetUserProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	req := &v1_user.GetUserByIDRequest{
		UserId: userID.(uint64),
	}

	var resp v1_user.UserResponse
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/GetUserByID", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp.User,
	})
}

func (h *GatewayHandler) UpdateUserProfile(c *gin.Context) {
	var req v1_user.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	var resp v1_user.UserResponse
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/UpdateUserInfo", &req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp.User,
	})
}

func (h *GatewayHandler) ListAddresses(c *gin.Context) {
	userID, _ := c.Get("user_id")

	req := &v1_user.ListAddressesRequest{
		UserId: userID.(uint64),
	}

	var resp v1_user.ListAddressesResponse
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/ListAddresses", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp.Addresses,
	})
}

func (h *GatewayHandler) AddAddress(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req v1_user.AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	req.UserId = userID.(uint64)

	var resp v1_user.Address
	if err := h.grpcProxy.ProxyRequest(c, "user", "/user.v1.UserService/AddAddress", &req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to add address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) UpdateAddress(c *gin.Context) {
	// 实现类似...
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) DeleteAddress(c *gin.Context) {
	// 实现类似...
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

// --- 商品服务代理 ---

func (h *GatewayHandler) ListProducts(c *gin.Context) {
	req := &v1_product.ListProductsRequest{
		Page:     c.GetInt32("page"),
		PageSize: c.GetInt32("page_size"),
	}

	var resp v1_product.ListProductsResponse
	if err := h.grpcProxy.ProxyRequest(c, "product", "/product.v1.ProductService/ListProducts", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) GetProduct(c *gin.Context) {
	id := c.GetUint64("id")

	req := &v1_product.GetProductByIDRequest{
		Id: id,
	}

	var resp v1_product.ProductInfo
	if err := h.grpcProxy.ProxyRequest(c, "product", "/product.v1.ProductService/GetProductByID", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) ListCategories(c *gin.Context) {
	req := &v1_product.ListCategoriesRequest{
		ParentId: c.GetUint64("parent_id"),
	}

	var resp v1_product.ListCategoriesResponse
	if err := h.grpcProxy.ProxyRequest(c, "product", "/product.v1.ProductService/ListCategories", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp.Categories,
	})
}

func (h *GatewayHandler) ListBrands(c *gin.Context) {
	req := &v1_product.ListBrandsRequest{
		Page:     c.GetInt32("page"),
		PageSize: c.GetInt32("page_size"),
	}

	var resp v1_product.ListBrandsResponse
	if err := h.grpcProxy.ProxyRequest(c, "product", "/product.v1.ProductService/ListBrands", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list brands"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// --- 订单服务代理 ---

func (h *GatewayHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req v1_order.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	req.UserId = userID.(uint64)

	var resp v1_order.OrderInfo
	if err := h.grpcProxy.ProxyRequest(c, "order", "/order.v1.OrderService/CreateOrder", &req, &resp); err != nil {
		zap.S().Errorf("failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to create order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) ListOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	req := &v1_order.ListOrdersRequest{
		UserId:   userID.(uint64),
		Page:     c.GetInt32("page"),
		PageSize: c.GetInt32("page_size"),
	}

	var resp v1_order.ListOrdersResponse
	if err := h.grpcProxy.ProxyRequest(c, "order", "/order.v1.OrderService/ListOrders", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) GetOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.GetUint64("id")

	req := &v1_order.GetOrderByIDRequest{
		Id:     orderID,
		UserId: userID.(uint64),
	}

	var resp v1_order.OrderInfo
	if err := h.grpcProxy.ProxyRequest(c, "order", "/order.v1.OrderService/GetOrderByID", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) CancelOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.GetUint64("id")

	var reqBody struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	req := &v1_order.CancelOrderRequest{
		Id:     orderID,
		UserId: userID.(uint64),
		Reason: reqBody.Reason,
	}

	var resp v1_order.OrderInfo
	if err := h.grpcProxy.ProxyRequest(c, "order", "/order.v1.OrderService/CancelOrder", req, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

func (h *GatewayHandler) PayOrder(c *gin.Context) {
	// 实现支付逻辑...
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

// --- 购物车服务代理 ---

func (h *GatewayHandler) GetCart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) AddCartItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) UpdateCartItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) RemoveCartItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) ClearCart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

// --- 搜索服务代理 ---

func (h *GatewayHandler) SearchProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) SearchSuggestions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

// --- 推荐服务代理 ---

func (h *GatewayHandler) GetRecommendations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}

func (h *GatewayHandler) GetSimilarProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "not implemented"})
}
