package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/warehouse/application"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Warehouse模块的HTTP处理层。
type Handler struct {
	app    *application.WarehouseService
	logger *slog.Logger
}

// NewHandler 创建并返回一个新的 Warehouse HTTP Handler 实例。
func NewHandler(app *application.WarehouseService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateWarehouse 处理创建仓库的HTTP请求。
func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	warehouse, err := h.app.CreateWarehouse(c.Request.Context(), req.Code, req.Name)
	if err != nil {
		h.logger.Error("Failed to create warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create warehouse", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Warehouse created successfully", warehouse)
}

// ListWarehouses 处理获取仓库列表的HTTP请求。
func (h *Handler) ListWarehouses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListWarehouses(c.Request.Context(), page, pageSize)
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

// GetWarehouse 处理获取仓库详情的HTTP请求。
func (h *Handler) GetWarehouse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	warehouse, err := h.app.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get warehouse", err.Error())
		return
	}
	if warehouse == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Warehouse not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Warehouse retrieved successfully", warehouse)
}

// UpdateStock 处理更新库存的HTTP请求。
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

	if err := h.app.UpdateStock(c.Request.Context(), req.WarehouseID, req.SkuID, req.Quantity); err != nil {
		h.logger.Error("Failed to update stock", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Stock updated successfully", nil)
}

// GetStock 处理获取库存详情的HTTP请求。
func (h *Handler) GetStock(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Query("warehouse_id"), 10, 64)
	if err != nil || warehouseID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid or missing warehouse_id", "")
		return
	}
	skuID, err := strconv.ParseUint(c.Query("sku_id"), 10, 64)
	if err != nil || skuID == 0 {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid or missing sku_id", "")
		return
	}

	stock, err := h.app.GetStock(c.Request.Context(), warehouseID, skuID)
	if err != nil {
		h.logger.Error("Failed to get stock", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get stock", err.Error())
		return
	}
	if stock == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Stock not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Stock retrieved successfully", stock)
}

// CreateTransfer 处理创建库存调拨单的HTTP请求。
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

	transfer, err := h.app.CreateTransfer(c.Request.Context(), req.FromID, req.ToID, req.SkuID, req.Quantity, req.CreatedBy)
	if err != nil {
		h.logger.Error("Failed to create transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create transfer", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Transfer created successfully", transfer)
}

// CompleteTransfer 处理完成库存调拨的HTTP请求。
func (h *Handler) CompleteTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.app.CompleteTransfer(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to complete transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete transfer", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Transfer completed successfully", nil)
}

// GetTransfer 处理获取调拨单详情的HTTP请求。
func (h *Handler) GetTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	transfer, err := h.app.GetTransfer(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get transfer", err.Error())
		return
	}
	if transfer == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Transfer not found", "")
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Transfer retrieved successfully", transfer)
}

// ListTransfers 处理获取调拨单列表的HTTP请求。
func (h *Handler) ListTransfers(c *gin.Context) {
	var fromID, toID uint64
	if fromStr := c.Query("from_warehouse_id"); fromStr != "" {
		fromID, _ = strconv.ParseUint(fromStr, 10, 64)
	}
	if toStr := c.Query("to_warehouse_id"); toStr != "" {
		toID, _ = strconv.ParseUint(toStr, 10, 64)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.app.ListTransfers(c.Request.Context(), fromID, toID, nil, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list transfers", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list transfers", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Transfers listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Warehouse模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/warehouse")
	{
		group.POST("", h.CreateWarehouse)
		group.GET("", h.ListWarehouses)
		group.GET("/:id", h.GetWarehouse)
		group.POST("/stock", h.UpdateStock)
		group.GET("/stock", h.GetStock)
		group.POST("/transfer", h.CreateTransfer)
		group.GET("/transfer", h.ListTransfers)
		group.GET("/transfer/:id", h.GetTransfer)
		group.POST("/transfer/:id/complete", h.CompleteTransfer)
	}
}
