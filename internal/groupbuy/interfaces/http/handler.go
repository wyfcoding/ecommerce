package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/groupbuy/application" // 导入拼团模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                  // 导入统一的响应处理工具。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
	"log/slog"                 // 导入结构化日志库。
)

// Handler 结构体定义了Groupbuy模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.GroupbuyService // 依赖Groupbuy应用服务，处理核心业务逻辑。
	logger  *slog.Logger                 // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 Groupbuy HTTP Handler 实例。
func NewHandler(service *application.GroupbuyService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateGroupbuy 处理创建拼团活动的HTTP请求。
// Method: POST
// Path: /groupbuy
func (h *Handler) CreateGroupbuy(c *gin.Context) {
	// 定义请求体结构，用于接收拼团活动的创建信息。
	var req struct {
		Name          string    `json:"name" binding:"required"`           // 活动名称，必填。
		ProductID     uint64    `json:"product_id" binding:"required"`     // 商品ID，必填。
		SkuID         uint64    `json:"sku_id" binding:"required"`         // SKU ID，必填。
		OriginalPrice uint64    `json:"original_price" binding:"required"` // 原价，必填。
		GroupPrice    uint64    `json:"group_price" binding:"required"`    // 拼团价，必填。
		MinPeople     int32     `json:"min_people" binding:"required"`     // 最小成团人数，必填。
		MaxPeople     int32     `json:"max_people" binding:"required"`     // 最大成团人数，必填。
		TotalStock    int32     `json:"total_stock" binding:"required"`    // 总库存，必填。
		StartTime     time.Time `json:"start_time" binding:"required"`     // 开始时间，必填。
		EndTime       time.Time `json:"end_time" binding:"required"`       // 结束时间，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层创建拼团活动。
	groupbuy, err := h.service.CreateGroupbuy(c.Request.Context(), req.Name, req.ProductID, req.SkuID, req.OriginalPrice, req.GroupPrice,
		req.MinPeople, req.MaxPeople, req.TotalStock, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to create groupbuy", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create groupbuy", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Groupbuy created successfully", groupbuy)
}

// ListGroupbuys 处理获取拼团活动列表的HTTP请求。
// Method: GET
// Path: /groupbuy
func (h *Handler) ListGroupbuys(c *gin.Context) {
	// 从查询参数中获取页码和每页大小，并设置默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用应用服务层获取拼团活动列表。
	list, total, err := h.service.ListGroupbuys(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list groupbuys", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list groupbuys", err.Error())
		return
	}

	// 返回包含分页信息的成功响应。
	response.SuccessWithStatus(c, http.StatusOK, "Groupbuys listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// InitiateTeam 处理发起拼团团队的HTTP请求。
// Method: POST
// Path: /groupbuy/initiate
func (h *Handler) InitiateTeam(c *gin.Context) {
	// 定义请求体结构，用于接收发起团队所需的信息。
	var req struct {
		GroupbuyID uint64 `json:"groupbuy_id" binding:"required"` // 拼团活动ID，必填。
		UserID     uint64 `json:"user_id" binding:"required"`     // 发起用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层发起拼团团队。
	team, order, err := h.service.InitiateTeam(c.Request.Context(), req.GroupbuyID, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to initiate team", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to initiate team", err.Error())
		return
	}

	// 返回成功的响应，包含团队信息和团长订单信息。
	response.SuccessWithStatus(c, http.StatusOK, "Team initiated successfully", gin.H{
		"team":  team,
		"order": order,
	})
}

// JoinTeam 处理加入拼团团队的HTTP请求。
// Method: POST
// Path: /groupbuy/join
func (h *Handler) JoinTeam(c *gin.Context) {
	// 定义请求体结构，用于接收加入团队所需的信息。
	var req struct {
		TeamNo string `json:"team_no" binding:"required"` // 团队编号，必填。
		UserID uint64 `json:"user_id" binding:"required"` // 加入用户ID，必填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层加入拼团团队。
	order, err := h.service.JoinTeam(c.Request.Context(), req.TeamNo, req.UserID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to join team", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to join team", err.Error())
		return
	}

	// 返回成功的响应，包含加入团队后的订单信息。
	response.SuccessWithStatus(c, http.StatusOK, "Joined team successfully", order)
}

// GetTeamDetails 处理获取拼团团队详情的HTTP请求。
// Method: GET
// Path: /groupbuy/team/:team_id
func (h *Handler) GetTeamDetails(c *gin.Context) {
	// 从URL路径中解析团队ID。
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid team ID", err.Error())
		return
	}

	// 调用应用服务层获取团队详情。
	team, orders, err := h.service.GetTeamDetails(c.Request.Context(), teamID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get team details", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get team details", err.Error())
		return
	}

	// 返回成功的响应，包含团队信息和成员订单列表。
	response.SuccessWithStatus(c, http.StatusOK, "Team details retrieved successfully", gin.H{
		"team":   team,
		"orders": orders,
	})
}

// RegisterRoutes 在给定的Gin路由组中注册Groupbuy模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /groupbuy 路由组，用于所有拼团相关接口。
	group := r.Group("/groupbuy")
	{
		group.POST("", h.CreateGroupbuy)              // 创建拼团活动。
		group.GET("", h.ListGroupbuys)                // 获取拼团活动列表。
		group.POST("/initiate", h.InitiateTeam)       // 发起拼团。
		group.POST("/join", h.JoinTeam)               // 加入拼团。
		group.GET("/team/:team_id", h.GetTeamDetails) // 获取团队详情。
		// TODO: 补充获取拼团活动详情、更新活动、取消活动等接口。
	}
}
