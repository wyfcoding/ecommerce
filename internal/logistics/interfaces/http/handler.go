package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/application"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/entity"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.LogisticsService
	logger  *slog.Logger
}

func NewHandler(service *application.LogisticsService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateLogistics 创建物流单
func (h *Handler) CreateLogistics(c *gin.Context) {
	var req struct {
		OrderID         uint64  `json:"order_id" binding:"required"`
		OrderNo         string  `json:"order_no" binding:"required"`
		TrackingNo      string  `json:"tracking_no" binding:"required"`
		Carrier         string  `json:"carrier" binding:"required"`
		CarrierCode     string  `json:"carrier_code"`
		SenderName      string  `json:"sender_name" binding:"required"`
		SenderPhone     string  `json:"sender_phone" binding:"required"`
		SenderAddress   string  `json:"sender_address" binding:"required"`
		SenderLat       float64 `json:"sender_lat" binding:"required"`
		SenderLon       float64 `json:"sender_lon" binding:"required"`
		ReceiverName    string  `json:"receiver_name" binding:"required"`
		ReceiverPhone   string  `json:"receiver_phone" binding:"required"`
		ReceiverAddress string  `json:"receiver_address" binding:"required"`
		ReceiverLat     float64 `json:"receiver_lat" binding:"required"`
		ReceiverLon     float64 `json:"receiver_lon" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	logistics, err := h.service.CreateLogistics(c.Request.Context(), req.OrderID, req.OrderNo, req.TrackingNo, req.Carrier, req.CarrierCode,
		req.SenderName, req.SenderPhone, req.SenderAddress, req.SenderLat, req.SenderLon,
		req.ReceiverName, req.ReceiverPhone, req.ReceiverAddress, req.ReceiverLat, req.ReceiverLon)
	if err != nil {
		h.logger.Error("Failed to create logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create logistics", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Logistics created successfully", logistics)
}

// GetLogistics 获取物流信息
func (h *Handler) GetLogistics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	logistics, err := h.service.GetLogistics(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get logistics", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Logistics retrieved successfully", logistics)
}

// UpdateStatus 更新物流状态
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Status      int    `json:"status" binding:"required"`
		Location    string `json:"location"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, entity.LogisticsStatus(req.Status), req.Location, req.Description); err != nil {
		h.logger.Error("Failed to update status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update status", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Status updated successfully", nil)
}

// AddTrace 添加物流轨迹
func (h *Handler) AddTrace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Location    string `json:"location" binding:"required"`
		Description string `json:"description" binding:"required"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AddTrace(c.Request.Context(), id, req.Location, req.Description, req.Status); err != nil {
		h.logger.Error("Failed to add trace", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add trace", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Trace added successfully", nil)
}

// SetEstimatedTime 设置预计送达时间
func (h *Handler) SetEstimatedTime(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		EstimatedTime time.Time `json:"estimated_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.SetEstimatedTime(c.Request.Context(), id, req.EstimatedTime); err != nil {
		h.logger.Error("Failed to set estimated time", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to set estimated time", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Estimated time set successfully", nil)
}

// ListLogistics 获取物流列表
func (h *Handler) ListLogistics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListLogistics(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list logistics", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Logistics listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/logistics")
	{
		group.POST("", h.CreateLogistics)
		group.GET("", h.ListLogistics)
		group.GET("/:id", h.GetLogistics)
		group.PUT("/:id/status", h.UpdateStatus)
		group.POST("/:id/traces", h.AddTrace)
		group.PUT("/:id/estimated_time", h.SetEstimatedTime)
	}
}
