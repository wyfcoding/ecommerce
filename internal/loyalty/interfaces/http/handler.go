package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/loyalty/application"
	"ecommerce/internal/loyalty/domain/entity"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.LoyaltyService
	logger  *slog.Logger
}

func NewHandler(service *application.LoyaltyService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetAccount 获取会员账户
func (h *Handler) GetAccount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	account, err := h.service.GetOrCreateAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get account", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Account retrieved successfully", account)
}

// UpdatePoints 更新积分 (Add/Deduct)
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
		opErr = h.service.AddPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	} else {
		opErr = h.service.DeductPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	}

	if opErr != nil {
		h.logger.Error("Failed to update points", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update points", opErr.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Points updated successfully", nil)
}

// GetTransactions 获取交易记录
func (h *Handler) GetTransactions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.GetPointsTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list transactions", "error", err)
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

// AddBenefit 添加权益
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

	benefit, err := h.service.AddBenefit(c.Request.Context(), entity.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		h.logger.Error("Failed to add benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add benefit", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Benefit added successfully", benefit)
}

// ListBenefits 获取权益列表
func (h *Handler) ListBenefits(c *gin.Context) {
	level := c.Query("level")
	list, err := h.service.ListBenefits(c.Request.Context(), entity.MemberLevel(level))
	if err != nil {
		h.logger.Error("Failed to list benefits", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list benefits", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Benefits listed successfully", list)
}

// DeleteBenefit 删除权益
func (h *Handler) DeleteBenefit(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteBenefit(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete benefit", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Benefit deleted successfully", nil)
}

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
