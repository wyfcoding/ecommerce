package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/search/application"   // 导入搜索模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/search/domain/entity" // 导入搜索模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                  // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Search模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.SearchService // 依赖Search应用服务，处理核心业务逻辑。
	logger  *slog.Logger               // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Search HTTP Handler 实例。
func NewHandler(service *application.SearchService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Search 处理搜索请求。支持POST（复杂过滤）和GET（简单关键词搜索）。
// Method: POST / GET
// Path: /search
func (h *Handler) Search(c *gin.Context) {
	// 定义请求体结构，用于接收搜索过滤条件。
	var req struct {
		Keyword    string   `json:"keyword"`     // 搜索关键词。
		CategoryID uint64   `json:"category_id"` // 分类ID。
		BrandID    uint64   `json:"brand_id"`    // 品牌ID。
		PriceMin   float64  `json:"price_min"`   // 价格下限。
		PriceMax   float64  `json:"price_max"`   // 价格上限。
		Sort       string   `json:"sort"`        // 排序方式。
		Page       int      `json:"page"`        // 页码。
		PageSize   int      `json:"page_size"`   // 每页数量。
		Tags       []string `json:"tags"`        // 标签。
		UserID     uint64   `json:"user_id"`     // 用户ID，可选，可从上下文或请求体获取。
	}

	// 尝试绑定JSON请求体。如果失败，则回退到从URL查询参数中获取。
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Keyword = c.Query("keyword")
		req.Sort = c.Query("sort")
		req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
		// 对于CategoryID, BrandID, PriceMin, PriceMax, Tags等，如果JSON绑定失败，则需要从查询参数中单独解析。
		// 这里简化处理，只解析了部分，复杂过滤器建议使用POST请求。
	}

	// TODO: 如果用户已登录，应从认证中间件设置的上下文（c.Get("userID")）中获取 userID，而不是从请求体。
	// userID := c.GetUint64("userID")

	// 构建SearchFilter实体。
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

	// 调用应用服务层执行搜索。
	result, err := h.service.Search(c.Request.Context(), req.UserID, filter)
	if err != nil {
		h.logger.Error("Failed to search", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to search", err.Error())
		return
	}

	// 返回成功的响应，包含搜索结果。
	response.SuccessWithStatus(c, http.StatusOK, "Search completed successfully", result)
}

// GetHotKeywords 处理获取热搜词列表的HTTP请求。
// Method: GET
// Path: /search/hot
func (h *Handler) GetHotKeywords(c *gin.Context) {
	// 从查询参数中获取数量限制，并设置默认值。
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用应用服务层获取热搜词。
	keywords, err := h.service.GetHotKeywords(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get hot keywords", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get hot keywords", err.Error())
		return
	}

	// 返回成功的响应，包含热搜词列表。
	response.SuccessWithStatus(c, http.StatusOK, "Hot keywords retrieved successfully", keywords)
}

// GetHistory 处理获取用户搜索历史的HTTP请求。
// Method: GET
// Path: /search/history
func (h *Handler) GetHistory(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}
	// 从查询参数中获取数量限制，并设置默认值。
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 调用应用服务层获取搜索历史。
	history, err := h.service.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to get search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get search history", err.Error())
		return
	}

	// 返回成功的响应，包含搜索历史列表。
	response.SuccessWithStatus(c, http.StatusOK, "Search history retrieved successfully", history)
}

// ClearHistory 处理清空用户搜索历史的HTTP请求。
// Method: DELETE
// Path: /search/history
func (h *Handler) ClearHistory(c *gin.Context) {
	// 定义请求体结构，用于接收用户ID。
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层清空搜索历史。
	if err := h.service.ClearSearchHistory(c.Request.Context(), req.UserID); err != nil {
		h.logger.Error("Failed to clear search history", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to clear search history", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Search history cleared successfully", nil)
}

// Suggest 处理获取搜索建议的HTTP请求。
// Method: GET
// Path: /search/suggest
func (h *Handler) Suggest(c *gin.Context) {
	// 从查询参数中获取关键词。
	keyword := c.Query("keyword")
	if keyword == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Keyword is required", "")
		return
	}

	// 调用应用服务层获取搜索建议。
	suggestions, err := h.service.Suggest(c.Request.Context(), keyword)
	if err != nil {
		h.logger.Error("Failed to get suggestions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get suggestions", err.Error())
		return
	}

	// 返回成功的响应，包含搜索建议列表。
	response.SuccessWithStatus(c, http.StatusOK, "Suggestions retrieved successfully", suggestions)
}

// RegisterRoutes 在给定的Gin路由组中注册Search模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /search 路由组，用于所有搜索相关接口。
	group := r.Group("/search")
	{
		group.POST("", h.Search)                 // POST用于支持更复杂的搜索过滤器（请求体）。
		group.GET("", h.Search)                  // GET用于简单的关键词搜索（查询参数）。
		group.GET("/hot", h.GetHotKeywords)      // 获取热搜词。
		group.GET("/history", h.GetHistory)      // 获取搜索历史。
		group.DELETE("/history", h.ClearHistory) // 清空搜索历史。
		group.GET("/suggest", h.Suggest)         // 获取搜索建议。
		// TODO: 补充获取搜索日志等接口。
	}
}
