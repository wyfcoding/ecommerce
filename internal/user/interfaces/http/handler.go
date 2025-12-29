package http

import (
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
		// Auth
		v1.POST("/register", h.Register)
		v1.POST("/login", h.Login)

		// Profile (Need Auth Middleware in real world, but keeping simple for structure now)
		v1.GET("/:id", h.GetUser)
		v1.PUT("/:id", h.UpdateProfile)

		// Address
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

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.app.Manager.Register(c.Request.Context(), &req)
	if err != nil {
		slog.ErrorContext(c, "register failed", "err", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"user_id": user.ID})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, expiresAt, err := h.app.Manager.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// GetUser 获取用户
func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.app.Query.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新资料
func (h *Handler) UpdateProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req application.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.app.Manager.UpdateProfile(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// AddAddress 添加地址
func (h *Handler) AddAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req application.AddressDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	addr, err := h.app.Manager.AddAddress(c.Request.Context(), uint(userID), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addr)
}

// ListAddresses 地址列表
func (h *Handler) ListAddresses(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	list, err := h.app.Query.ListAddresses(c.Request.Context(), uint(userID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, list)
}

// GetAddress 获取地址
func (h *Handler) GetAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	addrIDStr := c.Param("addressId")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	addr, err := h.app.Query.GetAddress(c.Request.Context(), uint(userID), uint(addrID))
	if err != nil {
		response.Error(c, err)
		return
	}
	if addr == nil {
		response.NotFound(c, "address not found")
		return
	}

	response.Success(c, addr)
}

// UpdateAddress 更新地址
func (h *Handler) UpdateAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	addrIDStr := c.Param("addressId")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	var req application.AddressDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	addr, err := h.app.Manager.UpdateAddress(c.Request.Context(), uint(userID), uint(addrID), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addr)
}

// DeleteAddress 删除地址
func (h *Handler) DeleteAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	addrIDStr := c.Param("addressId")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid address id")
		return
	}

	if err := h.app.Manager.DeleteAddress(c.Request.Context(), uint(userID), uint(addrID)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"status": "ok"})
}
