package handler

import (
	"net/http"

	"ecommerce/api/order/v1"
	"ecommerce/internal/order/service"
	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for the order service.
type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

// CreateOrder is a Gin handler for creating an order.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req v1.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}