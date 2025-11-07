package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/oauth/model"
	"ecommerce/internal/oauth/service"
	"ecommerce/pkg/response"
)

// OAuthHandler OAuth HTTP处理器
type OAuthHandler struct {
	service service.OAuthService
	logger  *zap.Logger
}

// NewOAuthHandler 创建OAuth HTTP处理器
func NewOAuthHandler(service service.OAuthService, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *OAuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	oauth := r.Group("/oauth")
	{
		// 获取授权URL
		oauth.GET("/:provider/authorize", h.GetAuthURL)
		
		// 处理回调
		oauth.GET("/:provider/callback", h.HandleCallback)
		
		// 绑定账号
		oauth.POST("/:provider/bind", h.BindAccount)
		
		// 解绑账号
		oauth.DELETE("/:provider/unbind", h.UnbindAccount)
		
		// 获取绑定列表
		oauth.GET("/bindings", h.GetBindings)
	}
}

// GetAuthURL 获取授权URL
func (h *OAuthHandler) GetAuthURL(c *gin.Context) {
	provider := model.OAuthProvider(c.Param("provider"))
	redirectURL := c.Query("redirect_url")
	
	if redirectURL == "" {
		response.Error(c, http.StatusBadRequest, "缺少redirect_url参数", nil)
		return
	}

	// 如果用户已登录，传递userID用于绑定
	userID := c.GetUint64("userID")

	authURL, err := h.service.GetAuthURL(c.Request.Context(), provider, redirectURL, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取授权URL失败", err)
		return
	}

	response.Success(c, gin.H{"authUrl": authURL})
}

// HandleCallback 处理OAuth回调
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	provider := model.OAuthProvider(c.Param("provider"))
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.Error(c, http.StatusBadRequest, "缺少code或state参数", nil)
		return
	}

	user, isNew, err := h.service.HandleCallback(c.Request.Context(), provider, code, state)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "OAuth登录失败", err)
		return
	}

	// TODO: 生成JWT token
	response.Success(c, gin.H{
		"user":  user,
		"isNew": isNew,
		// "token": token,
	})
}

// BindAccount 绑定第三方账号
func (h *OAuthHandler) BindAccount(c *gin.Context) {
	provider := model.OAuthProvider(c.Param("provider"))
	userID := c.GetUint64("userID")
	
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err)
		return
	}

	if err := h.service.BindAccount(c.Request.Context(), userID, provider, req.Code, req.State); err != nil {
		response.Error(c, http.StatusInternalServerError, "绑定账号失败", err)
		return
	}

	response.Success(c, nil)
}

// UnbindAccount 解绑第三方账号
func (h *OAuthHandler) UnbindAccount(c *gin.Context) {
	provider := model.OAuthProvider(c.Param("provider"))
	userID := c.GetUint64("userID")

	if err := h.service.UnbindAccount(c.Request.Context(), userID, provider); err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑账号失败", err)
		return
	}

	response.Success(c, nil)
}

// GetBindings 获取用户绑定列表
func (h *OAuthHandler) GetBindings(c *gin.Context) {
	userID := c.GetUint64("userID")

	bindings, err := h.service.GetUserBindings(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取绑定列表失败", err)
		return
	}

	response.Success(c, bindings)
}
