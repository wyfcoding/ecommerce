package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/pkg/response"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// Handler 结构体定义了Gateway模块的HTTP处理层。
type Handler struct {
	service *application.GatewayService
	logger  *slog.Logger
}

// NewHandler 创建并返回一个新的 Gateway HTTP Handler 实例。
func NewHandler(service *application.GatewayService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoute 处理注册API路由的HTTP请求。
func (h *Handler) RegisterRoute(c *gin.Context) {
	var req struct {
		Path        string `json:"path" binding:"required"`
		Method      string `json:"method" binding:"required"`
		Service     string `json:"service" binding:"required"`
		Backend     string `json:"backend" binding:"required"`
		Timeout     int32  `json:"timeout"`
		Retries     int32  `json:"retries"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	route, err := h.service.RegisterRoute(c.Request.Context(), req.Path, req.Method, req.Service, req.Backend, req.Timeout, req.Retries, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to register route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Route registered successfully", route)
}

// ListRoutes 处理获取路由列表的HTTP请求。
func (h *Handler) ListRoutes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRoutes(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list routes", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list routes", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Routes listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteRoute 处理删除API路由的HTTP请求。
func (h *Handler) DeleteRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteRoute(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Route deleted successfully", nil)
}

// AddRateLimitRule 处理添加限流规则的HTTP请求。
func (h *Handler) AddRateLimitRule(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Path        string `json:"path" binding:"required"`
		Method      string `json:"method" binding:"required"`
		Limit       int32  `json:"limit" binding:"required"`
		Window      int32  `json:"window" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	rule, err := h.service.AddRateLimitRule(c.Request.Context(), req.Name, req.Path, req.Method, req.Limit, req.Window, req.Description)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to add rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add rate limit rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Rate limit rule added successfully", rule)
}

// ListRateLimitRules 处理获取限流规则列表的HTTP请求。
func (h *Handler) ListRateLimitRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRateLimitRules(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to list rate limit rules", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list rate limit rules", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Rate limit rules listed successfully", gin.H{
		"data":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteRateLimitRule 处理删除限流规则的HTTP请求。
func (h *Handler) DeleteRateLimitRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteRateLimitRule(c.Request.Context(), id); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "Failed to delete rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete rate limit rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Rate limit rule deleted successfully", nil)
}

// RegisterRoutes 在给定的Gin路由组中注册Gateway模块的HTTP路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/gateway")
	{
		group.POST("/routes", h.RegisterRoute)
		group.GET("/routes", h.ListRoutes)
		group.DELETE("/routes/:id", h.DeleteRoute)
		group.POST("/ratelimits", h.AddRateLimitRule)
		group.GET("/ratelimits", h.ListRateLimitRules)
		group.DELETE("/ratelimits/:id", h.DeleteRateLimitRule)
	}
}
