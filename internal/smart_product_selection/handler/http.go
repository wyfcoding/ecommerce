package handler

import (
	"net/http"

	"ecommerce/internal/smart_product_selection/service"
	"github.com/gin-gonic/gin"
)

// SmartProductSelectionHandler handles HTTP requests for the smart_product_selection service.
type SmartProductSelectionHandler struct {
	service *service.SmartProductSelectionService
}

// NewSmartProductSelectionHandler creates a new SmartProductSelectionHandler.
func NewSmartProductSelectionHandler(s *service.SmartProductSelectionService) *SmartProductSelectionHandler {
	return &SmartProductSelectionHandler{service: s}
}

// SelectProducts is a Gin handler for selecting products.
func (h *SmartProductSelectionHandler) SelectProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "select products placeholder"})
}
