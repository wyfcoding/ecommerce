package handler

import (
	"net/http"

	"ecommerce/internal/search/service"
	"github.com/gin-gonic/gin"
)

// SearchHandler handles HTTP requests for the search service.
type SearchHandler struct {
	service *service.SearchService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(s *service.SearchService) *SearchHandler {
	return &SearchHandler{service: s}
}

// SearchProducts is a Gin handler for searching products.
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "search placeholder"})
}