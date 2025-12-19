package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/search/application"
	"github.com/wyfcoding/ecommerce/internal/search/domain"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Search模块的HTTP处理层。
type Handler struct {
	app    *application.SearchService // 依赖Search应用服务 (Facade)。
	logger *slog.Logger               // 日志记录器。
}

// NewHandler 创建并返回一个新的 Search HTTP Handler 实例。
func NewHandler(app *application.SearchService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// Search 处理搜索请求。
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
		UserID     uint64   `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Keyword = c.Query("keyword")
		req.Sort = c.Query("sort")
		req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	}

	filter := &domain.SearchFilter{
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

	result, err := h.app.Search(c.Request.Context(), req.UserID, filter)
	if err != nil {
		h.logger.Error("Failed to search", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to search", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search completed successfully", result)
}

// GetHotKeywords 处理获取热搜词列表的HTTP请求。
func (h *Handler) GetHotKeywords(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	keywords, err := h.app.GetHotKeywords(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get hot keywords", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get hot keywords", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Hot keywords retrieved successfully", keywords)
}

// GetHistory 处理获取用户搜索历史的HTTP请求。
func (h *Handler) GetHistory(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	history, err := h.app.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to get search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get search history", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search history retrieved successfully", history)
}

// ClearHistory 处理清空用户搜索历史的HTTP请求。
func (h *Handler) ClearHistory(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.app.ClearSearchHistory(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to clear search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear search history", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Search history cleared successfully", nil)
}

// Suggest 处理获取搜索建议的HTTP请求。
func (h *Handler) Suggest(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Keyword is required", "")
		return
	}

	suggestions, err := h.app.Suggest(c.Request.Context(), keyword)
	if err != nil {
		h.logger.Error("Failed to get suggestions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get suggestions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Suggestions retrieved successfully", suggestions)
}

// RegisterRoutes 注册路由.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/search")
	{
		group.POST("", h.Search)
		group.GET("", h.Search)
		group.GET("/hot", h.GetHotKeywords)
		group.GET("/history", h.GetHistory)
		group.DELETE("/history", h.ClearHistory)
		group.GET("/suggest", h.Suggest)
	}
}
