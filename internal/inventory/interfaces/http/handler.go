package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.InventoryService
	logger  *slog.Logger
}

func NewHandler(service *application.InventoryService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateInventory 创建库存
func (h *Handler) CreateInventory(c *gin.Context) {
	var req struct {
		SkuID            uint64 `json:"sku_id" binding:"required"`
		ProductID        uint64 `json:"product_id" binding:"required"`
		WarehouseID      uint64 `json:"warehouse_id" binding:"required"`
		TotalStock       int32  `json:"total_stock" binding:"required"`
		WarningThreshold int32  `json:"warning_threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	inventory, err := h.service.CreateInventory(c.Request.Context(), req.SkuID, req.ProductID, req.WarehouseID, req.TotalStock, req.WarningThreshold)
	if err != nil {
		h.logger.Error("Failed to create inventory", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create inventory", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Inventory created successfully", inventory)
}

// GetInventory 获取库存
func (h *Handler) GetInventory(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	inventory, err := h.service.GetInventory(c.Request.Context(), skuID)
	if err != nil {
		h.logger.Error("Failed to get inventory", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get inventory", err.Error())
		return
	}
	if inventory == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Inventory not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Inventory retrieved successfully", inventory)
}

// UpdateStock 更新库存 (Add/Deduct/Lock/Unlock/Confirm)
func (h *Handler) UpdateStock(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	var req struct {
		Action   string `json:"action" binding:"required,oneof=add deduct lock unlock confirm"`
		Quantity int32  `json:"quantity" binding:"required,gt=0"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error
	ctx := c.Request.Context()

	switch req.Action {
	case "add":
		opErr = h.service.AddStock(ctx, skuID, req.Quantity, req.Reason)
	case "deduct":
		opErr = h.service.DeductStock(ctx, skuID, req.Quantity, req.Reason)
	case "lock":
		opErr = h.service.LockStock(ctx, skuID, req.Quantity, req.Reason)
	case "unlock":
		opErr = h.service.UnlockStock(ctx, skuID, req.Quantity, req.Reason)
	case "confirm":
		opErr = h.service.ConfirmDeduction(ctx, skuID, req.Quantity, req.Reason)
	}

	if opErr != nil {
		h.logger.Error("Failed to update stock", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update stock", opErr.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Stock updated successfully", nil)
}

// ListInventories 获取库存列表
func (h *Handler) ListInventories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListInventories(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list inventories", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list inventories", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Inventories listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/inventory")
	{
		group.POST("", h.CreateInventory)
		group.GET("", h.ListInventories)
		group.GET("/:sku_id", h.GetInventory)
		group.POST("/:sku_id/stock", h.UpdateStock)
	}
}
