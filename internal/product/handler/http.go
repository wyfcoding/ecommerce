package handler

import (
	"net/http"

	"ecommerce/api/product/v1"
	"ecommerce/internal/product/service"
	"github.com/gin-gonic/gin"
)

// ProductHandler handles HTTP requests for the product service.
type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(s *service.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

// GetProduct is a Gin handler for getting a product by ID.
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &v1.GetProductRequest{Id: id}
	product, err := h.service.GetProduct(c.Request.Context(), grpcReq)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}