package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/loyalty/application"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Loyalty模块的HTTP处理层。
type Handler struct {
	app    *application.LoyaltyService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Loyalty HTTP Handler 实例。
func NewHandler(app *application.LoyaltyService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// GetAccount 处理获取会员账户信息的HTTP请求。
func (h *Handler) GetAccount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	account, err := h.app.GetOrCreateAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get account", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Account retrieved successfully", account)
}

// UpdatePoints 处理更新积分的HTTP请求（增加或扣减）。
func (h *Handler) UpdatePoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	var req struct {
		Action      string `json:"action" binding:"required,oneof=add deduct"`
		Points      int64  `json:"points" binding:"required,gt=0"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description"`
		OrderID     uint64 `json:"order_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error
	ctx := c.Request.Context()

	if req.Action == "add" {
		opErr = h.app.AddPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	} else {
		opErr = h.app.DeductPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	}

	if opErr != nil {
		h.logger.ErrorContext(ctx, "Failed to update points", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update points", opErr.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Points updated successfully", nil)
}

// GetTransactions 处理获取用户积分交易记录的HTTP请求。
func (h *Handler) GetTransactions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.GetPointsTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list transactions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list transactions", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Transactions listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddBenefit 处理添加会员权益的HTTP请求。
func (h *Handler) AddBenefit(c *gin.Context) {
	var req struct {
		Level        string  `json:"level" binding:"required"`
		Name         string  `json:"name" binding:"required"`
		Description  string  `json:"description"`
		DiscountRate float64 `json:"discount_rate"`
		PointsRate   float64 `json:"points_rate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	benefit, err := h.app.AddBenefit(c.Request.Context(), domain.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add benefit", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Benefit added successfully", benefit)
}

// ListBenefits 处理获取会员权益列表的HTTP请求。
func (h *Handler) ListBenefits(c *gin.Context) {
	level := c.Query("level")
	list, err := h.app.ListBenefits(c.Request.Context(), domain.MemberLevel(level))
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list benefits", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list benefits", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Benefits listed successfully", list)
}

// DeleteBenefit 处理删除会员权益的HTTP请求。
func (h *Handler) DeleteBenefit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.DeleteBenefit(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete benefit", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Benefit deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Loyalty模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/loyalty")
	{
		group.GET("/accounts/:user_id", h.GetAccount)
		group.POST("/accounts/:user_id/points", h.UpdatePoints)
		group.GET("/accounts/:user_id/transactions", h.GetTransactions)
		group.POST("/benefits", h.AddBenefit)
		group.GET("/benefits", h.ListBenefits)
		group.DELETE("/benefits/:id", h.DeleteBenefit)
	}
}
