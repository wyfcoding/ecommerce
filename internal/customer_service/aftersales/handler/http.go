package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/model"
	"ecommerce/internal/aftersales/service"
)

// AftersalesHandler 负责处理售后的 HTTP 请求。
type AftersalesHandler struct {
	svc    service.AftersalesService // 业务逻辑服务接口
	logger *zap.Logger
}

// NewAftersalesHandler 创建一个新的 AftersalesHandler 实例。
func NewAftersalesHandler(svc service.AftersalesService, logger *zap.Logger) *AftersalesHandler {
	return &AftersalesHandler{svc: svc, logger: logger}
}

// RegisterRoutes 在 Gin 引擎上注册所有售后相关的路由。
func (h *AftersalesHandler) RegisterRoutes(r *gin.Engine) {
	// 用户端路由
	userGroup := r.Group("/api/v1/aftersales/applications")
	// userGroup.Use(auth.AuthMiddleware(...)) // 认证中间件
	{
		userGroup.POST("", h.CreateApplication)
		userGroup.GET("", h.ListApplicationsForUser)
		userGroup.GET("/:id", h.GetApplicationForUser)
		userGroup.PUT("/:id/cancel", h.CancelApplicationForUser)
	}

	// 管理端路由
	adminGroup := r.Group("/api/v1/admin/aftersales/applications")
	// adminGroup.Use(auth.AuthMiddleware(...), auth.AdminMiddleware(...)) // 认证和权限中间件
	{
		adminGroup.GET("", h.ListApplicationsForAdmin)
		adminGroup.GET("/:id", h.GetApplicationForAdmin)
		adminGroup.PUT("/:id/approve", h.ApproveApplication)
		adminGroup.PUT("/:id/reject", h.RejectApplication)
		adminGroup.POST("/:id/process-return", h.ProcessReturn)
		adminGroup.PUT("/:id/complete", h.CompleteApplication)
		adminGroup.PUT("/:id/cancel", h.CancelApplicationForAdmin)
	}
}

// CreateApplicationRequest 定义了创建售后申请的请求体。
type CreateApplicationRequest struct {
	OrderID uint64                  `json:"order_id" binding:"required"`
	Type    model.ApplicationType `json:"type" binding:"required"` // RETURN, EXCHANGE, REPAIR
	Reason  string                  `json:"reason" binding:"required"`
	UserRemarks string              `json:"user_remarks"`
	Items   []struct {
		OrderItemID uint64 `json:"order_item_id" binding:"required"`
		ProductID   uint64 `json:"product_id" binding:"required"`
		ProductSKU  string `json:"product_sku" binding:"required"`
		Quantity    int    `json:"quantity" binding:"required,gt=0"`
	} `json:"items" binding:"required,min=1"`
}

// CreateApplication 处理用户提交售后申请的请求。
func (h *AftersalesHandler) CreateApplication(c *gin.Context) {
	// 从认证中间件获取用户ID，这里简化为假设已获取
	userID := uint(1) // 伪代码: 实际应从 JWT 或 Session 中获取

	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("CreateApplication: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	bizItems := make([]model.AftersalesItem, len(req.Items))
	for i, item := range req.Items {
		bizItems[i] = model.AftersalesItem{
			OrderItemID: uint(item.OrderItemID),
			ProductID:   uint(item.ProductID),
			ProductSKU:  item.ProductSKU,
			Quantity:    item.Quantity,
		}
	}

	app, err := h.svc.CreateApplication(c.Request.Context(), userID, uint(req.OrderID), req.Type, req.Reason, bizItems)
	if err != nil {
		h.logger.Error("Failed to create aftersales application", zap.Error(err), zap.Uint("user_id", userID), zap.Uint64("order_id", req.OrderID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Aftersales application created successfully", "application_sn": app.ApplicationSN, "application_id": app.ID})
}

// ListApplicationsForUser 处理获取用户售后申请列表的请求。
func (h *AftersalesHandler) ListApplicationsForUser(c *gin.Context) {
	userID := uint(1) // 伪代码: 实际应从 JWT 或 Session 中获取

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusFilterStr := c.Query("status")
	var statusFilter *model.ApplicationStatus
	if statusFilterStr != "" {
		s := model.ApplicationStatus(statusFilterStr)
		statusFilter = &s
	}

	apps, total, err := h.svc.ListApplications(c.Request.Context(), &userID, statusFilter, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list aftersales applications for user", zap.Error(err), zap.Uint("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list applications: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps, "total": total})
}

// GetApplicationForUser 处理获取用户售后申请详情的请求。
func (h *AftersalesHandler) GetApplicationForUser(c *gin.Context) {
	userID := uint(1) // 伪代码: 实际应从 JWT 或 Session 中获取

	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GetApplicationForUser: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	app, err := h.svc.GetApplication(c.Request.Context(), uint(appID), &userID, false) // 非管理员调用
	if err != nil {
		h.logger.Error("Failed to get aftersales application for user", zap.Error(err), zap.Uint("user_id", userID), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrPermissionDenied) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// CancelApplicationForUser 处理用户取消售后申请的请求。
func (h *AftersalesHandler) CancelApplicationForUser(c *gin.Context) {
	userID := uint(1) // 伪代码: 实际应从 JWT 或 Session 中获取

	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("CancelApplicationForUser: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	// 验证用户是否有权限取消此申请
	app, err := h.svc.GetApplication(c.Request.Context(), uint(appID), &userID, false)
	if err != nil {
		h.logger.Error("Failed to get application for cancellation", zap.Error(err), zap.Uint("user_id", userID), zap.Uint64("app_id", appID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application: " + err.Error()})
		return
	}

	if app.UserID != userID {
		h.logger.Warn("Permission denied to cancel application", zap.Uint("user_id", userID), zap.Uint64("app_id", appID))
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 调用业务逻辑层取消申请
	updatedApp, err := h.svc.CancelApplication(c.Request.Context(), uint(appID), "User cancelled")
	if err != nil {
		h.logger.Error("Failed to cancel aftersales application", zap.Error(err), zap.Uint("user_id", userID), zap.Uint64("app_id", appID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aftersales application cancelled successfully", "application": updatedApp})
}

// ListApplicationsForAdmin 处理管理员获取售后申请列表的请求。
func (h *AftersalesHandler) ListApplicationsForAdmin(c *gin.Context) {
	// 假设管理员ID从认证中间件获取，这里简化为不校验
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	statusFilterStr := c.Query("status")
	var statusFilter *model.ApplicationStatus
	if statusFilterStr != "" {
		s := model.ApplicationStatus(statusFilterStr)
		statusFilter = &s
	}

	apps, total, err := h.svc.ListApplications(c.Request.Context(), nil, statusFilter, page, pageSize) // 管理员可以查看所有用户申请
	if err != nil {
		h.logger.Error("Failed to list aftersales applications for admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list applications: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps, "total": total})
}

// GetApplicationForAdmin 处理管理员获取售后申请详情的请求。
func (h *AftersalesHandler) GetApplicationForAdmin(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GetApplicationForAdmin: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	app, err := h.svc.GetApplication(c.Request.Context(), uint(appID), nil, true) // 管理员调用
	if err != nil {
		h.logger.Error("Failed to get aftersales application for admin", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, app)
}

// ApproveApplication 处理管理员审核通过售后申请的请求。
type ApproveApplicationRequest struct {
	AdminRemarks string `json:"admin_remarks"`
}

func (h *AftersalesHandler) ApproveApplication(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("ApproveApplication: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req ApproveApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("ApproveApplication: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedApp, err := h.svc.ApproveApplication(c.Request.Context(), uint(appID), req.AdminRemarks)
	if err != nil {
		h.logger.Error("Failed to approve aftersales application", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrInvalidApplicationStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aftersales application approved successfully", "application": updatedApp})
}

// RejectApplication 处理管理员拒绝售后申请的请求。
type RejectApplicationRequest struct {
	AdminRemarks string `json:"admin_remarks"`
}

func (h *AftersalesHandler) RejectApplication(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("RejectApplication: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req RejectApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("RejectApplication: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedApp, err := h.svc.RejectApplication(c.Request.Context(), uint(appID), req.AdminRemarks)
	if err != nil {
		h.logger.Error("Failed to reject aftersales application", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrInvalidApplicationStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aftersales application rejected successfully", "application": updatedApp})
}

// ProcessReturn 处理管理员处理退货的请求。
type ProcessReturnRequest struct {
	RefundAmount float64 `json:"refund_amount" binding:"required,gte=0"`
}

func (h *AftersalesHandler) ProcessReturn(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("ProcessReturn: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req ProcessReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("ProcessReturn: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedApp, err := h.svc.ProcessReturnedGoods(c.Request.Context(), uint(appID), req.RefundAmount)
	if err != nil {
		h.logger.Error("Failed to process returned goods", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrInvalidApplicationStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process returned goods: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Returned goods processed successfully", "application": updatedApp})
}

// CompleteApplication 处理管理员完成售后申请的请求。
type CompleteApplicationRequest struct {
	AdminRemarks string `json:"admin_remarks"`
}

func (h *AftersalesHandler) CompleteApplication(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("CompleteApplication: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req CompleteApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("CompleteApplication: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedApp, err := h.svc.CompleteApplication(c.Request.Context(), uint(appID), req.AdminRemarks)
	if err != nil {
		h.logger.Error("Failed to complete aftersales application", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrInvalidApplicationStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aftersales application completed successfully", "application": updatedApp})
}

// CancelApplicationForAdmin 处理管理员取消售后申请的请求。
type CancelApplicationForAdminRequest struct {
	AdminRemarks string `json:"admin_remarks"`
}

func (h *AftersalesHandler) CancelApplicationForAdmin(c *gin.Context) {
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("CancelApplicationForAdmin: invalid application ID format", zap.String("app_id_str", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req CancelApplicationForAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("CancelApplicationForAdmin: invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedApp, err := h.svc.CancelApplication(c.Request.Context(), uint(appID), req.AdminRemarks)
	if err != nil {
		h.logger.Error("Failed to cancel aftersales application by admin", zap.Error(err), zap.Uint64("app_id", appID))
		if errors.Is(err, service.ErrApplicationNotFound) || errors.Is(err, service.ErrInvalidApplicationStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aftersales application cancelled by admin successfully", "application": updatedApp})
}
