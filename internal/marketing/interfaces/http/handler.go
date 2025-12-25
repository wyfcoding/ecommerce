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

func (h *Handler) CreateCampaign(c *gin.Context) {
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
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	campaign, err := h.app.CreateCampaign(c.Request.Context(), req.Name, domain.CampaignType(req.Type), req.Description, req.StartTime, req.EndTime, req.Budget, req.Rules)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create campaign", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Campaign created successfully", campaign)
}

func (h *Handler) GetCampaign(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	campaign, err := h.app.GetCampaign(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get campaign", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get campaign", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Campaign retrieved successfully", campaign)
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListCampaigns(c.Request.Context(), domain.CampaignStatus(status), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list campaigns", "error", err)
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

	banner, err := h.app.CreateBanner(c.Request.Context(), req.Title, req.ImageURL, req.LinkURL, req.Position, req.Priority, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create banner", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create banner", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Banner created successfully", banner)
}

func (h *Handler) ListBanners(c *gin.Context) {
	position := c.Query("position")
	activeOnly := c.Query("active_only") == "true"

	list, err := h.app.ListBanners(c.Request.Context(), position, activeOnly)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list banners", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list banners", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Banners listed successfully", list)
}

func (h *Handler) ClickBanner(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.ClickBanner(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to record click", "error", err)
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
