
package handler

import (
	"context"
	"strconv"

	productv1 "ecommerce/api/product/v1"
	orderv1 "ecommerce/api/order/v1"
	userv1 "ecommerce/api/user/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// Clients holds the gRPC clients for downstream services.
type Clients struct {
	Product productv1.ProductServiceClient
	Order   orderv1.OrderClient
	User    userv1.UserClient
}

// NewClients creates a new Clients struct.
func NewClients(productConn, orderConn, userConn *grpc.ClientConn) *Clients {
	return &Clients{
		Product: productv1.NewProductServiceClient(productConn),
		Order:   orderv1.NewOrderClient(orderConn),
		User:    userv1.NewUserClient(userConn),
	}
}

// Handler holds the gRPC clients.
type Handler struct {
	clients *Clients
}

// NewHandler creates a new Handler.
func NewHandler(clients *Clients) *Handler {
	return &Handler{clients: clients}
}

// RegisterRoutes registers all the routes for the gateway.
func (h *Handler) RegisterRoutes(e *gin.Engine) {
	v1 := e.Group("/v1")
	{
		// Product routes
		productGroup := v1.Group("/products")
		{
			productGroup.GET("/:id", h.GetProductDetail)
			// Add other product routes here...
		}

		// Add other service groups here...
	}
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

	resp, err := h.clients.Product.GetSpuDetail(context.Background(), req)
	if err != nil {
		// In a real app, you'd check the gRPC error code and return a more specific HTTP status.
		c.JSON(500, gin.H{"error": "Failed to get product detail"})
		return
	}

	c.JSON(200, resp)
}
