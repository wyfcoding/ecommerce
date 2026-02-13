package http

import (
	"net/http"

	"github.com/wyfcoding/ecommerce/internal/identity/application"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	identitySvc *application.IdentityApplicationService
	kycBridge   *application.KYCBridgeService
}

func NewHandler(identitySvc *application.IdentityApplicationService, kycBridge *application.KYCBridgeService) *Handler {
	return &Handler{
		identitySvc: identitySvc,
		kycBridge:   kycBridge,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/persona/:user_id", h.GetUserPersona)
	r.PUT("/persona/:user_id", h.UpdateUserPersona)
	r.GET("/accounts/:user_id", h.GetLinkedAccounts)
	r.POST("/accounts/bind", h.BindAccount)
	r.DELETE("/accounts/unbind", h.UnbindAccount)
	r.GET("/sessions/:user_id", h.GetActiveSessions)
	r.DELETE("/sessions/:session_id", h.RevokeSession)
	r.POST("/token/exchange", h.ExchangeTokenForTrading)
	r.POST("/token/verify", h.VerifyTradingSession)
	r.POST("/mapping/bind", h.BindUserMapping)
	r.GET("/mapping", h.GetUserMapping)
	r.POST("/kyc/sync", h.SyncKYCStatus)
	r.POST("/aml/check", h.CheckAML)
}

func (h *Handler) GetUserPersona(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) UpdateUserPersona(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) GetLinkedAccounts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) BindAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) UnbindAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) GetActiveSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) RevokeSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) ExchangeTokenForTrading(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.identitySvc.ExchangeTokenForTrading(c.Request.Context(), string(rune(req.UserID)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trading_token": token,
		"expires_in":    3600,
	})
}

func (h *Handler) VerifyTradingSession(c *gin.Context) {
	var req struct {
		TradingToken string `json:"trading_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.identitySvc.VerifyTradingSession(c.Request.Context(), req.TradingToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":       session.SessionID,
		"ecommerce_user_id": session.UserID,
		"sso_token":        session.SsoToken,
		"expires_at":       session.ExpiresAt,
	})
}

func (h *Handler) BindUserMapping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) GetUserMapping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

func (h *Handler) SyncKYCStatus(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.kycBridge == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kyc bridge not configured"})
		return
	}

	err := h.kycBridge.CompleteKYCAndSync(c.Request.Context(), req.UserID, "SYSTEM")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "synced successfully"})
}

func (h *Handler) CheckAML(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id"`
		IDNumber string `json:"id_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.kycBridge == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kyc bridge not configured"})
		return
	}

	passed, err := h.kycBridge.PerformGlobalAML(c.Request.Context(), req.UserID, req.IDNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passed": passed,
	})
}
