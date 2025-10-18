package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/auth/service"
)

// AuthHandler 负责处理认证的 HTTP 请求
type AuthHandler struct {
	svc    service.AuthService
	logger *zap.Logger
}

// NewAuthHandler 创建一个新的 AuthHandler 实例
func NewAuthHandler(svc service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有认证相关的路由
func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/v1/auth")
	{
		group.POST("/register", h.Register)
		group.POST("/login", h.Login)
		group.POST("/refresh", h.RefreshToken)

		// 这个端点用于演示令牌验证，需要认证中间件
		group.GET("/validate", AuthMiddleware(h.svc, h.logger), h.Validate)
	}
}

// RegisterRequest 定义了用户注册的请求体
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register 处理用户注册请求
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	userID, err := h.svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "user_id": userID})
}

// LoginRequest 定义了用户登录的请求体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login 处理用户登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	accessToken, refreshToken, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// RefreshTokenRequest 定义了刷新令牌的请求体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 处理刷新令牌的请求
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	newAccessToken, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "刷新令牌失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
}

// Validate 是一个受保护的端点，用于显示已认证用户的信息
func (h *AuthHandler) Validate(c *gin.Context) {
	userID, _ := c.Get("userID")
	roles, _ := c.Get("roles")
	c.JSON(http.StatusOK, gin.H{"message": "令牌有效", "user_id": userID, "roles": roles})
}

// AuthMiddleware 是一个 Gin 中间件，用于验证 JWT
func AuthMiddleware(svc service.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization 请求头"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization 请求头格式错误"})
			return
		}

		tokenString := parts[1]
		claims, err := svc.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌: " + err.Error()})
			return
		}

		// 将用户信息存入 Gin 的上下文，供后续处理函数使用
		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}