package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AuthHandler 处理 HTTP 或 gRPC 请求。
type AuthHandler struct {
	authService *application.AdminAuthService
	logger      *slog.Logger
}

// NewAuthHandler 处理 HTTP 或 gRPC 请求。
func NewAuthHandler(authService *application.AdminAuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register) // 通常受限，仅用于演示
		auth.GET("/me", h.Me)              // TODO: Add Middleware
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 暂时使用模拟 Token
	c.JSON(http.StatusOK, application.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req application.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.AdminUser{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	}

	if err := h.authService.CreateUser(c.Request.Context(), user, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// TODO: 获取 UserID from Context (Middleware)
	c.JSON(http.StatusOK, gin.H{"message": "current user info"})
}
