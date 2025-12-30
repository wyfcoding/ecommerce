package http

import (
	"net/http"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/pkg/response"
)

type Handler struct {
	app    *application.UserService
	logger *slog.Logger
}

func NewHandler(app *application.UserService, logger *slog.Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/user")
	{
		v1.POST("/register", h.Register)
		v1.POST("/login", h.Login)

		v1.GET("/:id", h.GetUser)
		v1.PUT("/:id", h.UpdateProfile)

		addressGroup := v1.Group("/:id/addresses")
		{
			addressGroup.POST("", h.AddAddress)
			addressGroup.GET("", h.ListAddresses)
			addressGroup.GET("/:addressId", h.GetAddress)
			addressGroup.PUT("/:addressId", h.UpdateAddress)
			addressGroup.DELETE("/:addressId", h.DeleteAddress)
		}
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	user, err := h.app.Manager.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "user registration failed", "username", req.Username, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{"user_id": user.ID})
}

func (h *Handler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid credentials format", "")
		return
	}

	token, expiresAt, err := h.app.Manager.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		h.logger.WarnContext(c.Request.Context(), "login attempt failed", "username", req.Username, "ip", c.ClientIP(), "error", err)
		response.ErrorWithStatus(c, http.StatusUnauthorized, "invalid username or password", "")
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID format", "")
		return
	}

	user, err := h.app.Query.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get user info", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}
	if user == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "user not found", "")
		return
	}

	response.Success(c, user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID format", "")
		return
	}

	var req application.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid update data: "+err.Error(), "")
		return
	}

	user, err := h.app.Manager.UpdateProfile(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to update user profile", "id", id, "error", err)
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, user)
}

func (h *Handler) AddAddress(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	var req application.AddressDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid address data", "")
		return
	}

	addr, err := h.app.Manager.AddAddress(c.Request.Context(), uint(userID), &req)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to add address: "+err.Error(), "")
		return
	}

	response.Success(c, addr)
}

func (h *Handler) ListAddresses(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	list, err := h.app.Query.ListAddresses(c.Request.Context(), uint(userID))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, list)
}

func (h *Handler) GetAddress(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	addrID, err := strconv.ParseUint(c.Param("addressId"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid address ID", "")
		return
	}

	addr, err := h.app.Query.GetAddress(c.Request.Context(), uint(userID), uint(addrID))
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}
	if addr == nil {
		response.ErrorWithStatus(c, http.StatusNotFound, "address not found", "")
		return
	}

	response.Success(c, addr)
}

func (h *Handler) UpdateAddress(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	addrID, err := strconv.ParseUint(c.Param("addressId"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid address ID", "")
		return
	}

	var req application.AddressDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid data", "")
		return
	}

	addr, err := h.app.Manager.UpdateAddress(c.Request.Context(), uint(userID), uint(addrID), &req)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, addr)
}

func (h *Handler) DeleteAddress(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid user ID", "")
		return
	}

	addrID, err := strconv.ParseUint(c.Param("addressId"), 10, 64)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid address ID", "")
		return
	}

	if err := h.app.Manager.DeleteAddress(c.Request.Context(), uint(userID), uint(addrID)); err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, err.Error(), "")
		return
	}

	response.Success(c, gin.H{"status": "deleted"})
}
