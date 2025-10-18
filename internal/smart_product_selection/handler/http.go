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

// RegisterRoutes registers all the routes for the smart product selection service.
func (h *SmartProductSelectionHandler) RegisterRoutes(e *gin.Engine) {
	api := e.Group("/api/v1/smart-product-selection")
	{
		api.POST("/recommendations", getSmartProductRecommendationsHandler(h.service))
	}
}

func getSmartProductRecommendationsHandler(s *service.SmartProductSelectionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MerchantID      string            `json:"merchant_id"`
			ContextFeatures map[string]string `json:"context_features"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		recommendations, err := s.GetSmartProductRecommendations(c.Request.Context(), req.MerchantID, req.ContextFeatures)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, recommendations)
	}
}
