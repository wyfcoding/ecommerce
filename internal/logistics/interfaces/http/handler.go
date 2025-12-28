package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/logistics/application" // 导入物流模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/logistics/domain"      // 导入物流模块的领域实体。
	"github.com/wyfcoding/pkg/response"                             // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了Logistics模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	app    *application.Logistics // 依赖Logistics应用服务，处理核心业务逻辑。
	logger *slog.Logger           // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Logistics HTTP Handler 实例。
func NewHandler(app *application.Logistics, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// CreateLogistics 处理创建物流单的HTTP请求。
// HTTP 方法: POST
// 请求路径: /logistics
func (h *Handler) CreateLogistics(c *gin.Context) {
	// 定义请求体结构，用于接收物流单的创建信息。
	var req struct {
		OrderID         uint64  `json:"order_id" binding:"required"`         // 订单ID，必填。
		OrderNo         string  `json:"order_no" binding:"required"`         // 订单号，必填。
		TrackingNo      string  `json:"tracking_no" binding:"required"`      // 运单号，必填。
		Carrier         string  `json:"carrier" binding:"required"`          // 承运商，必填。
		CarrierCode     string  `json:"carrier_code"`                        // 承运商编码，选填。
		SenderName      string  `json:"sender_name" binding:"required"`      // 发件人姓名，必填。
		SenderPhone     string  `json:"sender_phone" binding:"required"`     // 发件人电话，必填。
		SenderAddress   string  `json:"sender_address" binding:"required"`   // 发件人地址，必填。
		SenderLat       float64 `json:"sender_lat" binding:"required"`       // 发件人纬度，必填。
		SenderLon       float64 `json:"sender_lon" binding:"required"`       // 发件人经度，必填。
		ReceiverName    string  `json:"receiver_name" binding:"required"`    // 收件人姓名，必填。
		ReceiverPhone   string  `json:"receiver_phone" binding:"required"`   // 收件人电话，必填。
		ReceiverAddress string  `json:"receiver_address" binding:"required"` // 收件人地址，必填。
		ReceiverLat     float64 `json:"receiver_lat" binding:"required"`     // 收件人纬度，必填。
		ReceiverLon     float64 `json:"receiver_lon" binding:"required"`     // 收件人经度，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建物流单。
	logistics, err := h.app.CreateLogistics(c.Request.Context(), req.OrderID, req.OrderNo, req.TrackingNo, req.Carrier, req.CarrierCode,
		req.SenderName, req.SenderPhone, req.SenderAddress, req.SenderLat, req.SenderLon,
		req.ReceiverName, req.ReceiverPhone, req.ReceiverAddress, req.ReceiverLat, req.ReceiverLon)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create logistics", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Logistics created successfully", logistics)
}

// GetLogistics 处理获取物流单详情的HTTP请求。
// HTTP 方法: GET
// 请求路径: /logistics/:id
func (h *Handler) GetLogistics(c *gin.Context) {
	// 从URL路径中解析物流单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取物流单详情。
	logistics, err := h.app.GetLogistics(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get logistics", err.Error())
		return
	}

	// 返回成功的响应，包含物流单详情。
	response.SuccessWithStatus(c, http.StatusOK, "Logistics retrieved successfully", logistics)
}

// UpdateStatus 处理更新物流状态的HTTP请求。
// HTTP 方法: PUT
// 请求路径: /logistics/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	// 从URL路径中解析物流单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收新的状态、位置和描述。
	var req struct {
		Status      int    `json:"status" binding:"required"` // 新的状态，必填。
		Location    string `json:"location"`                  // 位置，选填。
		Description string `json:"description"`               // 描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层更新物流状态。
	if err := h.app.UpdateStatus(c.Request.Context(), id, domain.LogisticsStatus(req.Status), req.Location, req.Description); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to update status", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update status", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Status updated successfully", nil)
}

// AddTrace 处理添加物流轨迹记录的HTTP请求。
// HTTP 方法: POST
// 请求路径: /logistics/:id/traces
func (h *Handler) AddTrace(c *gin.Context) {
	// 从URL路径中解析物流单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收轨迹信息。
	var req struct {
		Location    string `json:"location" binding:"required"`    // 位置，必填。
		Description string `json:"description" binding:"required"` // 描述，必填。
		Status      string `json:"status"`                         // 状态描述，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加物流轨迹。
	if err := h.app.AddTrace(c.Request.Context(), id, req.Location, req.Description, req.Status); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add trace", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add trace", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Trace added successfully", nil)
}

// SetEstimatedTime 处理设置预计送达时间的HTTP请求。
// HTTP 方法: PUT
// 请求路径: /logistics/:id/estimated_time
func (h *Handler) SetEstimatedTime(c *gin.Context) {
	// 从URL路径中解析物流单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收预计送达时间。
	var req struct {
		EstimatedTime time.Time `json:"estimated_time" binding:"required"` // 预计送达时间，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层设置预计送达时间。
	if err := h.app.SetEstimatedTime(c.Request.Context(), id, req.EstimatedTime); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to set estimated time", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to set estimated time", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Estimated time set successfully", nil)
}

// ListLogistics 处理获取物流单列表的HTTP请求。
// HTTP 方法: GET
// 请求路径: /logistics
func (h *Handler) ListLogistics(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取物流单列表。
	list, total, err := h.app.ListLogistics(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list logistics", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list logistics", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Logistics listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Logistics模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /logistics 路由组，用于所有物流相关接口。
	group := r.Group("/logistics")
	{
		group.POST("", h.CreateLogistics)                    // 创建物流单。
		group.GET("", h.ListLogistics)                       // 获取物流单列表。
		group.GET("/:id", h.GetLogistics)                    // 获取物流单详情。
		group.PUT("/:id/status", h.UpdateStatus)             // 更新物流状态。
		group.POST("/:id/traces", h.AddTrace)                // 添加物流轨迹。
		group.PUT("/:id/estimated_time", h.SetEstimatedTime) // 设置预计送达时间。
		// TODO: 补充根据运单号查询、优化配送路线等接口。
	}
}
