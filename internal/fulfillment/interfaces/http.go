// Package interfaces 履约服务接口层
package interfaces

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/application"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
)

// HTTPHandler HTTP 接口处理器
type HTTPHandler struct {
	commandService *application.CommandService
	queryService   *application.QueryService
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(
	commandService *application.CommandService,
	queryService *application.QueryService,
) *HTTPHandler {
	return &HTTPHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

// RegisterRoutes 注册路由
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	fulfillments := r.Group("/fulfillments")
	{
		fulfillments.POST("", h.Create)
		fulfillments.GET("/:id", h.Get)
		fulfillments.GET("", h.List)
		fulfillments.POST("/:id/assign-picking", h.AssignPicking)
		fulfillments.POST("/:id/start-picking", h.StartPicking)
		fulfillments.POST("/:id/complete-picking", h.CompletePicking)
		fulfillments.POST("/:id/start-packing", h.StartPacking)
		fulfillments.POST("/:id/complete-packing", h.CompletePacking)
		fulfillments.POST("/:id/arrange-shipment", h.ArrangeShipment)
		fulfillments.POST("/:id/confirm-shipment", h.ConfirmShipment)
		fulfillments.POST("/:id/cancel", h.Cancel)
	}
}

// CreateFulfillmentRequest 创建履约单请求
type CreateFulfillmentRequest struct {
	OrderNo          string                           `json:"order_no" binding:"required"`
	MerchantID       uint64                           `json:"merchant_id" binding:"required"`
	StoreID          uint64                           `json:"store_id"`
	WarehouseID      uint64                           `json:"warehouse_id"`
	Type             int8                             `json:"type"`
	ExpectedShipTime *time.Time                       `json:"expected_ship_time"`
	Remark           string                           `json:"remark"`
	Address          application.ShippingAddress      `json:"address" binding:"required"`
	Items            []application.FulfillmentItemDTO `json:"items" binding:"required"`
}

// Create 创建履约单
func (h *HTTPHandler) Create(c *gin.Context) {
	var req CreateFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CreateFulfillmentCommand{
		OrderNo:          req.OrderNo,
		MerchantID:       req.MerchantID,
		StoreID:          req.StoreID,
		WarehouseID:      req.WarehouseID,
		Type:             domain.FulfillmentType(req.Type),
		ExpectedShipTime: req.ExpectedShipTime,
		Remark:           req.Remark,
		Address:          req.Address,
		Items:            req.Items,
	}

	result, err := h.commandService.CreateFulfillment(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// AssignPicking 分配拣货
func (h *HTTPHandler) AssignPicking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		PickerID   uint64 `json:"picker_id" binding:"required"`
		PickerName string `json:"picker_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.AssignPickingCommand{
		FulfillmentID: uint(id),
		PickerID:      req.PickerID,
		PickerName:    req.PickerName,
	}

	if err := h.commandService.AssignPicking(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "picker assigned"})
}

// StartPicking 开始拣货
func (h *HTTPHandler) StartPicking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		PickerID uint64 `json:"picker_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.StartPickingCommand{
		FulfillmentID: uint(id),
		PickerID:      req.PickerID,
	}

	if err := h.commandService.StartPicking(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "picking started"})
}

// CompletePicking 完成拣货
func (h *HTTPHandler) CompletePicking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Items map[string]int32 `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CompletePickingCommand{
		FulfillmentID: uint(id),
		Items:         req.Items,
	}

	if err := h.commandService.CompletePicking(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "picking completed"})
}

// StartPacking 开始打包
func (h *HTTPHandler) StartPacking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		PackerID   uint64 `json:"packer_id" binding:"required"`
		PackerName string `json:"packer_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.StartPackingCommand{
		FulfillmentID: uint(id),
		PackerID:      req.PackerID,
		PackerName:    req.PackerName,
	}

	if err := h.commandService.StartPacking(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "packing started"})
}

// CompletePacking 完成打包
func (h *HTTPHandler) CompletePacking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Packages []application.PackageDTO `json:"packages" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CompletePackingCommand{
		FulfillmentID: uint(id),
		Packages:      req.Packages,
	}

	if err := h.commandService.CompletePacking(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "packing completed"})
}

// ArrangeShipment 安排发货
func (h *HTTPHandler) ArrangeShipment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		CarrierCode string `json:"carrier_code" binding:"required"`
		CarrierName string `json:"carrier_name" binding:"required"`
		ShippingFee int64  `json:"shipping_fee"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ArrangeShipmentCommand{
		FulfillmentID: uint(id),
		CarrierCode:   req.CarrierCode,
		CarrierName:   req.CarrierName,
		ShippingFee:   req.ShippingFee,
	}

	if err := h.commandService.ArrangeShipment(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "shipment arranged"})
}

// ConfirmShipment 确认发货
func (h *HTTPHandler) ConfirmShipment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		TrackingNo string `json:"tracking_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ConfirmShipmentCommand{
		FulfillmentID: uint(id),
		TrackingNo:    req.TrackingNo,
	}

	if err := h.commandService.ConfirmShipment(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "shipment confirmed"})
}

// Cancel 取消履约
func (h *HTTPHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Reason   string `json:"reason" binding:"required"`
		Operator string `json:"operator" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CancelFulfillmentCommand{
		FulfillmentID: uint(id),
		Reason:        req.Reason,
		Operator:      req.Operator,
	}

	if err := h.commandService.CancelFulfillment(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}

// Get 获取履约单详情
func (h *HTTPHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fulfillment, err := h.queryService.GetFulfillment(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fulfillment)
}

// List 履约单列表
func (h *HTTPHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	merchantID, _ := strconv.ParseUint(c.Query("merchant_id"), 10, 64)
	warehouseID, _ := strconv.ParseUint(c.Query("warehouse_id"), 10, 64)

	filter := &domain.FulfillmentFilter{
		MerchantID:  merchantID,
		WarehouseID: warehouseID,
		OrderNo:     c.Query("order_no"),
		Page:        page,
		PageSize:    pageSize,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			s := domain.FulfillmentStatus(status)
			filter.Status = &s
		}
	}

	result, err := h.queryService.ListFulfillments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
