package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/application"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.MarketingService
	logger  *slog.Logger
}

func NewHandler(service *application.MarketingService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateCampaign 创建活动
func (h *Handler) CreateCampaign(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Type        string                 `json:"type" binding:"required"`
		Description string                 `json:"description"`
		StartTime   time.Time              `json:"start_time" binding:"required"`
		EndTime     time.Time              `json:"end_time" binding:"required"`
		Budget      uint64                 `json:"budget" binding:"required"`
		Rules       map[string]interface{} `json:"rules"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	campaign, err := h.service.CreateCampaign(c.Request.Context(), req.Name, entity.CampaignType(req.Type), req.Description, req.StartTime, req.EndTime, req.Budget, req.Rules)
	if err != nil {
		h.logger.Error("Failed to create campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create campaign", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Campaign created successfully", campaign)
}

// GetCampaign 获取活动
func (h *Handler) GetCampaign(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	campaign, err := h.service.GetCampaign(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get campaign", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Campaign retrieved successfully", campaign)
}

// ListCampaigns 获取活动列表
func (h *Handler) ListCampaigns(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListCampaigns(c.Request.Context(), entity.CampaignStatus(status), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list campaigns", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list campaigns", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Campaigns listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateBanner 创建Banner
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	banner, err := h.service.CreateBanner(c.Request.Context(), req.Title, req.ImageURL, req.LinkURL, req.Position, req.Priority, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.Error("Failed to create banner", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create banner", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Banner created successfully", banner)
}

// ListBanners 获取Banner列表
func (h *Handler) ListBanners(c *gin.Context) {
	position := c.Query("position")
	activeOnly := c.Query("active_only") == "true"

	list, err := h.service.ListBanners(c.Request.Context(), position, activeOnly)
	if err != nil {
		h.logger.Error("Failed to list banners", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list banners", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Banners listed successfully", list)
}

// ClickBanner 点击Banner
func (h *Handler) ClickBanner(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.ClickBanner(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to record click", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record click", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Click recorded successfully", nil)
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
