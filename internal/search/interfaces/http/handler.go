package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/search/application"
	"ecommerce/internal/search/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.SearchService
	logger  *slog.Logger
}

func NewHandler(service *application.SearchService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Search 搜索
func (h *Handler) Search(c *gin.Context) {
	var req struct {
		Keyword    string   `json:"keyword"`
		CategoryID uint64   `json:"category_id"`
		BrandID    uint64   `json:"brand_id"`
		PriceMin   float64  `json:"price_min"`
		PriceMax   float64  `json:"price_max"`
		Sort       string   `json:"sort"`
		Page       int      `json:"page"`
		PageSize   int      `json:"page_size"`
		Tags       []string `json:"tags"`
		UserID     uint64   `json:"user_id"` // Optional, from context or body
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback to query params for simple search
		req.Keyword = c.Query("keyword")
		req.Sort = c.Query("sort")
		req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	}

	// If UserID is in context (from auth middleware), use it
	// userID := c.GetUint64("userID")

	filter := &entity.SearchFilter{
		Keyword:    req.Keyword,
		CategoryID: req.CategoryID,
		BrandID:    req.BrandID,
		PriceMin:   req.PriceMin,
		PriceMax:   req.PriceMax,
		Sort:       req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Tags:       req.Tags,
	}

	result, err := h.service.Search(c.Request.Context(), req.UserID, filter)
	if err != nil {
		h.logger.Error("Failed to search", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to search", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search completed successfully", result)
}

// GetHotKeywords 获取热搜
func (h *Handler) GetHotKeywords(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	keywords, err := h.service.GetHotKeywords(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get hot keywords", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get hot keywords", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Hot keywords retrieved successfully", keywords)
}

// GetHistory 获取历史
func (h *Handler) GetHistory(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	history, err := h.service.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to get search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get search history", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search history retrieved successfully", history)
}

// ClearHistory 清空历史
func (h *Handler) ClearHistory(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.ClearSearchHistory(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to clear search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear search history", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search history cleared successfully", nil)
}

// Suggest 建议
func (h *Handler) Suggest(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Keyword is required", "")
		return
	}

	suggestions, err := h.service.Suggest(c.Request.Context(), keyword)
	if err != nil {
		h.logger.Error("Failed to get suggestions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get suggestions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Suggestions retrieved successfully", suggestions)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/search")
	{
		group.POST("", h.Search) // POST for complex filter
		group.GET("", h.Search)  // GET for simple search
		group.GET("/hot", h.GetHotKeywords)
		group.GET("/history", h.GetHistory)
		group.DELETE("/history", h.ClearHistory)
		group.GET("/suggest", h.Suggest)
	}
}
