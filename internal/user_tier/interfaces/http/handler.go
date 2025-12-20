package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。

	"github.com/wyfcoding/ecommerce/internal/user_tier/application" // 导入用户等级模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"
	"github.com/wyfcoding/pkg/response" // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了UserTier模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.UserTierService // 依赖UserTier应用服务，处理核心业务逻辑。
	logger  *slog.Logger                 // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 UserTier HTTP Handler 实例。
func NewHandler(service *application.UserTierService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetTier 处理获取用户等级信息的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id
func (h *Handler) GetTier(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取用户等级。
	tier, err := h.service.GetUserTier(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user tier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user tier", err.Error())
		return
	}

	// 返回成功的响应，包含用户等级信息。
	response.SuccessWithStatus(c, http.StatusOK, "User tier retrieved successfully", tier)
}

// GetPoints 处理获取用户积分余额的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id/points
func (h *Handler) GetPoints(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 调用应用服务层获取用户积分。
	points, err := h.service.GetPoints(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get points", err.Error())
		return
	}

	// 返回成功的响应，包含积分余额。
	response.SuccessWithStatus(c, http.StatusOK, "Points retrieved successfully", gin.H{"points": points})
}

// Exchange 处理兑换商品的HTTP请求。
// Method: POST
// Path: /user_tier/exchange
func (h *Handler) Exchange(c *gin.Context) {
	// 定义请求体结构，用于接收兑换商品的详细信息。
	var req struct {
		UserID     uint64 `json:"user_id" binding:"required"`     // 用户ID，必填。
		ExchangeID uint64 `json:"exchange_id" binding:"required"` // 兑换商品ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层执行兑换。
	if err := h.service.Exchange(c.Request.Context(), req.UserID, req.ExchangeID); err != nil {
		h.logger.Error("Failed to exchange", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Exchange successful", nil)
}

// ListExchanges 处理获取可兑换商品列表的HTTP请求。
// Method: GET
// Path: /user_tier/exchanges
func (h *Handler) ListExchanges(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取可兑换商品列表。
	list, total, err := h.service.ListExchanges(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list exchanges", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list exchanges", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Exchanges listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListPointsLogs 处理获取用户积分日志列表的HTTP请求。
// Method: GET
// Path: /user_tier/:user_id/points/logs
func (h *Handler) ListPointsLogs(c *gin.Context) {
	// 从URL路径中解析用户ID。
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取积分日志列表。
	list, total, err := h.service.ListPointsLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list points logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list points logs", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	c.JSON(http.StatusOK, gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateTierConfig 处理创建或更新等级配置的请求。
// Method: POST
// Path: /user_tier/configs
func (h *Handler) CreateTierConfig(c *gin.Context) {
	var req struct {
		Level    int     `json:"level"`
		Name     string  `json:"name"`
		MinScore int64   `json:"min_score"`
		Discount float64 `json:"discount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 假设 entity.TierLevel 是 int 兼容的
	if err := h.service.CreateTierConfig(c.Request.Context(), entity.TierLevel(req.Level), req.Name, req.MinScore, req.Discount); err != nil {
		h.logger.Error("Failed to create tier config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ListTierConfigs 处理列出所有等级配置的请求。
// Method: GET
// Path: /user_tier/configs
func (h *Handler) ListTierConfigs(c *gin.Context) {
	configs, err := h.service.ListTierConfigs(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list tier configs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// CreateExchange 处理创建新兑换商品的请求。
// Method: POST
// Path: /user_tier/exchanges
func (h *Handler) CreateExchange(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Points      int64  `json:"points"`
		Stock       int32  `json:"stock"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateExchange(c.Request.Context(), req.Name, req.Description, req.Points, req.Stock)
	if err != nil {
		h.logger.Error("Failed to create exchange item", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// ListExchangeRecords 处理列出用户兑换记录的请求。
// Method: GET
// Path: /user_tier/:user_id/exchanges/records
func (h *Handler) ListExchangeRecords(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	records, total, err := h.service.ListExchangeRecords(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list exchange records", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddScore 处理给用户增加积分的请求。
// Method: POST
// Path: /user_tier/:user_id/score/add
func (h *Handler) AddScore(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	var req struct {
		Score int64 `json:"score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddScore(c.Request.Context(), userID, req.Score); err != nil {
		h.logger.Error("Failed to add score", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// AddPoints 处理给用户增加积分的请求。
// Method: POST
// Path: /user_tier/:user_id/points/add
func (h *Handler) AddPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	var req struct {
		Points int64  `json:"points"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddPoints(c.Request.Context(), userID, req.Points, req.Reason); err != nil {
		h.logger.Error("Failed to add points", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeductPoints 处理从用户扣除积分的请求。
// Method: POST
// Path: /user_tier/:user_id/points/deduct
func (h *Handler) DeductPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	var req struct {
		Points int64  `json:"points"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.DeductPoints(c.Request.Context(), userID, req.Points, req.Reason); err != nil {
		h.logger.Error("Failed to deduct points", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// RegisterRoutes 在给定的Gin路由组中注册UserTier模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /user_tier 路由组，用于所有用户等级和积分相关接口。
	group := r.Group("/user_tier")
	{
		group.GET("/:user_id", h.GetTier)                               // 获取用户等级信息。
		group.GET("/:user_id/points", h.GetPoints)                      // 获取用户积分余额。
		group.GET("/:user_id/points/logs", h.ListPointsLogs)            // 获取用户积分日志列表。
		group.POST("/exchange", h.Exchange)                             // 兑换商品。
		group.GET("/exchanges", h.ListExchanges)                        // 获取可兑换商品列表。
		group.GET("/:user_id/exchanges/records", h.ListExchangeRecords) // 获取用户兑换记录。

		// 管理接口 (TODO: 应该有权限控制)
		group.POST("/configs", h.CreateTierConfig)            // 创建/更新等级配置。
		group.GET("/configs", h.ListTierConfigs)              // 获取等级配置列表。
		group.POST("/exchanges", h.CreateExchange)            // 创建兑换商品。
		group.POST("/:user_id/score/add", h.AddScore)         // 增加成长值。
		group.POST("/:user_id/points/add", h.AddPoints)       // 增加积分。
		group.POST("/:user_id/points/deduct", h.DeductPoints) // 扣除积分。
	}
}
