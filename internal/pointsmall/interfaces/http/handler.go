package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"   // 导入积分商城模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity" // 导入积分商城模块的领域实体。
	"github.com/wyfcoding/ecommerce/pkg/response"                      // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Pointsmall模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.PointsService // 依赖Points应用服务，处理核心业务逻辑。
	logger  *slog.Logger               // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Pointsmall HTTP Handler 实例。
func NewHandler(service *application.PointsService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateProduct 处理创建积分商品的HTTP请求。
// Method: POST
// Path: /pointsmall/products
func (h *Handler) CreateProduct(c *gin.Context) {
	// 定义请求体结构，使用 entity.PointsProduct 结构体直接绑定。
	var req entity.PointsProduct
	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建商品。
	if err := h.service.CreateProduct(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Product created successfully", req)
}

// ListProducts 处理获取积分商品列表的HTTP请求。
// Method: GET
// Path: /pointsmall/products
func (h *Handler) ListProducts(c *gin.Context) {
	// 从查询参数中获取状态字符串，并尝试转换为 int 类型。
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

	// 调用应用服务层获取商品列表。
	list, total, err := h.service.ListProducts(c.Request.Context(), status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list products", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list products", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Products listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ExchangeProduct 处理兑换积分商品的HTTP请求。
// Method: POST
// Path: /pointsmall/exchange
func (h *Handler) ExchangeProduct(c *gin.Context) {
	// 定义请求体结构，用于接收兑换商品的详细信息。
	var req struct {
		UserID    uint64 `json:"user_id" binding:"required"`    // 用户ID，必填。
		ProductID uint64 `json:"product_id" binding:"required"` // 商品ID，必填。
		Quantity  int32  `json:"quantity" binding:"required"`   // 兑换数量，必填。
		Address   string `json:"address" binding:"required"`    // 收货地址，必填。
		Phone     string `json:"phone" binding:"required"`      // 联系电话，必填。
		Receiver  string `json:"receiver" binding:"required"`   // 收货人姓名，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层兑换商品。
	order, err := h.service.ExchangeProduct(c.Request.Context(), req.UserID, req.ProductID, req.Quantity, req.Address, req.Phone, req.Receiver)
	if err != nil {
		h.logger.Error("Failed to exchange product", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange product", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created，包含兑换订单信息。
	response.SuccessWithStatus(c, http.StatusCreated, "Product exchanged successfully", order)
}

// GetAccount 处理获取用户积分账户信息的HTTP请求。
// Method: GET
// Path: /pointsmall/account
// 注意：这里设计为查询参数 user_id，而不是路径参数，方便匿名访问（如果用户ID从token获取）。
func (h *Handler) GetAccount(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取用户积分账户。
	account, err := h.service.GetAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get account", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get account", err.Error())
		return
	}

	// 返回成功的响应，包含积分账户信息。
	response.SuccessWithStatus(c, http.StatusOK, "Account retrieved successfully", account)
}

// AddPoints 处理增加用户积分的HTTP请求（通常由管理员或系统触发）。
// Method: POST
// Path: /pointsmall/points
func (h *Handler) AddPoints(c *gin.Context) {
	// 定义请求体结构，用于接收增加积分的详细信息。
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
		Points      int64  `json:"points" binding:"required"`  // 增加的积分数量，必填。
		Description string `json:"description"`                // 描述，选填。
		RefID       string `json:"ref_id"`                     // 关联ID，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层增加积分。
	if err := h.service.AddPoints(c.Request.Context(), req.UserID, req.Points, req.Description, req.RefID); err != nil {
		h.logger.Error("Failed to add points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add points", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Points added successfully", nil)
}

// ListOrders 处理获取积分订单列表的HTTP请求。
// Method: GET
// Path: /pointsmall/orders
func (h *Handler) ListOrders(c *gin.Context) {
	// 从查询参数中获取用户ID。
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取状态字符串，并尝试转换为 int 类型。
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

	// 调用应用服务层获取积分订单列表。
	list, total, err := h.service.ListOrders(c.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Orders listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Pointsmall模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /pointsmall 路由组，用于所有积分商城相关接口。
	group := r.Group("/pointsmall")
	{
		// 积分商品接口。
		group.POST("/products", h.CreateProduct) // 创建积分商品。
		group.GET("/products", h.ListProducts)   // 获取积分商品列表。
		// TODO: 补充更新商品、删除商品、上架/下架商品等接口。

		// 积分兑换和订单接口。
		group.POST("/exchange", h.ExchangeProduct) // 兑换商品。
		group.GET("/orders", h.ListOrders)         // 获取积分订单列表。
		// TODO: 补充获取订单详情、更新订单状态等接口。

		// 积分账户接口。
		group.GET("/account", h.GetAccount) // 获取用户积分账户信息。
		group.POST("/points", h.AddPoints)  // 增加用户积分。
		// TODO: 补充积分流水列表、扣减积分等接口。
	}
}
