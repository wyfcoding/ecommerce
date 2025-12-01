package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.GroupbuyService
	logger  *slog.Logger
}

func NewHandler(service *application.GroupbuyService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateGroupbuy 创建拼团活动
func (h *Handler) CreateGroupbuy(c *gin.Context) {
	var req struct {
		Name          string    `json:"name" binding:"required"`
		ProductID     uint64    `json:"product_id" binding:"required"`
		SkuID         uint64    `json:"sku_id" binding:"required"`
		OriginalPrice uint64    `json:"original_price" binding:"required"`
		GroupPrice    uint64    `json:"group_price" binding:"required"`
		MinPeople     int32     `json:"min_people" binding:"required"`
		MaxPeople     int32     `json:"max_people" binding:"required"`
		TotalStock    int32     `json:"total_stock" binding:"required"`
		StartTime     time.Time `json:"start_time" binding:"required"`
		EndTime       time.Time `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	groupbuy, err := h.service.CreateGroupbuy(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.GroupPrice,
		req.MinPeople, req.MaxPeople, req.TotalStock, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create groupbuy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create groupbuy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Groupbuy created successfully", groupbuy)
}

// ListGroupbuys 获取拼团活动列表
func (h *Handler) ListGroupbuys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	groupbuys, total, err := h.service.ListGroupbuys(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list groupbuys", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list groupbuys", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Groupbuys listed successfully", gin.H{
		"data":      groupbuys,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// InitiateTeam 发起拼团
func (h *Handler) InitiateTeam(c *gin.Context) {
	var req struct {
		GroupbuyID uint64 `json:"groupbuy_id" binding:"required"`
		UserID     uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	team, order, err := h.service.InitiateTeam(c.Request.Context(), req.GroupbuyID, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to initiate team", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to initiate team", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Team initiated successfully", gin.H{
		"team":  team,
		"order": order,
	})
}

// JoinTeam 加入拼团
func (h *Handler) JoinTeam(c *gin.Context) {
	var req struct {
		TeamNo string `json:"team_no" binding:"required"`
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	order, err := h.service.JoinTeam(c.Request.Context(), req.TeamNo, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to join team", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to join team", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Joined team successfully", order)
}

// GetTeamDetails 获取团队详情
func (h *Handler) GetTeamDetails(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	team, orders, err := h.service.GetTeamDetails(c.Request.Context(), teamID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get team details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get team details", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Team details retrieved successfully", gin.H{
		"team":   team,
		"orders": orders,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/groupbuy")
	{
		group.POST("", h.CreateGroupbuy)
		group.GET("", h.ListGroupbuys)
		group.POST("/initiate", h.InitiateTeam)
		group.POST("/join", h.JoinTeam)
		group.GET("/team/:team_id", h.GetTeamDetails)
	}
}
