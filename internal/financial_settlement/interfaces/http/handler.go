package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/financial_settlement/application" // 导入财务结算模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                              // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了FinancialSettlement模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.FinancialSettlementService // 依赖FinancialSettlement应用服务，处理核心业务逻辑。
	logger  *slog.Logger                            // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 FinancialSettlement HTTP Handler 实例。
func NewHandler(service *application.FinancialSettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSettlement 处理创建结算单的HTTP请求。
// Method: POST
// Path: /settlements
func (h *Handler) CreateSettlement(c *gin.Context) {
	// 定义请求体结构，用于接收结算单的创建信息。
	var req struct {
		SellerID  uint64    `json:"seller_id" binding:"required"`  // 卖家ID，必填。
		Period    string    `json:"period" binding:"required"`     // 结算周期，必填。
		StartDate time.Time `json:"start_date" binding:"required"` // 开始日期，必填。
		EndDate   time.Time `json:"end_date" binding:"required"`   // 结束日期，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建结算单。
	settlement, err := h.service.CreateSettlement(c.Request.Context(), req.SellerID, req.Period, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.Error("Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// ApproveSettlement 处理审核批准结算单的HTTP请求。
// Method: POST
// Path: /settlements/:id/approve
func (h *Handler) ApproveSettlement(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收批准人信息。
	var req struct {
		ApprovedBy string `json:"approved_by" binding:"required"` // 批准人，必填。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层批准结算单。
	if err := h.service.ApproveSettlement(c.Request.Context(), id, req.ApprovedBy); err != nil {
		h.logger.Error("Failed to approve settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve settlement", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Settlement approved successfully", nil)
}

// RejectSettlement 处理审核拒绝结算单的HTTP请求。
// Method: POST
// Path: /settlements/:id/reject
func (h *Handler) RejectSettlement(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收拒绝原因。
	var req struct {
		Reason string `json:"reason" binding:"required"` // 拒绝原因，必填。
	}
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层拒绝结算单。
	if err := h.service.RejectSettlement(c.Request.Context(), id, req.Reason); err != nil {
		h.logger.Error("Failed to reject settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject settlement", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Settlement rejected successfully", nil)
}

// GetSettlement 处理获取结算单详情的HTTP请求。
// Method: GET
// Path: /settlements/:id
func (h *Handler) GetSettlement(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取结算单详情。
	settlement, err := h.service.GetSettlement(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get settlement", err.Error())
		return
	}

	// 返回成功的响应，包含结算单详情。
	response.SuccessWithStatus(c, http.StatusOK, "Settlement retrieved successfully", settlement)
}

// ListSettlements 处理获取结算单列表的HTTP请求。
// Method: GET
// Path: /settlements
func (h *Handler) ListSettlements(c *gin.Context) {
	// 从查询参数中获取卖家ID、页码和每页大小，并设置默认值。
	sellerID, _ := strconv.ParseUint(c.Query("seller_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取结算单列表。
	list, total, err := h.service.ListSettlements(c.Request.Context(), sellerID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list settlements", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list settlements", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Settlements listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ProcessPayment 处理结算单支付的HTTP请求。
// Method: POST
// Path: /settlements/:id/pay
func (h *Handler) ProcessPayment(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层处理支付。
	payment, err := h.service.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to process payment", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process payment", err.Error())
		return
	}

	// 返回成功的响应，包含支付记录。
	response.SuccessWithStatus(c, http.StatusOK, "Payment processed successfully", payment)
}

// RegisterRoutes 在给定的Gin路由组中注册FinancialSettlement模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /settlements 路由组，用于所有结算相关接口。
	group := r.Group("/settlements")
	{
		group.POST("", h.CreateSettlement)              // 创建结算单。
		group.GET("/:id", h.GetSettlement)              // 获取结算单详情。
		group.GET("", h.ListSettlements)                // 获取结算单列表。
		group.POST("/:id/approve", h.ApproveSettlement) // 批准结算单。
		group.POST("/:id/reject", h.RejectSettlement)   // 拒绝结算单。
		group.POST("/:id/pay", h.ProcessPayment)        // 处理结算支付。
		// TODO: 补充更新结算单、删除结算单、获取结算单订单明细等接口。
	}
}
