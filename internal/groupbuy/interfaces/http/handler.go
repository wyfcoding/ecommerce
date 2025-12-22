package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Groupbuy模块的HTTP处理层。
type Handler struct {
	app    *application.GroupbuyService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Groupbuy HTTP Handler 实例。
func NewHandler(app *application.GroupbuyService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateGroupbuy 处理创建拼团活动的HTTP请求。
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

	groupbuy, err := h.app.CreateGroupbuy(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.GroupPrice,
		req.MinPeople, req.MaxPeople, req.TotalStock, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create groupbuy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create groupbuy", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Groupbuy created successfully", groupbuy)
}

// ListGroupbuys 处理获取拼团活动列表的HTTP请求。
func (h *Handler) ListGroupbuys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListGroupbuys(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list groupbuys", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list groupbuys", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Groupbuys listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// InitiateTeam 处理发起拼团团队的HTTP请求。
func (h *Handler) InitiateTeam(c *gin.Context) {
	var req struct {
		GroupbuyID uint64 `json:"groupbuy_id" binding:"required"`
		UserID     uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	team, order, err := h.app.InitiateTeam(c.Request.Context(), req.GroupbuyID, req.UserID)
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

// JoinTeam 处理加入拼团团队的HTTP请求。
func (h *Handler) JoinTeam(c *gin.Context) {
	var req struct {
		TeamNo string `json:"team_no" binding:"required"`
		UserID uint64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	order, err := h.app.JoinTeam(c.Request.Context(), req.TeamNo, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to join team", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to join team", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Joined team successfully", order)
}

// GetTeamDetails 处理获取拼团团队详情的HTTP请求。
func (h *Handler) GetTeamDetails(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	team, orders, err := h.app.GetTeamDetails(c.Request.Context(), teamID)
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

// RegisterRoutes 在给定的Gin路由组中注册Groupbuy模块的HTTP路由。
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
