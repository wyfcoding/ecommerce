package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/user/application"
	"github.com/wyfcoding/pkg/response"
)

type UserHandler struct {
	commandService *application.UserCommandService
	queryService   *application.UserQuery
	logger         *slog.Logger
}

func NewUserHandler(
	commandService *application.UserCommandService,
	queryService *application.UserQuery,
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

		// Address routes
		v1.POST("/:id/addresses", h.AddAddress)
		v1.GET("/:id/addresses", h.ListAddresses)
		v1.GET("/:id/addresses/:address_id", h.GetAddress)
		v1.PUT("/:id/addresses/:address_id", h.UpdateAddress)
		v1.DELETE("/:id/addresses/:address_id", h.DeleteAddress)
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req application.CreateUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	user, err := h.commandService.Register(c, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req application.LoginCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	ip := c.ClientIP()
	resp, err := h.commandService.Login(c, &req, ip)
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

	var req application.UpdateUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	req.ID = uint(id)

	if err := h.commandService.UpdateProfile(c, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) AddAddress(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req application.AddAddressCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	req.UserID = uint(userID)

	addr, err := h.commandService.AddAddress(c, &req)
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

	var req application.UpdateAddressCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	req.UserID = uint(userID)
	req.ID = uint(addrID)

	if err := h.commandService.UpdateAddress(c, &req); err != nil {
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
