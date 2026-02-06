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
	cmd    *application.MarketingCommandService
	query  *application.MarketingQueryService
	logger *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(cmd *application.MarketingCommandService, query *application.MarketingQueryService, logger *slog.Logger) *Handler {
	return &Handler{
		cmd:    cmd,
		query:  query,
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

	campaign, err := h.cmd.CreateCampaign(c.Request.Context(), req.Name, domain.CampaignType(req.Type), req.Description, req.StartTime, req.EndTime, req.Budget, req.Rules)
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

	campaign, err := h.query.GetCampaign(c.Request.Context(), id)
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

	list, total, err := h.query.ListCampaigns(c.Request.Context(), domain.CampaignStatus(status), page, pageSize)
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

	banner, err := h.cmd.CreateBanner(c.Request.Context(), req.Title, req.ImageURL, req.LinkURL, req.Position, req.Priority, req.StartTime, req.EndTime)
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

	list, err := h.query.ListBanners(c.Request.Context(), position, activeOnly)
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

	if err := h.cmd.ClickBanner(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to record banner click", "banner_id", id, "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateCampaignStatus 更新营销活动状态。
func (h *Handler) UpdateCampaignStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}
	if err := h.cmd.UpdateCampaignStatus(c.Request.Context(), id, domain.CampaignStatus(req.Status)); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update campaign status", "campaign_id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// RecordParticipation 记录活动参与。
func (h *Handler) RecordParticipation(c *gin.Context) {
	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		OrderID  uint64 `json:"order_id"`
		Discount uint64 `json:"discount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}
	if err := h.cmd.RecordParticipation(c.Request.Context(), campaignID, req.UserID, req.OrderID, req.Discount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to record participation", "campaign_id", campaignID, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// GetBanner 获取广告位详情。
func (h *Handler) GetBanner(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}
	banner, err := h.query.GetBanner(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get banner", "banner_id", id, "error", err)
		response.Error(c, err)
		return
	}
	response.Success(c, banner)
}

// DeleteBanner 删除广告位。
func (h *Handler) DeleteBanner(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id format", "")
		return
	}
	if err := h.cmd.DeleteBanner(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to delete banner", "banner_id", id, "error", err)
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
		group.PATCH("/campaigns/:id/status", h.UpdateCampaignStatus)
		group.POST("/campaigns/:id/participations", h.RecordParticipation)

		group.POST("/banners", h.CreateBanner)
		group.GET("/banners", h.ListBanners)
		group.GET("/banners/:id", h.GetBanner)
		group.POST("/banners/:id/click", h.ClickBanner)
		group.DELETE("/banners/:id", h.DeleteBanner)
	}
}
