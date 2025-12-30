package http

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/xerrors"
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
		response.Error(c, xerrors.InvalidArg("invalid request body").WithDetail(err.Error()))
		return
	}

	// 调用 Manager 层：现在 Manager 负责身份校验、角色提取和 Token 签发
	// 默认配置在此传入，生产环境建议从 Handler 的 Config 字段读取
	token, user, err := h.svc.Manager.Login(
		c.Request.Context(),
		req.Username,
		req.Password,
		"ecommerce-secret-key",
		"ecommerce",
		24*time.Hour,
	)

	if err != nil {
		response.Error(c, xerrors.Unauthenticated("login failed").WithDetail(err.Error()))
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
		response.Error(c, xerrors.InvalidArg("invalid input").WithDetail(err.Error()))
		return
	}

	_, err := h.svc.Manager.RegisterAdmin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, xerrors.Internal("failed to create admin", err))
		return
	}

	response.Success(c, gin.H{"message": "admin user registered successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	response.Success(c, gin.H{"msg": "profile info"})
}
