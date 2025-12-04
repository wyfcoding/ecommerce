package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/loyalty/application"   // 导入忠诚度模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity" // 导入忠诚度模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                   // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Loyalty模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.LoyaltyService // 依赖Loyalty应用服务，处理核心业务逻辑。
	logger  *slog.Logger                // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Loyalty HTTP Handler 实例。
func NewHandler(service *application.LoyaltyService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetAccount 处理获取会员账户信息的HTTP请求。
// Method: GET
// Path: /loyalty/accounts/:user_id
func (h *Handler) GetAccount(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取或创建会员账户。
	account, err := h.service.GetOrCreateAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get account", err.Error())
		return
	}

	// 返回成功的响应，包含会员账户信息。
	response.SuccessWithStatus(c, http.StatusOK, "Account retrieved successfully", account)
}

// UpdatePoints 处理更新积分的HTTP请求（增加或扣减）。
// Method: POST
// Path: /loyalty/accounts/:user_id/points
func (h *Handler) UpdatePoints(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 定义请求体结构，用于接收积分更新信息。
	var req struct {
		Action      string `json:"action" binding:"required,oneof=add deduct"` // 操作类型（add或deduct），必填。
		Points      int64  `json:"points" binding:"required,gt=0"`             // 积分数量，必填且大于0。
		Type        string `json:"type" binding:"required"`                    // 交易类型，必填。
		Description string `json:"description"`                                // 描述，选填。
		OrderID     uint64 `json:"order_id"`                                   // 关联订单ID，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var opErr error // 操作结果错误。
	ctx := c.Request.Context()

	// 根据操作类型调用应用服务层的相应方法。
	if req.Action == "add" {
		opErr = h.service.AddPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	} else {
		opErr = h.service.DeductPoints(ctx, userID, req.Points, req.Type, req.Description, req.OrderID)
	}

	if opErr != nil {
		h.logger.Error("Failed to update points", "error", opErr)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to update points", opErr.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Points updated successfully", nil)
}

// GetTransactions 处理获取用户积分交易记录的HTTP请求。
// Method: GET
// Path: /loyalty/accounts/:user_id/transactions
func (h *Handler) GetTransactions(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取积分交易记录列表。
	list, total, err := h.service.GetPointsTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list transactions", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list transactions", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Transactions listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddBenefit 处理添加会员权益的HTTP请求。
// Method: POST
// Path: /loyalty/benefits
func (h *Handler) AddBenefit(c *gin.Context) {
	// 定义请求体结构，用于接收会员权益信息。
	var req struct {
		Level        string  `json:"level" binding:"required"` // 会员等级，必填。
		Name         string  `json:"name" binding:"required"`  // 权益名称，必填。
		Description  string  `json:"description"`              // 描述，选填。
		DiscountRate float64 `json:"discount_rate"`            // 折扣率，选填。
		PointsRate   float64 `json:"points_rate"`              // 积分倍率，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层添加会员权益。
	benefit, err := h.service.AddBenefit(c.Request.Context(), entity.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		h.logger.Error("Failed to add benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add benefit", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Benefit added successfully", benefit)
}

// ListBenefits 处理获取会员权益列表的HTTP请求。
// Method: GET
// Path: /loyalty/benefits
func (h *Handler) ListBenefits(c *gin.Context) {
	// 从查询参数中获取会员等级。
	level := c.Query("level")
	// 调用应用服务层获取会员权益列表。
	list, err := h.service.ListBenefits(c.Request.Context(), entity.MemberLevel(level))
	if err != nil {
		h.logger.Error("Failed to list benefits", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list benefits", err.Error())
		return
	}

	// 返回成功的响应，包含会员权益列表。
	response.SuccessWithStatus(c, http.StatusOK, "Benefits listed successfully", list)
}

// DeleteBenefit 处理删除会员权益的HTTP请求。
// Method: DELETE
// Path: /loyalty/benefits/:id
func (h *Handler) DeleteBenefit(c *gin.Context) {
	// 从URL路径中解析权益ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层删除会员权益。
	if err := h.service.DeleteBenefit(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete benefit", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete benefit", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Benefit deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Loyalty模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /loyalty 路由组，用于所有忠诚度相关接口。
	group := r.Group("/loyalty")
	{
		// 会员账户接口。
		group.GET("/accounts/:user_id", h.GetAccount)                   // 获取会员账户。
		group.POST("/accounts/:user_id/points", h.UpdatePoints)         // 更新积分（增减）。
		group.GET("/accounts/:user_id/transactions", h.GetTransactions) // 获取积分交易记录。
		// TODO: 补充增加消费金额、冻结/解冻积分等接口。

		// 会员权益接口。
		group.POST("/benefits", h.AddBenefit)          // 添加会员权益。
		group.GET("/benefits", h.ListBenefits)         // 获取会员权益列表。
		group.DELETE("/benefits/:id", h.DeleteBenefit) // 删除会员权益。
		// TODO: 补充更新权益、获取权益详情等接口。
	}
}
