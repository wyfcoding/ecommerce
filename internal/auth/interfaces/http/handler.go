package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/auth/application"
	"github.com/wyfcoding/pkg/response"
)

type AuthHandler struct {
	appService *application.AuthService
}

func NewAuthHandler(app *application.AuthService) *AuthHandler {
	return &AuthHandler{appService: app}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	result, err := h.appService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
