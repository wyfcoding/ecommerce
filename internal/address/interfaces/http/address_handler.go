package http

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wyfcoding/ecommerce/internal/address/application"
	"github.com/wyfcoding/ecommerce/internal/address/domain"
	"github.com/wyfcoding/pkg/response"
)

type AddressHandler struct {
	svc    *application.AddressService
	logger *slog.Logger
}

func NewAddressHandler(svc *application.AddressService, logger *slog.Logger) *AddressHandler {
	return &AddressHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *AddressHandler) RegisterRoutes(router gin.IRouter) {
	group := router.Group("/api/v1/addresses")
	{
		group.POST("", h.CreateAddress)
		group.PUT("/:id", h.UpdateAddress)
		group.DELETE("/:id", h.DeleteAddress)
		group.PUT("/:id/default", h.SetDefaultAddress)
		group.GET("", h.ListAddresses)
		group.GET("/:id", h.GetAddress)
	}
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {
	var req struct {
		UserID        int64  `json:"user_id" binding:"required"`
		RecipientName string `json:"recipient_name" binding:"required"`
		PhoneNumber   string `json:"phone_number" binding:"required"`
		Country       string `json:"country"`
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		DetailAddress string `json:"detail_address" binding:"required"`
		PostalCode    string `json:"postal_code"`
		IsDefault     bool   `json:"is_default"`
		Type          int32  `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	addr := &domain.Address{
		UserID:        req.UserID,
		RecipientName: req.RecipientName,
		PhoneNumber:   req.PhoneNumber,
		Country:       req.Country,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		Type:          req.Type,
	}

	if err := h.svc.CreateAddress(c.Request.Context(), addr); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addr)
}

func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		UserID        int64  `json:"user_id" binding:"required"`
		RecipientName string `json:"recipient_name" binding:"required"`
		PhoneNumber   string `json:"phone_number" binding:"required"`
		Country       string `json:"country"`
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		DetailAddress string `json:"detail_address" binding:"required"`
		PostalCode    string `json:"postal_code"`
		IsDefault     bool   `json:"is_default"`
		Type          int32  `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	addr := &domain.Address{
		ID:            id,
		UserID:        req.UserID,
		RecipientName: req.RecipientName,
		PhoneNumber:   req.PhoneNumber,
		Country:       req.Country,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		Type:          req.Type,
	}

	if err := h.svc.UpdateAddress(c.Request.Context(), addr); err != nil {
		response.Error(c, err)
		return
	}

	updatedAddr, err := h.svc.GetAddress(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, updatedAddr)
}

func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	id := c.Param("id")
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.DeleteAddress(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.SetDefaultAddress(c.Request.Context(), id, req.UserID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *AddressHandler) ListAddresses(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, err)
		return
	}

	addrs, err := h.svc.ListAddresses(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addrs)
}

func (h *AddressHandler) GetAddress(c *gin.Context) {
	id := c.Param("id")
	addr, err := h.svc.GetAddress(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, addr)
}
