package http

import (
	"net/http"
	"strconv"

	"ecommerce/internal/admin/application"
	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type Handler struct {
	service *application.AdminService
	logger  *slog.Logger
}

func NewHandler(service *application.AdminService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterAdmin 注册管理员
func (h *Handler) RegisterAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		RealName string `json:"real_name"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	admin, err := h.service.RegisterAdmin(c.Request.Context(), req.Username, req.Email, req.Password, req.RealName, req.Phone)
	if err != nil {
		h.logger.Error("Failed to register admin", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to register admin", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Admin registered successfully", admin)
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		h.logger.Warn("Login failed", "username", req.Username, "error", err)
		response.ErrorWithStatus(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Login successful", gin.H{"token": token})
}

// GetProfile 获取个人信息
func (h *Handler) GetProfile(c *gin.Context) {
	// Assuming middleware sets "userID"
	userID, exists := c.Get("userID")
	if !exists {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "Unauthorized", "User ID not found in context")
		return
	}

	admin, err := h.service.GetAdminProfile(c.Request.Context(), userID.(uint64))
	if err != nil {
		h.logger.Error("Failed to get profile", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to get profile", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Profile retrieved successfully", admin)
}

// ListAdmins 获取管理员列表
func (h *Handler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	admins, total, err := h.service.ListAdmins(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list admins", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to list admins", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Admins listed successfully", gin.H{
		"data":      admins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	role, err := h.service.CreateRole(c.Request.Context(), req.Name, req.Code, req.Description)
	if err != nil {
		h.logger.Error("Failed to create role", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to create role", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusCreated, "Role created successfully", role)
}

// AssignRole 分配角色
func (h *Handler) AssignRole(c *gin.Context) {
	var req struct {
		AdminID uint64 `json:"admin_id" binding:"required"`
		RoleID  uint64 `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.service.AssignRoleToAdmin(c.Request.Context(), req.AdminID, req.RoleID); err != nil {
		h.logger.Error("Failed to assign role", "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, "Failed to assign role", err.Error())
		return
	}

	response.SuccessWithStatus(c, http.StatusOK, "Role assigned successfully", nil)
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", h.RegisterAdmin)
		authGroup.POST("/login", h.Login)
	}

	adminGroup := r.Group("/admin")
	// Middleware for auth would go here
	{
		adminGroup.GET("/profile", h.GetProfile)
		adminGroup.GET("/list", h.ListAdmins)
		adminGroup.POST("/role", h.CreateRole)
		adminGroup.POST("/assign-role", h.AssignRole)
	}
}
