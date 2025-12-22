package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"github.com/wyfcoding/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 处理 HTTP 或 gRPC 请求。
type Handler struct {
	service *application.PointsmallService
	logger  *slog.Logger
}

// NewHandler 处理 HTTP 或 gRPC 请求。
func NewHandler(service *application.PointsmallService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req domain.PointsProduct
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.CreateProduct(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Product created successfully", req)
}

func (h *Handler) ListProducts(c *gin.Context) {
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListProducts(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list products", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list products", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Products listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) ExchangeProduct(c *gin.Context) {
	var req struct {
		UserID    uint64 `json:"user_id" binding:"required"`
		ProductID uint64 `json:"product_id" binding:"required"`
		Quantity  int32  `json:"quantity" binding:"required"`
		Address   string `json:"address" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
		Receiver  string `json:"receiver" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	order, err := h.service.ExchangeProduct(c.Request.Context(), req.UserID, req.ProductID, req.Quantity, req.Address, req.Phone, req.Receiver)
	if err != nil {
		h.logger.Error("Failed to exchange product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange product", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Product exchanged successfully", order)
}

func (h *Handler) GetAccount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	account, err := h.service.GetAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get account", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Account retrieved successfully", account)
}

func (h *Handler) AddPoints(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		Points      int64  `json:"points" binding:"required"`
		Description string `json:"description"`
		RefID       string `json:"ref_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AddPoints(c.Request.Context(), req.UserID, req.Points, req.Description, req.RefID); err != nil {
		h.logger.Error("Failed to add points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add points", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Points added successfully", nil)
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.service.ListOrders(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/pointsmall")
	{
		group.POST("/products", h.CreateProduct)
		group.GET("/products", h.ListProducts)
		group.POST("/exchange", h.ExchangeProduct)
		group.GET("/orders", h.ListOrders)
		group.GET("/account", h.GetAccount)
		group.POST("/points", h.AddPoints)
	}
}
