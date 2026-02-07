package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/pkg/response"
)

type UserHandler struct {
	commandService *application.UserCommandService
	queryService   *application.UserQueryService
	logger         *slog.Logger
}

func NewUserHandler(
	commandService *application.UserCommandService,
	queryService *application.UserQueryService,
) *UserHandler {
	return &UserHandler{
		commandService: commandService,
		queryService:   queryService,
		logger:         slog.Default(),
	}
}

// RegisterHandlers 注册路由
func (h *UserHandler) RegisterHandlers(router *gin.Engine) {
	v1 := router.Group("/api/v1/users")
	{
		v1.POST("/register", h.Register)
		v1.POST("/login", h.Login)
		v1.GET("/:id", h.GetProfile)
		v1.PUT("/:id", h.UpdateProfile)
		v1.GET("/search", h.SearchUsers)

		// Address routes
		v1.POST("/:id/addresses", h.AddAddress)
		v1.GET("/:id/addresses", h.ListAddresses)
		v1.GET("/:id/addresses/:address_id", h.GetAddress)
		v1.PUT("/:id/addresses/:address_id", h.UpdateAddress)
		v1.DELETE("/:id/addresses/:address_id", h.DeleteAddress)
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	user, err := h.commandService.Register(c, application.CreateUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	ip := c.ClientIP()
	resp, err := h.commandService.Login(c, application.LoginCommand{
		Username: req.Username,
		Password: req.Password,
	}, ip)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	user, err := h.queryService.GetUser(c, uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}
	if user == nil {
		// response.Error(c, errors.New("not found"))
		c.Status(http.StatusNotFound)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req struct {
		Nickname string     `json:"nickname"`
		Avatar   string     `json:"avatar"`
		Gender   int8       `json:"gender"`
		Birthday *time.Time `json:"birthday"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	cmd := application.UpdateUserCommand{
		ID:       uint(id),
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Birthday: req.Birthday,
	}

	if err := h.commandService.UpdateProfile(c, cmd); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, total, err := h.queryService.SearchUsers(c, keyword, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"total": total,
		"items": users,
	})
}

func (h *UserHandler) AddAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req struct {
		RecipientName   string `json:"recipient_name" binding:"required"`
		PhoneNumber     string `json:"phone_number" binding:"required"`
		Province        string `json:"province" binding:"required"`
		City            string `json:"city" binding:"required"`
		District        string `json:"district" binding:"required"`
		DetailedAddress string `json:"detailed_address" binding:"required"`
		PostalCode      string `json:"postal_code"`
		IsDefault       bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	addr, err := h.commandService.AddAddress(c, application.AddAddressCommand{
		UserID:          uint(userID),
		RecipientName:   req.RecipientName,
		PhoneNumber:     req.PhoneNumber,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		DetailedAddress: req.DetailedAddress,
		PostalCode:      req.PostalCode,
		IsDefault:       req.IsDefault,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addr)
}

func (h *UserHandler) ListAddresses(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	addrs, err := h.queryService.ListAddresses(c, uint(userID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addrs)
}

func (h *UserHandler) GetAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	addrIDStr := c.Param("address_id")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	addr, err := h.queryService.GetAddress(c, uint(userID), uint(addrID))
	if err != nil {
		response.Error(c, err)
		return
	}
	if addr == nil {
		c.Status(http.StatusNotFound)
		return
	}

	response.Success(c, addr)
}

func (h *UserHandler) UpdateAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	addrIDStr := c.Param("address_id")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req struct {
		RecipientName   string `json:"recipient_name"`
		PhoneNumber     string `json:"phone_number"`
		Province        string `json:"province"`
		City            string `json:"city"`
		District        string `json:"district"`
		DetailedAddress string `json:"detailed_address"`
		PostalCode      string `json:"postal_code"`
		IsDefault       bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	cmd := application.UpdateAddressCommand{
		ID:              uint(addrID),
		UserID:          uint(userID),
		RecipientName:   req.RecipientName,
		PhoneNumber:     req.PhoneNumber,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		DetailedAddress: req.DetailedAddress,
		PostalCode:      req.PostalCode,
		IsDefault:       req.IsDefault,
	}

	if err := h.commandService.UpdateAddress(c, cmd); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) DeleteAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	addrIDStr := c.Param("address_id")
	addrID, err := strconv.ParseUint(addrIDStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.commandService.DeleteAddress(c, uint(userID), uint(addrID)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
