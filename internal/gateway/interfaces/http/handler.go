package http

import (
	"net/http"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/gateway/application"
	"github.com/wyfcoding/ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.GatewayService
	logger  *slog.Logger
}

func NewHandler(service *application.GatewayService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoute 注册路由
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
		h.logger.Error("Failed to register route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Route registered successfully", route)
}

// ListRoutes 获取路由列表
func (h *Handler) ListRoutes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRoutes(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list routes", "error", err)
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

// DeleteRoute 删除路由
func (h *Handler) DeleteRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteRoute(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete route", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete route", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Route deleted successfully", nil)
}

// AddRateLimitRule 添加限流规则
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
		h.logger.Error("Failed to add rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to add rate limit rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Rate limit rule added successfully", rule)
}

// ListRateLimitRules 获取限流规则列表
func (h *Handler) ListRateLimitRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.service.ListRateLimitRules(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list rate limit rules", "error", err)
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

// DeleteRateLimitRule 删除限流规则
func (h *Handler) DeleteRateLimitRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	if err := h.service.DeleteRateLimitRule(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete rate limit rule", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to delete rate limit rule", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Rate limit rule deleted successfully", nil)
}

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
