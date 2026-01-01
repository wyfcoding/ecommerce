package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/application"
	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.AdvancedCouponService
	logger  *slog.Logger
}

func NewHandler(service *application.AdvancedCouponService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string    `json:"code" binding:"required"`
		Type          string    `json:"type" binding:"required"`
		DiscountValue int64     `json:"discount_value" binding:"required"`
		ValidFrom     time.Time `json:"valid_from" binding:"required"`
		ValidUntil    time.Time `json:"valid_until" binding:"required"`
		TotalQuantity int64     `json:"total_quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request: "+err.Error(), "")
		return
	}

	coupon, err := h.service.CreateCoupon(c.Request.Context(), req.Code, domain.CouponType(req.Type), req.DiscountValue, req.ValidFrom, req.ValidUntil, req.TotalQuantity)
	if err != nil {
		h.logger.Error("Failed to create coupon", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, coupon)
}

func (h *Handler) ListCoupons(c *gin.Context) {
	status := c.Query("status")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page parameter", "")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid page_size parameter", "")
		return
	}

	list, total, err := h.service.ListCoupons(c.Request.Context(), domain.CouponStatus(status), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list coupons", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *Handler) UseCoupon(c *gin.Context) {
	var req struct {
		UserID  uint64 `json:"user_id" binding:"required"`
		Code    string `json:"code" binding:"required"`
		OrderID uint64 `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request: "+err.Error(), "")
		return
	}

	if err := h.service.UseCoupon(c.Request.Context(), req.UserID, req.Code, req.OrderID); err != nil {
		h.logger.Error("Failed to use coupon", "user_id", req.UserID, "code", req.Code, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/advanced-coupons")
	{
		group.POST("/use", h.UseCoupon)
		group.POST("", h.CreateCoupon)
		group.GET("", h.ListCoupons)
	}
}
