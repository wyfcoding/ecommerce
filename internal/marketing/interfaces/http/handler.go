package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/marketing/application"   // 导入营销模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity" // 导入营销模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                     // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Marketing模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.MarketingService // 依赖Marketing应用服务，处理核心业务逻辑。
	logger  *slog.Logger                  // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Marketing HTTP Handler 实例。
func NewHandler(service *application.MarketingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateCampaign 处理创建营销活动的HTTP请求。
// Method: POST
// Path: /marketing/campaigns
func (h *Handler) CreateCampaign(c *gin.Context) {
	// 定义请求体结构，用于接收营销活动的创建信息。
	var req struct {
		Name        string                 `json:"name" binding:"required"`       // 活动名称，必填。
		Type        string                 `json:"type" binding:"required"`       // 活动类型，必填。
		Description string                 `json:"description"`                   // 描述，选填。
		StartTime   time.Time              `json:"start_time" binding:"required"` // 开始时间，必填。
		EndTime     time.Time              `json:"end_time" binding:"required"`   // 结束时间，必填。
		Budget      uint64                 `json:"budget" binding:"required"`     // 预算，必填。
		Rules       map[string]interface{} `json:"rules"`                         // 规则配置，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建营销活动。
	campaign, err := h.service.CreateCampaign(c.Request.Context(), req.Name, entity.CampaignType(req.Type), req.Description, req.StartTime, req.EndTime, req.Budget, req.Rules)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create campaign", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Campaign created successfully", campaign)
}

// GetCampaign 处理获取营销活动详情的HTTP请求。
// Method: GET
// Path: /marketing/campaigns/:id
func (h *Handler) GetCampaign(c *gin.Context) {
	// 从URL路径中解析活动ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取营销活动详情。
	campaign, err := h.service.GetCampaign(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get campaign", err.Error())
		return
	}

	// 返回成功的响应，包含营销活动详情。
	response.SuccessWithStatus(c, http.StatusOK, "Campaign retrieved successfully", campaign)
}

// ListCampaigns 处理获取营销活动列表的HTTP请求。
// Method: GET
// Path: /marketing/campaigns
func (h *Handler) ListCampaigns(c *gin.Context) {
	// 从查询参数中获取状态、页码和每页大小，并设置默认值。
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取营销活动列表。
	list, total, err := h.service.ListCampaigns(c.Request.Context(), entity.CampaignStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list campaigns", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list campaigns", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Campaigns listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateBanner 处理创建广告横幅的HTTP请求。
// Method: POST
// Path: /marketing/banners
func (h *Handler) CreateBanner(c *gin.Context) {
	// 定义请求体结构，用于接收Banner的创建信息。
	var req struct {
		Title     string    `json:"title" binding:"required"`      // 标题，必填。
		ImageURL  string    `json:"image_url" binding:"required"`  // 图片URL，必填。
		LinkURL   string    `json:"link_url"`                      // 跳转URL，选填。
		Position  string    `json:"position" binding:"required"`   // 位置，必填。
		Priority  int32     `json:"priority"`                      // 优先级，选填。
		StartTime time.Time `json:"start_time" binding:"required"` // 开始时间，必填。
		EndTime   time.Time `json:"end_time" binding:"required"`   // 结束时间，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建Banner。
	banner, err := h.service.CreateBanner(c.Request.Context(), req.Title, req.ImageURL, req.LinkURL, req.Position, req.Priority, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create banner", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create banner", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Banner created successfully", banner)
}

// ListBanners 处理获取广告横幅列表的HTTP请求。
// Method: GET
// Path: /marketing/banners
func (h *Handler) ListBanners(c *gin.Context) {
	// 从查询参数中获取位置和是否只显示活跃横幅。
	position := c.Query("position")
	activeOnly := c.Query("active_only") == "true"

	// 调用应用服务层获取Banner列表。
	list, err := h.service.ListBanners(c.Request.Context(), position, activeOnly)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list banners", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list banners", err.Error())
		return
	}

	// 返回成功的响应，包含Banner列表。
	response.SuccessWithStatus(c, http.StatusOK, "Banners listed successfully", list)
}

// ClickBanner 处理记录广告横幅点击事件的HTTP请求。
// Method: POST
// Path: /marketing/banners/:id/click
func (h *Handler) ClickBanner(c *gin.Context) {
	// 从URL路径中解析Banner ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层记录Banner点击事件。
	if err := h.service.ClickBanner(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to record click", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record click", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Click recorded successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Marketing模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /marketing 路由组，用于所有营销活动和Banner管理相关接口。
	group := r.Group("/marketing")
	{
		// 营销活动接口。
		group.POST("/campaigns", h.CreateCampaign) // 创建营销活动。
		group.GET("/campaigns", h.ListCampaigns)   // 获取营销活动列表。
		group.GET("/campaigns/:id", h.GetCampaign) // 获取营销活动详情。
		// TODO: 补充更新活动状态、更新活动、删除活动、记录参与等接口。

		// 广告横幅接口。
		group.POST("/banners", h.CreateBanner)          // 创建Banner。
		group.GET("/banners", h.ListBanners)            // 获取Banner列表。
		group.POST("/banners/:id/click", h.ClickBanner) // 记录Banner点击事件。
		// TODO: 补充更新Banner、删除Banner、启用/禁用Banner等接口。
	}
}
