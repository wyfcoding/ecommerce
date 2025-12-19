package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/aftersales/application"       // 导入售后模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"     // 导入售后模块的领域实体。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository" // 导入售后模块的领域仓储查询对象。
	"github.com/wyfcoding/pkg/response"                                    // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了AfterSales模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.AfterSalesService // 依赖AfterSales应用服务，处理核心业务逻辑。
	logger  *slog.Logger                   // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 AfterSales HTTP Handler 实例。
func NewHandler(service *application.AfterSalesService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Create 处理创建售后申请的HTTP请求。
// Method: POST
// Path: /aftersales
func (h *Handler) Create(c *gin.Context) {
	// 定义请求体结构，用于接收售后申请的详细信息。
	var req struct {
		OrderID     uint64                   `json:"order_id" binding:"required"` // 订单ID，必填。
		OrderNo     string                   `json:"order_no" binding:"required"` // 订单号，必填。
		UserID      uint64                   `json:"user_id" binding:"required"`  // 用户ID，必填。
		Type        entity.AfterSalesType    `json:"type" binding:"required"`     // 售后类型，必填。
		Reason      string                   `json:"reason" binding:"required"`   // 申请原因，必填。
		Description string                   `json:"description"`                 // 详细描述，选填。
		Images      []string                 `json:"images"`                      // 凭证图片URL列表，选填。
		Items       []*entity.AfterSalesItem `json:"items" binding:"required"`    // 申请售后的商品列表，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建售后申请。
	afterSales, err := h.service.CreateAfterSales(c.Request.Context(), req.OrderID, req.OrderNo, req.UserID, req.Type, req.Reason, req.Description, req.Images, req.Items)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create after-sales", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "After-sales created successfully", afterSales)
}

// Approve 处理批准售后申请的HTTP请求。
// Method: POST
// Path: /aftersales/:id/approve
func (h *Handler) Approve(c *gin.Context) {
	// 从URL路径中解析售后申请ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收操作人员和批准金额。
	var req struct {
		Operator string `json:"operator" binding:"required"` // 操作人员，必填。
		Amount   int64  `json:"amount" binding:"required"`   // 批准金额（单位：分），必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层批准售后申请。
	if err := h.service.Approve(c.Request.Context(), id, req.Operator, req.Amount); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to approve after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to approve after-sales", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "After-sales approved successfully", nil)
}

// Reject 处理拒绝售后申请的HTTP请求。
// Method: POST
// Path: /aftersales/:id/reject
func (h *Handler) Reject(c *gin.Context) {
	// 从URL路径中解析售后申请ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收操作人员和拒绝原因。
	var req struct {
		Operator string `json:"operator" binding:"required"` // 操作人员，必填。
		Reason   string `json:"reason" binding:"required"`   // 拒绝原因，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层拒绝售后申请。
	if err := h.service.Reject(c.Request.Context(), id, req.Operator, req.Reason); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to reject after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to reject after-sales", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "After-sales rejected successfully", nil)
}

// List 处理列出售后申请的HTTP请求。
// Method: GET
// Path: /aftersales
// 支持分页和基于用户ID的过滤。
func (h *Handler) List(c *gin.Context) {
	// 从查询参数中获取页码、每页大小和用户ID，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	// 构建查询对象。
	query := &repository.AfterSalesQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}

	// 调用应用服务层获取售后申请列表。
	list, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list after-sales", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list after-sales", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "After-sales listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetails 处理获取售后申请详情的HTTP请求。
// Method: GET
// Path: /aftersales/:id
func (h *Handler) GetDetails(c *gin.Context) {
	// 从URL路径中解析售后申请ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层获取售后申请详情。
	details, err := h.service.GetDetails(c.Request.Context(), id)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get after-sales details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get after-sales details", err.Error())
		return
	}

	// 返回成功的响应，包含售后申请详情。
	response.SuccessWithStatus(c, http.StatusOK, "After-sales details retrieved successfully", details)
}

// RegisterRoutes 在给定的Gin路由组中注册AfterSales模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /aftersales 路由组，用于所有售后相关接口。
	group := r.Group("/aftersales")
	{
		group.POST("", h.Create)              // 创建售后申请。
		group.GET("", h.List)                 // 获取售后申请列表。
		group.GET("/:id", h.GetDetails)       // 获取售后申请详情。
		group.POST("/:id/approve", h.Approve) // 批准售后申请。
		group.POST("/:id/reject", h.Reject)   // 拒绝售后申请。
	}
}
