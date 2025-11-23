package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/warehouse/application"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.WarehouseService
	logger  *slog.Logger
}

func NewHandler(service *application.WarehouseService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateWarehouse 创建仓库
func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	warehouse, err := h.service.CreateWarehouse(c.Request.Context(), req.Code, req.Name)
	if err != nil {
		h.logger.Error("Failed to create warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create warehouse", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Warehouse created successfully", warehouse)
}

// ListWarehouses 仓库列表
func (h *Handler) ListWarehouses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListWarehouses(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list warehouses", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list warehouses", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Warehouses listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateStock 更新库存
func (h *Handler) UpdateStock(c *gin.Context) {
	var req struct {
		WarehouseID uint64 `json:"warehouse_id" binding:"required"`
		SkuID       uint64 `json:"sku_id" binding:"required"`
		Quantity    int32  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.UpdateStock(c.Request.Context(), req.WarehouseID, req.SkuID, req.Quantity); err != nil {
		h.logger.Error("Failed to update stock", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Stock updated successfully", nil)
}

// CreateTransfer 创建调拨
func (h *Handler) CreateTransfer(c *gin.Context) {
	var req struct {
		FromID    uint64 `json:"from_id" binding:"required"`
		ToID      uint64 `json:"to_id" binding:"required"`
		SkuID     uint64 `json:"sku_id" binding:"required"`
		Quantity  int32  `json:"quantity" binding:"required"`
		CreatedBy uint64 `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	transfer, err := h.service.CreateTransfer(c.Request.Context(), req.FromID, req.ToID, req.SkuID, req.Quantity, req.CreatedBy)
	if err != nil {
		h.logger.Error("Failed to create transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create transfer", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Transfer created successfully", transfer)
}

// CompleteTransfer 完成调拨
func (h *Handler) CompleteTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.CompleteTransfer(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to complete transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete transfer", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Transfer completed successfully", nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/warehouse")
	{
		group.POST("", h.CreateWarehouse)
		group.GET("", h.ListWarehouses)
		group.POST("/stock", h.UpdateStock)
		group.POST("/transfer", h.CreateTransfer)
		group.POST("/transfer/:id/complete", h.CompleteTransfer)
	}
}
