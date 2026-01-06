package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/application"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	app    *application.Marketing
	logger *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(app *application.Marketing, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateCampaign 处理创建营销活动的 HTTP 请求。
func (h *Handler) CreateCampaign(c *gin.Context) {
	// ... (请求参数定义保持不变) ...
	var req struct {
		Name        string         `json:"name" binding:"required"`
		Type        string         `json:"type" binding:"required"`
		Description string         `json:"description"`
		StartTime   time.Time      `json:"start_time" binding:"required"`
		EndTime     time.Time      `json:"end_time" binding:"required"`
		Budget      uint64         `json:"budget" binding:"required"`
		Rules       map[string]any `json:"rules"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	campaign, err := h.app.CreateCampaign(c.Request.Context(), req.Name, domain.CampaignType(req.Type), req.Description, req.StartTime, req.EndTime, req.Budget, req.Rules)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create campaign", "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "campaign created successfully", campaign)
}

// GetCampaign 获取营销活动的详细信息。
func (h *Handler) GetCampaign(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	campaign, err := h.app.GetCampaign(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get campaign detail", "campaign_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, campaign)
}

// ListCampaigns 分页列出营销活动。
func (h *Handler) ListCampaigns(c *gin.Context) {
	status, err := strconv.Atoi(c.DefaultQuery("status", "0"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid status", err.Error())
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := h.app.ListCampaigns(c.Request.Context(), domain.CampaignStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list campaigns", "status", status, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, list, total, int32(page), int32(pageSize))
}

// CreateBanner 处理创建横幅广告的请求。
func (h *Handler) CreateBanner(c *gin.Context) {
	var req struct {
		Title     string    `json:"title" binding:"required"`
		ImageURL  string    `json:"image_url" binding:"required"`
		LinkURL   string    `json:"link_url"`
		Position  string    `json:"position" binding:"required"`
		Priority  int32     `json:"priority"`
		StartTime time.Time `json:"start_time" binding:"required"`
		EndTime   time.Time `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	banner, err := h.app.CreateBanner(c.Request.Context(), req.Title, req.ImageURL, req.LinkURL, req.Position, req.Priority, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create banner", "title", req.Title, "error", err)
		response.Error(c, err)
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "banner created successfully", banner)
}

// ListBanners 根据位置获取活跃的横幅列表。
func (h *Handler) ListBanners(c *gin.Context) {
	position := c.Query("position")
	activeOnly := c.Query("active_only") == "true"

	list, err := h.app.ListBanners(c.Request.Context(), position, activeOnly)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to list banners", "position", position, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, list)
}

// ClickBanner 处理横幅点击统计请求。
func (h *Handler) ClickBanner(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}

	if err := h.app.ClickBanner(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to record banner click", "banner_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/marketing")
	{
		group.POST("/campaigns", h.CreateCampaign)
		group.GET("/campaigns", h.ListCampaigns)
		group.GET("/campaigns/:id", h.GetCampaign)

		group.POST("/banners", h.CreateBanner)
		group.GET("/banners", h.ListBanners)
		group.POST("/banners/:id/click", h.ClickBanner)
	}
}
