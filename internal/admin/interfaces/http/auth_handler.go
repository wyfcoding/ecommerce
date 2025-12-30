package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/pkg/response"
)

// AuthHandler 处理登录与注册。
type AuthHandler struct {
	svc    *application.AdminService
	logger *slog.Logger
}

// NewAuthHandler 创建 AuthHandler 实例。
func NewAuthHandler(svc *application.AdminService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		svc:    svc,
		logger: logger,
	}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup, secret string) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.GET("/me", h.Me)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 适配原始 response 包：使用 BadRequest
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	// 调用 Manager 层
	token, user, err := h.svc.Manager.Login(
		c.Request.Context(),
		req.Username,
		req.Password,
		"ecommerce-secret-key",
		"ecommerce",
		24*time.Hour,
	)
	if err != nil {
		// 适配原始 response 包：使用 Unauthorized
		response.ErrorWithStatus(c, http.StatusUnauthorized, "login failed: "+err.Error(), "")
		return
	}

	// 统一成功响应结构
	response.Success(c, gin.H{
		"token": token,
		"user": application.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
		},
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req application.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid input: "+err.Error(), "")
		return
	}

	_, err := h.svc.Manager.RegisterAdmin(c.Request.Context(), &req)
	if err != nil {
		// 适配原始 response 包：使用 InternalError
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to create admin: "+err.Error(), "")
		return
	}

	response.Success(c, gin.H{"message": "admin user registered successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	response.Success(c, gin.H{"msg": "profile info"})
}
