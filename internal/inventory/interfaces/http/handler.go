package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/inventory/application" // 导入库存模块的应用服务。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Inventory模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	app    *application.Inventory // 依赖Inventory应用服务，处理核心业务逻辑。
	logger *slog.Logger           // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Inventory HTTP Handler 实例。
func NewHandler(app *application.Inventory, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateInventory 处理创建库存记录的HTTP请求。
// HTTP 方法: POST
// 请求路径: /inventory
func (h *Handler) CreateInventory(c *gin.Context) {
	// 定义请求体结构，用于接收库存创建信息。
	var req struct {
		SkuID            uint64 `json:"sku_id" binding:"required"`       // SKU ID，必填。
		ProductID        uint64 `json:"product_id" binding:"required"`   // 商品ID，必填。
		WarehouseID      uint64 `json:"warehouse_id" binding:"required"` // 仓库ID，必填。
		TotalStock       int32  `json:"total_stock" binding:"required"`  // 总库存量，必填。
		WarningThreshold int32  `json:"warning_threshold"`               // 预警阈值，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建库存。
	inventory, err := h.app.CreateInventory(c.Request.Context(), req.SkuID, req.ProductID, req.WarehouseID, req.TotalStock, req.WarningThreshold)
	if err != nil {
		h.logger.Error("Failed to create inventory", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create inventory", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Inventory created successfully", inventory)
}

// GetInventory 处理获取指定SKU库存信息的HTTP请求。
// HTTP 方法: GET
// 请求路径: /inventory/:sku_id
func (h *Handler) GetInventory(c *gin.Context) {
	// 从URL路径中解析SKU ID。
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	// 调用应用服务层获取库存信息。
	inventory, err := h.app.GetInventory(c.Request.Context(), skuID)
	if err != nil {
		h.logger.Error("Failed to get inventory", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get inventory", err.Error())
		return
	}
	// 如果库存记录未找到，返回404。
	if inventory == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Inventory not found", "")
		return
	}

	// 返回成功的响应，包含库存信息。
	response.SuccessWithStatus(c, http.StatusOK, "Inventory retrieved successfully", inventory)
}

// UpdateStock 处理更新库存数量的HTTP请求（增加、扣减、锁定、解锁、确认扣减）。
// HTTP 方法: POST
// 请求路径: /inventory/:sku_id/stock
func (h *Handler) UpdateStock(c *gin.Context) {
	// 从URL路径中解析SKU ID。
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收操作类型、数量和原因。
	var req struct {
		Action   string `json:"action" binding:"required,oneof=add deduct lock unlock confirm"` // 操作类型，必填。
		Quantity int32  `json:"quantity" binding:"required,gt=0"`                               // 数量，必填且必须大于0。
		Reason   string `json:"reason"`                                                         // 原因，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error // 操作结果错误。
	ctx := c.Request.Context()

	// 根据操作类型调用应用服务层的相应方法。
	switch req.Action {
	case "add":
		opErr = h.app.AddStock(ctx, skuID, req.Quantity, req.Reason)
	case "deduct":
		opErr = h.app.DeductStock(ctx, skuID, req.Quantity, req.Reason)
	case "lock":
		opErr = h.app.LockStock(ctx, skuID, req.Quantity, req.Reason)
	case "unlock":
		opErr = h.app.UnlockStock(ctx, skuID, req.Quantity, req.Reason)
	case "confirm":
		opErr = h.app.ConfirmDeduction(ctx, skuID, req.Quantity, req.Reason)
	}

	if opErr != nil {
		h.logger.Error("Failed to update stock", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update stock", opErr.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Stock updated successfully", nil)
}

// ListInventories 处理获取库存列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /inventory
func (h *Handler) ListInventories(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 调用应用服务层获取库存列表。
	list, total, err := h.app.ListInventories(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list inventories", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list inventories", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Inventories listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteInventory 处理删除库存记录的HTTP请求。
func (h *Handler) DeleteInventory(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	if err := h.app.DeleteInventory(c.Request.Context(), skuID); err != nil {
		h.logger.Error("Failed to delete inventory", "sku_id", skuID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete inventory", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Inventory deleted successfully", nil)
}

// GetInventoryLogs 处理获取库存变更日志的HTTP请求。
func (h *Handler) GetInventoryLogs(c *gin.Context) {
	skuID, err := strconv.ParseUint(c.Param("sku_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid SKU ID", err.Error())
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 这里的业务逻辑可能需要先根据 sku_id 查到 inventory_id，
	// 但 domain.Inventory 模型中 sku_id 已经是 key。
	// 这里假设 GetInventoryLogs 接受的第一个参数是 inventory_id (uint64)，
	// 我们可以先获取 inventory 实体。
	inv, err := h.app.GetInventory(c.Request.Context(), skuID)
	if err != nil || inv == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "Inventory not found", "")
		return
	}

	list, total, err := h.app.GetInventoryLogs(c.Request.Context(), uint64(inv.ID), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get inventory logs", "sku_id", skuID, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get logs", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Inventory logs retrieved successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Inventory模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /inventory 路由组，用于所有库存相关接口。
	group := r.Group("/inventory")
	{
		group.POST("", h.CreateInventory)           // 创建库存。
		group.GET("", h.ListInventories)            // 获取库存列表。
		group.GET("/:sku_id", h.GetInventory)       // 获取指定SKU库存信息。
		group.DELETE("/:sku_id", h.DeleteInventory) // 删除库存。
		group.POST("/:sku_id/stock", h.UpdateStock) // 更新库存。
		group.GET("/:sku_id/logs", h.GetInventoryLogs) // 获取库存日志。
	}
}
