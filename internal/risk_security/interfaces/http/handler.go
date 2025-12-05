package http

import (
	"net/http" // 导入HTTP状态码。
	"strconv"  // 导入字符串和数字转换工具。
	"time"     // 导入时间包，用于时间解析。

	"github.com/wyfcoding/ecommerce/internal/risk_security/application" // 导入风控安全模块的应用服务。
	"github.com/wyfcoding/ecommerce/pkg/response"                       // 导入统一的响应处理工具。

	"log/slog" // 导入结构化日志库。

	"github.com/gin-gonic/gin" // 导入Gin Web框架。
)

// Handler 结构体定义了RiskSecurity模块的HTTP处理层。
// 它是DDD分层架构中的接口层，负责接收HTTP请求，调用应用服务处理业务逻辑，并将结果封装为HTTP响应。
type Handler struct {
	service *application.RiskService // 依赖Risk应用服务，处理核心业务逻辑。
	logger  *slog.Logger             // 日志记录器，用于记录请求处理过程中的信息和错误。
}

// NewHandler 创建并返回一个新的 RiskSecurity HTTP Handler 实例。
func NewHandler(service *application.RiskService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// EvaluateRisk 处理评估风险的HTTP请求。
// Method: POST
// Path: /risk/evaluate
func (h *Handler) EvaluateRisk(c *gin.Context) {
	// 定义请求体结构，用于接收风险评估的参数。
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
		IP       string `json:"ip" binding:"required"`      // IP地址，必填。
		DeviceID string `json:"device_id"`                  // 设备ID，选填。
		Amount   int64  `json:"amount"`                     // 金额，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层评估风险。
	result, err := h.service.EvaluateRisk(c.Request.Context(), req.UserID, req.IP, req.DeviceID, req.Amount)
	if err != nil {
		h.logger.Error("Failed to evaluate risk", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to evaluate risk", err.Error())
		return
	}

	// 返回成功的响应，包含风险评估结果。
	response.SuccessWithStatus(c, http.StatusOK, "Risk evaluated successfully", result)
}

// AddToBlacklist 处理添加实体到黑名单的HTTP请求。
// Method: POST
// Path: /risk/blacklist
func (h *Handler) AddToBlacklist(c *gin.Context) {
	// 定义请求体结构，用于接收黑名单条目信息。
	var req struct {
		Type     string `json:"type" binding:"required"`  // 黑名单类型，必填。
		Value    string `json:"value" binding:"required"` // 黑名单值，必填。
		Reason   string `json:"reason"`                   // 原因，选填。
		Duration string `json:"duration"`                 // 有效时长，例如 "24h"，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 解析持续时间字符串，如果failed to parse，则使用默认的24小时。
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		duration = 24 * time.Hour // 默认有效期24小时。
	}

	// 调用应用服务层将实体添加到黑名单。
	if err := h.service.AddToBlacklist(c.Request.Context(), req.Type, req.Value, req.Reason, duration); err != nil {
		h.logger.Error("Failed to add to blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add to blacklist", err.Error())
		return
	}

	// 返回成功的响应，状态码为201 Created。
	response.SuccessWithStatus(c, http.StatusCreated, "Added to blacklist successfully", nil)
}

// RemoveFromBlacklist 处理从黑名单中移除实体的HTTP请求。
// Method: DELETE
// Path: /risk/blacklist/:id
func (h *Handler) RemoveFromBlacklist(c *gin.Context) {
	// 从URL路径中解析黑名单条目ID。
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	// 调用应用服务层从黑名单中移除实体。
	if err := h.service.RemoveFromBlacklist(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to remove from blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to remove from blacklist", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Removed from blacklist successfully", nil)
}

// RecordBehavior 处理记录用户行为的HTTP请求。
// Method: POST
// Path: /risk/behavior
func (h *Handler) RecordBehavior(c *gin.Context) {
	// 定义请求体结构，用于接收用户行为信息。
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"` // 用户ID，必填。
		IP       string `json:"ip" binding:"required"`      // IP地址，必填。
		DeviceID string `json:"device_id"`                  // 设备ID，选填。
	}

	// 绑定并验证请求JSON数据。
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// 调用应用服务层记录用户行为。
	if err := h.service.RecordUserBehavior(c.Request.Context(), req.UserID, req.IP, req.DeviceID); err != nil {
		h.logger.Error("Failed to record behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to record behavior", err.Error())
		return
	}

	// 返回成功的响应。
	response.SuccessWithStatus(c, http.StatusOK, "Behavior recorded successfully", nil)
}

// GetRiskAnalysisResult handles the request to get risk analysis result.
// Method: GET
// Path: /risk/result/:user_id
func (h *Handler) GetRiskAnalysisResult(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	result, err := h.service.GetRiskAnalysisResult(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get risk analysis result", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get risk analysis result", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Risk analysis result retrieved", result)
}

// CheckBlacklist handles the request to check if a value is blacklisted.
// Method: GET
// Path: /risk/blacklist
func (h *Handler) CheckBlacklist(c *gin.Context) {
	bType := c.Query("type")
	value := c.Query("value")

	if bType == "" || value == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Missing type or value", "type and value query parameters are required")
		return
	}

	blacklist, err := h.service.CheckBlacklist(c.Request.Context(), bType, value)
	if err != nil {
		// If not found, it might return error or nil depending on repo implementation.
		// Assuming error if not found or db error.
		// If it's a "not found" error, we should return 404 or just say not blacklisted.
		// For simplicity, let's return error for now.
		h.logger.Error("Failed to check blacklist", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to check blacklist", err.Error())
		return
	}

	if blacklist == nil {
		response.SuccessWithStatus(c, http.StatusOK, "Not blacklisted", gin.H{"blacklisted": false})
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Blacklisted", gin.H{"blacklisted": true, "entry": blacklist})
}

// GetUserBehavior handles the request to get user behavior.
// Method: GET
// Path: /risk/behavior/:user_id
func (h *Handler) GetUserBehavior(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid User ID", err.Error())
		return
	}

	behavior, err := h.service.GetUserBehavior(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user behavior", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get user behavior", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "User behavior retrieved", behavior)
}

// RegisterRoutes 在给定的Gin路由组中注册RiskSecurity模块的HTTP路由。
// r: Gin的路由组。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// /risk 路由组，用于所有风控安全相关接口。
	group := r.Group("/risk")
	{
		group.POST("/evaluate", h.EvaluateRisk)                // 评估风险。
		group.POST("/blacklist", h.AddToBlacklist)             // 添加到黑名单。
		group.GET("/blacklist", h.CheckBlacklist)              // 检查黑名单。
		group.DELETE("/blacklist/:id", h.RemoveFromBlacklist)  // 从黑名单移除。
		group.POST("/behavior", h.RecordBehavior)              // 记录用户行为。
		group.GET("/behavior/:user_id", h.GetUserBehavior)     // 获取用户行为。
		group.GET("/result/:user_id", h.GetRiskAnalysisResult) // 获取风险分析结果。
	}
}
