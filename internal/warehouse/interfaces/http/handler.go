package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/warehouse/application" // 导入仓库模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                   // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Warehouse模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.WarehouseService // 依赖Warehouse应用服务，处理核心业务逻辑。
	logger  *slog.Logger                  // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Warehouse HTTP Handler 实例。
func NewHandler(service *application.WarehouseService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateWarehouse 处理创建仓库的HTTP请求。
// Method: POST
// Path: /warehouse
func (h *Handler) CreateWarehouse(c *gin.Context) {
	// 定义请求体结构，用于接收仓库的创建信息。
	var req struct {
		Code string `json:"code" binding:"required"` // 仓库代码，必填。
		Name string `json:"name" binding:"required"` // 仓库名称，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建仓库。
	warehouse, err := h.service.CreateWarehouse(c.Request.Context(), req.Code, req.Name)
	if err != nil {
		h.logger.Error("Failed to create warehouse", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create warehouse", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Warehouse created successfully", warehouse)
}

// ListWarehouses 处理获取仓库列表的HTTP请求。
// Method: GET
// Path: /warehouse
func (h *Handler) ListWarehouses(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取仓库列表。
	list, total, err := h.service.ListWarehouses(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list warehouses", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list warehouses", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Warehouses listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateStock 处理更新库存的HTTP请求。
// Method: POST
// Path: /warehouse/stock
func (h *Handler) UpdateStock(c *gin.Context) {
	// 定义请求体结构，用于接收库存更新信息。
	var req struct {
		WarehouseID uint64 `json:"warehouse_id" binding:"required"` // 仓库ID，必填。
		SkuID       uint64 `json:"sku_id" binding:"required"`       // SKU ID，必填。
		Quantity    int32  `json:"quantity" binding:"required"`     // 数量（正数增加，负数减少），必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层更新库存。
	if err := h.service.UpdateStock(c.Request.Context(), req.WarehouseID, req.SkuID, req.Quantity); err != nil {
		h.logger.Error("Failed to update stock", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Stock updated successfully", nil)
}

// CreateTransfer 处理创建库存调拨单的HTTP请求。
// Method: POST
// Path: /warehouse/transfer
func (h *Handler) CreateTransfer(c *gin.Context) {
	// 定义请求体结构，用于接收调拨单的创建信息。
	var req struct {
		FromID    uint64 `json:"from_id" binding:"required"`    // 调出仓库ID，必填。
		ToID      uint64 `json:"to_id" binding:"required"`      // 调入仓库ID，必填。
		SkuID     uint64 `json:"sku_id" binding:"required"`     // SKU ID，必填。
		Quantity  int32  `json:"quantity" binding:"required"`   // 调拨数量，必填。
		CreatedBy uint64 `json:"created_by" binding:"required"` // 创建人ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建调拨单。
	transfer, err := h.service.CreateTransfer(c.Request.Context(), req.FromID, req.ToID, req.SkuID, req.Quantity, req.CreatedBy)
	if err != nil {
		h.logger.Error("Failed to create transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create transfer", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Transfer created successfully", transfer)
}

// CompleteTransfer 处理完成库存调拨的HTTP请求。
// Method: POST
// Path: /warehouse/transfer/:id/complete
func (h *Handler) CompleteTransfer(c *gin.Context) {
	// 从URL路径中解析调拨单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层完成调拨单。
	if err := h.service.CompleteTransfer(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to complete transfer", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete transfer", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Transfer completed successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Warehouse模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /warehouse 路由组，用于所有仓库相关接口。
	group := r.Group("/warehouse")
	{
		group.POST("", h.CreateWarehouse)                        // 创建仓库。
		group.GET("", h.ListWarehouses)                          // 获取仓库列表。
		group.POST("/stock", h.UpdateStock)                      // 更新库存。
		group.POST("/transfer", h.CreateTransfer)                // 创建调拨单。
		group.POST("/transfer/:id/complete", h.CompleteTransfer) // 完成调拨单。
		// TODO: 补充获取仓库详情、获取库存详情、获取调拨单详情、列出调拨单等接口。
	}
}
