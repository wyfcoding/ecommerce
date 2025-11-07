package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/warehouse/service"
	"ecommerce/pkg/response"
)

// WarehouseHandler 仓库HTTP处理器
type WarehouseHandler struct {
	service service.WarehouseService
	logger  *zap.Logger
}

// NewWarehouseHandler 创建仓库HTTP处理器
func NewWarehouseHandler(service service.WarehouseService, logger *zap.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *WarehouseHandler) RegisterRoutes(r *gin.RouterGroup) {
	warehouse := r.Group("/warehouses")
	{
		// 仓库管理
		warehouse.GET("", h.ListWarehouses)
		warehouse.GET("/:id", h.GetWarehouse)
		warehouse.GET("/:id/stocks", h.ListWarehouseStocks)
		
		// 库存调拨
		warehouse.POST("/transfers", h.CreateStockTransfer)
		warehouse.GET("/transfers/:no", h.GetStockTransfer)
		warehouse.POST("/transfers/:id/approve", h.ApproveStockTransfer)
		warehouse.POST("/transfers/:id/ship", h.ShipStockTransfer)
		warehouse.POST("/transfers/:id/receive", h.ReceiveStockTransfer)
		warehouse.POST("/transfers/:id/cancel", h.CancelStockTransfer)
		
		// 库存盘点
		warehouse.POST("/stocktakings", h.CreateStocktaking)
		warehouse.GET("/stocktakings/:no", h.GetStocktaking)
		warehouse.POST("/stocktakings/:id/start", h.StartStocktaking)
		warehouse.POST("/stocktakings/:id/items", h.RecordStocktakingItem)
		warehouse.POST("/stocktakings/:id/complete", h.CompleteStocktaking)
	}
}

// ListWarehouses 获取仓库列表
func (h *WarehouseHandler) ListWarehouses(c *gin.Context) {
	status := c.Query("status")
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	warehouses, total, err := h.service.ListWarehouses(c.Request.Context(), status, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取仓库列表失败", err)
		return
	}

	response.SuccessWithPagination(c, warehouses, total, int32(pageNum), int32(pageSize))
}

// GetWarehouse 获取仓库详情
func (h *WarehouseHandler) GetWarehouse(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	warehouse, err := h.service.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "仓库不存在", err)
		return
	}

	response.Success(c, warehouse)
}

// ListWarehouseStocks 获取仓库库存列表
func (h *WarehouseHandler) ListWarehouseStocks(c *gin.Context) {
	warehouseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	stocks, total, err := h.service.ListWarehouseStocks(c.Request.Context(), warehouseID, int32(pageSize), int32(pageNum))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取库存列表失败", err)
		return
	}

	response.SuccessWithPagination(c, stocks, total, int32(pageNum), int32(pageSize))
}

// CreateStockTransfer 创建库存调拨
func (h *WarehouseHandler) CreateStockTransfer(c *gin.Context) {
	var req struct {
		FromWarehouseID uint64 `json:"fromWarehouseId" binding:"required"`
		ToWarehouseID   uint64 `json:"toWarehouseId" binding:"required"`
		SkuID           uint64 `json:"skuId" binding:"required"`
		Quantity        int32  `json:"quantity" binding:"required"`
		Reason          string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	transfer, err := h.service.CreateStockTransfer(c.Request.Context(), req.FromWarehouseID, req.ToWarehouseID, req.SkuID, req.Quantity, req.Reason, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建调拨失败", err)
		return
	}

	response.Success(c, transfer)
}

// GetStockTransfer 获取调拨单详情
func (h *WarehouseHandler) GetStockTransfer(c *gin.Context) {
	transferNo := c.Param("no")

	transfer, err := h.service.GetStockTransfer(c.Request.Context(), transferNo)
	if err != nil {
		response.Error(c, http.StatusNotFound, "调拨单不存在", err)
		return
	}

	response.Success(c, transfer)
}

// ApproveStockTransfer 审核调拨单
func (h *WarehouseHandler) ApproveStockTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	approverID := c.GetUint64("userID")

	if err := h.service.ApproveStockTransfer(c.Request.Context(), id, approverID); err != nil {
		response.Error(c, http.StatusInternalServerError, "审核调拨失败", err)
		return
	}

	response.Success(c, nil)
}

// ShipStockTransfer 调拨发货
func (h *WarehouseHandler) ShipStockTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.ShipStockTransfer(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "发货失败", err)
		return
	}

	response.Success(c, nil)
}

// ReceiveStockTransfer 调拨收货
func (h *WarehouseHandler) ReceiveStockTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.ReceiveStockTransfer(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "收货失败", err)
		return
	}

	response.Success(c, nil)
}

// CancelStockTransfer 取消调拨
func (h *WarehouseHandler) CancelStockTransfer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.CancelStockTransfer(c.Request.Context(), id, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "取消调拨失败", err)
		return
	}

	response.Success(c, nil)
}

// CreateStocktaking 创建库存盘点
func (h *WarehouseHandler) CreateStocktaking(c *gin.Context) {
	var req struct {
		WarehouseID uint64 `json:"warehouseId" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Remark      string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	userID := c.GetUint64("userID")

	stocktaking, err := h.service.CreateStocktaking(c.Request.Context(), req.WarehouseID, req.Type, req.Remark, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建盘点失败", err)
		return
	}

	response.Success(c, stocktaking)
}

// GetStocktaking 获取盘点单详情
func (h *WarehouseHandler) GetStocktaking(c *gin.Context) {
	stockNo := c.Param("no")

	stocktaking, err := h.service.GetStocktaking(c.Request.Context(), stockNo)
	if err != nil {
		response.Error(c, http.StatusNotFound, "盘点单不存在", err)
		return
	}

	response.Success(c, stocktaking)
}

// StartStocktaking 开始盘点
func (h *WarehouseHandler) StartStocktaking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.StartStocktaking(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "开始盘点失败", err)
		return
	}

	response.Success(c, nil)
}

// RecordStocktakingItem 记录盘点明细
func (h *WarehouseHandler) RecordStocktakingItem(c *gin.Context) {
	stocktakingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	var req struct {
		SkuID       uint64 `json:"skuId" binding:"required"`
		BookStock   int32  `json:"bookStock" binding:"required"`
		ActualStock int32  `json:"actualStock" binding:"required"`
		Remark      string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.RecordStocktakingItem(c.Request.Context(), stocktakingID, req.SkuID, req.BookStock, req.ActualStock, req.Remark); err != nil {
		response.Error(c, http.StatusInternalServerError, "记录盘点明细失败", err)
		return
	}

	response.Success(c, nil)
}

// CompleteStocktaking 完成盘点
func (h *WarehouseHandler) CompleteStocktaking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.CompleteStocktaking(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "完成盘点失败", err)
		return
	}

	response.Success(c, nil)
}
