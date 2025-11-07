package handler

import (
	"context"
	"strconv"

	authv1 "ecommerce/api/auth/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	userv1 "ecommerce/api/user/v1"
	"ecommerce/internal/gateway/client"

	"github.com/gin-gonic/gin"
)

// Handler holds the gRPC clients.
type Handler struct {
	clients *client.Clients
}

// NewHandler creates a new Handler.
func NewHandler(clients *client.Clients) *Handler {
	return &Handler{clients: clients}
}

// RegisterRoutes registers all the routes for the gateway.
func (h *Handler) RegisterRoutes(e *gin.Engine) {
	v1 := e.Group("/v1")
	{
		// Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", h.Login)
			// Add other auth routes here...
		}

		// Product routes
		productGroup := v1.Group("/products")
		{
			productGroup.GET("/:id", h.GetProductDetail)
			// Add other product routes here...
		}

		// Order routes
		orderGroup := v1.Group("/orders")
		{
			orderGroup.POST("", h.CreateOrder)
			orderGroup.GET("/:id", h.GetOrderDetail)
			// Add other order routes here...
		}

		// User routes
		userGroup := v1.Group("/users")
		{
			userGroup.POST("/register", h.RegisterUser)
			userGroup.GET("/:id", h.GetUserDetails)
			// Add other user routes here...
		}
	}
}

// Login handles user login.
func (h *Handler) Login(c *gin.Context) {
	var req authv1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.clients.AuthClient.Login(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Login failed"})
		return
	}

	c.JSON(200, resp)
}

// GetProductDetail handles the HTTP request to get a product's detail.
func (h *Handler) GetProductDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return
	}

	req := &productv1.GetSpuDetailRequest{SpuId: id}

	resp, err := h.clients.ProductClient.GetSpuDetail(context.Background(), req)
	if err != nil {
		// In a real app, you'd check the gRPC error code and return a more specific HTTP status.
		c.JSON(500, gin.H{"error": "Failed to get product detail"})
		return
	}

	c.JSON(200, resp)
}

// CreateOrder handles the HTTP request to create an order.
func (h *Handler) CreateOrder(c *gin.Context) {
	var req orderv1.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.clients.OrderClient.CreateOrder(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(200, resp)
}

// GetOrderDetail handles the HTTP request to get an order's detail.
func (h *Handler) GetOrderDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid order ID"})
		return
	}

	req := &orderv1.GetOrderDetailRequest{OrderId: fmt.Sprintf("%d", id)}

	resp, err := h.clients.OrderClient.GetOrderDetail(context.Background(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get order detail"})
		return
	}

	c.JSON(200, resp)
}

// RegisterUser handles the HTTP request to register a user.
func (h *Handler) RegisterUser(c *gin.Context) {
	var req userv1.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.clients.UserClient.Register(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(200, resp)
}

// GetUserDetails handles the HTTP request to get a user's details.
func (h *Handler) GetUserDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	req := &userv1.GetUserRequest{UserId: id}

	resp, err := h.clients.UserClient.GetUser(context.Background(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user details"})
		return
	}

	c.JSON(200, resp)
}
