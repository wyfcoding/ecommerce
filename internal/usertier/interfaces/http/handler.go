package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/usertier/application"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了用户等级模块的HTTP处理层。
type Handler struct {
	service *application.UserTierService
	logger  *slog.Logger
}

// NewHandler 创建并返回一个新的 UserTier HTTP Handler 实例。
func NewHandler(service *application.UserTierService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetUserTier 处理获取用户等级的HTTP请求。
func (h *Handler) GetUserTier(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	tier, err := h.service.GetUserTier(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get user tier", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user tier", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "User tier retrieved successfully", tier)
}

// AddScore 处理增加用户成长值的HTTP请求。
func (h *Handler) AddScore(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Score  int64  `json:"score" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AddScore(c.Request.Context(), req.UserID, req.Score); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add score", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add score", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Score added successfully", nil)
}

// GetPoints 处理获取用户积分的HTTP请求。
func (h *Handler) GetPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	points, err := h.service.GetPoints(c.Request.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to get points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get points", err.Error())
		return
	}

	response.Success(c, gin.H{"points": points})
}

// AddPoints 处理增加用户积分的HTTP请求。
func (h *Handler) AddPoints(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Points int64  `json:"points" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AddPoints(c.Request.Context(), req.UserID, req.Points, req.Reason); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add points", err.Error())
		return
	}

	response.Success(c, nil)
}

// DeductPoints 处理扣除用户积分的HTTP请求。
func (h *Handler) DeductPoints(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Points int64  `json:"points" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.DeductPoints(c.Request.Context(), req.UserID, req.Points, req.Reason); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to deduct points", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to deduct points", err.Error())
		return
	}

	response.Success(c, nil)
}

// ListPointsLogs 处理获取积分日志列表的HTTP请求。
func (h *Handler) ListPointsLogs(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListPointsLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list points logs", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list points logs", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Exchange 处理积分兑换商品的HTTP请求。
func (h *Handler) Exchange(c *gin.Context) {
	var req struct {
		UserID     uint64 `json:"user_id" binding:"required"`
		ExchangeID uint64 `json:"exchange_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.Exchange(c.Request.Context(), req.UserID, req.ExchangeID); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to exchange", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to exchange", err.Error())
		return
	}

	response.Success(c, nil)
}

// ListExchanges 处理获取兑换商品列表的HTTP请求。
func (h *Handler) ListExchanges(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListExchanges(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list exchanges", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list exchanges", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/usertier")
	{
		group.GET("/:user_id", h.GetUserTier)
		group.POST("/score", h.AddScore)
		group.GET("/points/:user_id", h.GetPoints)
		group.POST("/points/add", h.AddPoints)
		group.POST("/points/deduct", h.DeductPoints)
		group.GET("/points/logs", h.ListPointsLogs)
		group.POST("/exchange", h.Exchange)
		group.GET("/exchanges", h.ListExchanges)
	}
}
