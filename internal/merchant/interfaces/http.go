// Package interfaces 商家服务接口层
// 生成摘要：
// 1) 实现 HTTP Gin handler
// 2) 实现 gRPC handler
package interfaces

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/merchant/application"
	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
)

// HTTPHandler HTTP 接口处理器
type HTTPHandler struct {
	commandService *application.CommandService
	queryService   *application.QueryService
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(
	commandService *application.CommandService,
	queryService *application.QueryService,
) *HTTPHandler {
	return &HTTPHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

// RegisterRoutes 注册路由
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	merchants := r.Group("/merchants")
	{
		merchants.POST("/apply", h.Apply)
		merchants.POST("/:id/approve", h.Approve)
		merchants.POST("/:id/reject", h.Reject)
		merchants.PUT("/:id", h.Update)
		merchants.GET("/:id", h.Get)
		merchants.GET("/user/:userId", h.GetByUserID)
		merchants.GET("", h.List)
		merchants.POST("/:id/disable", h.Disable)
		merchants.POST("/:id/enable", h.Enable)
		merchants.PUT("/:id/settings", h.UpdateSettings)
		merchants.GET("/:id/settings", h.GetSettings)
	}

	stores := r.Group("/stores")
	{
		stores.POST("", h.CreateStore)
		stores.PUT("/:id", h.UpdateStore)
		stores.GET("/:id", h.GetStore)
		stores.GET("/merchant/:merchantId", h.ListStores)
	}
}

// ApplyRequest 入驻申请请求
type ApplyRequest struct {
	UserID       uint64                      `json:"user_id" binding:"required"`
	Name         string                      `json:"name" binding:"required"`
	LegalName    string                      `json:"legal_name" binding:"required"`
	LegalIDCard  string                      `json:"legal_id_card" binding:"required"`
	ContactName  string                      `json:"contact_name" binding:"required"`
	ContactPhone string                      `json:"contact_phone" binding:"required"`
	ContactEmail string                      `json:"contact_email"`
	Type         int8                        `json:"type"`
	License      *application.LicenseDTO     `json:"license"`
	BankAccount  *application.BankAccountDTO `json:"bank_account"`
	LogoURL      string                      `json:"logo_url"`
	Description  string                      `json:"description"`
}

// Apply 入驻申请
func (h *HTTPHandler) Apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ApplyCommand{
		UserID:       req.UserID,
		Name:         req.Name,
		LegalName:    req.LegalName,
		LegalIDCard:  req.LegalIDCard,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Type:         domain.MerchantType(req.Type),
		License:      req.License,
		BankAccount:  req.BankAccount,
		LogoURL:      req.LogoURL,
		Description:  req.Description,
	}

	result, err := h.commandService.Apply(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"merchant_id": result.MerchantID,
		"merchant_no": result.MerchantNo,
	})
}

// ApproveRequest 审核通过请求
type ApproveRequest struct {
	CommissionRate float64 `json:"commission_rate"`
	Operator       string  `json:"operator" binding:"required"`
	Remark         string  `json:"remark"`
}

// Approve 审核通过
func (h *HTTPHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.ApproveCommand{
		MerchantID:     uint(id),
		CommissionRate: req.CommissionRate,
		Operator:       req.Operator,
		Remark:         req.Remark,
	}

	if err := h.commandService.Approve(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "approved"})
}

// RejectRequest 审核拒绝请求
type RejectRequest struct {
	Reason   string `json:"reason" binding:"required"`
	Operator string `json:"operator" binding:"required"`
}

// Reject 审核拒绝
func (h *HTTPHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.RejectCommand{
		MerchantID: uint(id),
		Reason:     req.Reason,
		Operator:   req.Operator,
	}

	if err := h.commandService.Reject(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rejected"})
}

// Update 更新商家信息
func (h *HTTPHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req struct {
		Name         *string                     `json:"name"`
		ContactName  *string                     `json:"contact_name"`
		ContactPhone *string                     `json:"contact_phone"`
		ContactEmail *string                     `json:"contact_email"`
		LogoURL      *string                     `json:"logo_url"`
		Description  *string                     `json:"description"`
		BankAccount  *application.BankAccountDTO `json:"bank_account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.UpdateMerchantCommand{
		MerchantID:   uint(id),
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		LogoURL:      req.LogoURL,
		Description:  req.Description,
		BankAccount:  req.BankAccount,
	}

	if err := h.commandService.UpdateMerchant(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Get 获取商家信息
func (h *HTTPHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	merchant, err := h.queryService.GetMerchant(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// GetByUserID 根据用户ID获取商家
func (h *HTTPHandler) GetByUserID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	merchant, err := h.queryService.GetMerchantByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// List 商家列表
func (h *HTTPHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := &domain.MerchantFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			s := domain.MerchantStatus(status)
			filter.Status = &s
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		if t, err := strconv.Atoi(typeStr); err == nil {
			mt := domain.MerchantType(t)
			filter.Type = &mt
		}
	}

	result, err := h.queryService.ListMerchants(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Disable 禁用商家
func (h *HTTPHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req struct {
		Reason   string `json:"reason" binding:"required"`
		Operator string `json:"operator" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.DisableCommand{
		MerchantID: uint(id),
		Reason:     req.Reason,
		Operator:   req.Operator,
	}

	if err := h.commandService.Disable(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disabled"})
}

// Enable 启用商家
func (h *HTTPHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req struct {
		Operator string `json:"operator" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.EnableCommand{
		MerchantID: uint(id),
		Operator:   req.Operator,
	}

	if err := h.commandService.Enable(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "enabled"})
}

// UpdateSettings 更新商家设置
func (h *HTTPHandler) UpdateSettings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var req application.SettingsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.UpdateSettingsCommand{
		MerchantID: uint(id),
		Settings:   &req,
	}

	settings, err := h.commandService.UpdateSettings(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetSettings 获取商家设置
func (h *HTTPHandler) GetSettings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	settings, err := h.queryService.GetSettings(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// CreateStoreRequest 创建店铺请求
type CreateStoreRequest struct {
	MerchantID    uint     `json:"merchant_id" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	LogoURL       string   `json:"logo_url"`
	BannerURL     string   `json:"banner_url"`
	Description   string   `json:"description"`
	Categories    []string `json:"categories"`
	BusinessHours string   `json:"business_hours"`
	Address       string   `json:"address"`
}

// CreateStore 创建店铺
func (h *HTTPHandler) CreateStore(c *gin.Context) {
	var req CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CreateStoreCommand{
		MerchantID:    req.MerchantID,
		Name:          req.Name,
		LogoURL:       req.LogoURL,
		BannerURL:     req.BannerURL,
		Description:   req.Description,
		Categories:    req.Categories,
		BusinessHours: req.BusinessHours,
		Address:       req.Address,
	}

	store, err := h.commandService.CreateStore(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, store)
}

// UpdateStore 更新店铺
func (h *HTTPHandler) UpdateStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	var req struct {
		Name          *string  `json:"name"`
		LogoURL       *string  `json:"logo_url"`
		BannerURL     *string  `json:"banner_url"`
		Description   *string  `json:"description"`
		Announcement  *string  `json:"announcement"`
		Categories    []string `json:"categories"`
		IsOpen        *bool    `json:"is_open"`
		BusinessHours *string  `json:"business_hours"`
		Address       *string  `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.UpdateStoreCommand{
		StoreID:       uint(id),
		Name:          req.Name,
		LogoURL:       req.LogoURL,
		BannerURL:     req.BannerURL,
		Description:   req.Description,
		Announcement:  req.Announcement,
		Categories:    req.Categories,
		IsOpen:        req.IsOpen,
		BusinessHours: req.BusinessHours,
		Address:       req.Address,
	}

	store, err := h.commandService.UpdateStore(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, store)
}

// GetStore 获取店铺
func (h *HTTPHandler) GetStore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	store, err := h.queryService.GetStore(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, store)
}

// ListStores 店铺列表
func (h *HTTPHandler) ListStores(c *gin.Context) {
	merchantID, err := strconv.ParseUint(c.Param("merchantId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.queryService.ListStores(c.Request.Context(), uint(merchantID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
