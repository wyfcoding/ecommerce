package handler

import (
	"net/http"
	"strconv"

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

// RegisterRoutes registers all the routes for the search service.
func (h *SearchHandler) RegisterRoutes(e *gin.Engine) {
	api := e.Group("/api/v1/search")
	{
		api.GET("/products", h.SearchProducts)
	}
}

// SearchProducts is a Gin handler for searching products.
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	query := c.Query("query")
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
	pageToken, _ := strconv.ParseInt(c.DefaultQuery("page_token", "0"), 10, 32)

	products, total, err := h.service.SearchProducts(c.Request.Context(), query, int32(pageSize), int32(pageToken))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products, "total": total})
}