package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/settlement/application" // 导入结算模块的应用服务。
	"github.com/wyfcoding/pkg/response"                              // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Settlement模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.SettlementService // 依赖Settlement应用服务，处理核心业务逻辑。
	logger  *slog.Logger                   // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Settlement HTTP Handler 实例。
func NewHandler(service *application.SettlementService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSettlement 处理创建结算单的HTTP请求。
// Method: POST
// Path: /settlement
func (h *Handler) CreateSettlement(c *gin.Context) {
	// 定义请求体结构，用于接收结算单的创建信息。
	var req struct {
		MerchantID uint64 `json:"merchant_id" binding:"required"` // 商户ID，必填。
		Cycle      string `json:"cycle" binding:"required"`       // 结算周期，必填。
		StartDate  string `json:"start_date" binding:"required"`  // 开始日期（格式：YYYY-MM-DD），必填。
		EndDate    string `json:"end_date" binding:"required"`    // 结束日期（格式：YYYY-MM-DD），必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 解析开始日期和结束日期字符串为time.Time类型。
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid StartDate format", err.Error())
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid EndDate format", err.Error())
		return
	}

	// 调用应用服务层创建结算单。
	settlement, err := h.service.CreateSettlement(c.Request.Context(), req.MerchantID, req.Cycle, start, end)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create settlement", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Settlement created successfully", settlement)
}

// AddOrder 处理添加订单到结算单的HTTP请求。
// Method: POST
// Path: /settlement/:id/orders
func (h *Handler) AddOrder(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收订单详情。
	var req struct {
		OrderID uint64 `json:"order_id" binding:"required"` // 订单ID，必填。
		OrderNo string `json:"order_no" binding:"required"` // 订单号，必填。
		Amount  uint64 `json:"amount" binding:"required"`   // 订单金额，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加订单到结算单。
	if err := h.service.AddOrderToSettlement(c.Request.Context(), id, req.OrderID, req.OrderNo, req.Amount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add order", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add order", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Order added successfully", nil)
}

// Process 处理结算单的HTTP请求。
// Method: POST
// Path: /settlement/:id/process
func (h *Handler) Process(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层处理结算单。
	if err := h.service.ProcessSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to process settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to process settlement", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Settlement processing started", nil)
}

// Complete 处理完成结算单的HTTP请求。
// Method: POST
// Path: /settlement/:id/complete
func (h *Handler) Complete(c *gin.Context) {
	// 从URL路径中解析结算单ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层完成结算单。
	if err := h.service.CompleteSettlement(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to complete settlement", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to complete settlement", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Settlement completed successfully", nil)
}

// List 处理获取结算单列表的HTTP请求。
// Method: GET
// Path: /settlement
func (h *Handler) List(c *gin.Context) {
	// 从查询参数中获取商户ID、状态、页码和每页大小，并设置默认值。
	merchantID, _ := strconv.ParseUint(c.Query("merchant_id"), 10, 64)
	statusStr := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var status *int
	if statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil { // 只有当状态字符串能成功转换为int时才设置过滤状态。
			status = &s
		}
	}

	// 调用应用服务层获取结算单列表。
	list, total, err := h.service.ListSettlements(c.Request.Context(), merchantID, status, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list settlements", "error", err)
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

// GetAccount 处理获取商户账户信息的HTTP请求。
// Method: GET
// Path: /settlement/accounts/:merchant_id
func (h *Handler) GetAccount(c *gin.Context) {
	// 从URL路径中解析商户ID。
	merchantID, err := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid Merchant ID", err.Error())
		return
	}

	// 调用应用服务层获取商户账户。
	account, err := h.service.GetMerchantAccount(c.Request.Context(), merchantID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get merchant account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get merchant account", err.Error())
		return
	}

	// 返回成功的响应，包含商户账户信息。
	response.SuccessWithStatus(c, http.StatusOK, "Merchant account retrieved successfully", account)
}

// RegisterRoutes 在给定的Gin路由组中注册Settlement模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /settlement 路由组，用于所有结算相关接口。
	group := r.Group("/settlement")
	{
		group.POST("", h.CreateSettlement)                // 创建结算单。
		group.POST("/:id/orders", h.AddOrder)             // 添加订单到结算单。
		group.POST("/:id/process", h.Process)             // 处理结算单。
		group.POST("/:id/complete", h.Complete)           // 完成结算单。
		group.GET("", h.List)                             // 获取结算单列表。
		group.GET("/accounts/:merchant_id", h.GetAccount) // 获取商户账户信息。
		// TODO: 补充获取结算单详情、获取结算明细列表等接口。
	}
}
