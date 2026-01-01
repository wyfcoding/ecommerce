package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/xerrors"
)

// AdminHandler 统一处理管理后台相关的 HTTP 请求。
type AdminHandler struct {
	svc    *application.AdminService
	logger *slog.Logger
}

// NewAdminHandler 创建 AdminHandler 实例。
func NewAdminHandler(svc *application.AdminService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		svc:    svc,
		logger: logger,
	}
}

// RegisterRoutes 注册所有管理后台路由。
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup, secret string) {
	// 1. 公开接口 (Auth)
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.GET("/me", h.Me)
	}

	// 2. 需鉴权接口 (Workflow & Others)
	protected := r.Group("/")
	protected.Use(middleware.JWTAuth(secret))
	{
		wf := protected.Group("/workflow")
		{
			wf.POST("/apply", h.Apply)
			wf.POST("/:id/action", middleware.HasRole("ADMIN"), h.Action)
		}
	}
}

// --- Auth Handlers ---

func (h *AdminHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request body: "+err.Error(), "")
		return
	}

	token, user, err := h.svc.Manager.Login(
		c.Request.Context(),
		req.Username,
		req.Password,
		"ecommerce-secret-key",
		"ecommerce",
		24*time.Hour,
	)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "login failed: "+err.Error(), "")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": application.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
		},
	})
}

func (h *AdminHandler) Register(c *gin.Context) {
	var req application.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid input: "+err.Error(), "")
		return
	}

	_, err := h.svc.Manager.RegisterAdmin(c.Request.Context(), &req)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusInternalServerError, "failed to create admin: "+err.Error(), "")
		return
	}

	response.Success(c, gin.H{"message": "admin user registered successfully"})
}

func (h *AdminHandler) Me(c *gin.Context) {
	response.Success(c, gin.H{"msg": "profile info"})
}

// --- Workflow Handlers ---

func (h *AdminHandler) Apply(c *gin.Context) {
	var req application.ApprovalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized: failed to get user ID", "")
		return
	}

	domainReq := &domain.ApprovalRequest{
		RequesterID: uint(requesterID),
		ActionType:  req.ActionType,
		Description: req.Description,
		Payload:     req.Payload,
	}

	if err := h.svc.Manager.CreateRequest(c.Request.Context(), domainReq); err != nil {
		response.Error(c, xerrors.Internal("failed to submit approval request", err))
		return
	}

	response.Success(c, gin.H{"id": domainReq.ID})
}

func (h *AdminHandler) Action(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid workflow ID format", "")
		return
	}

	var req application.ApprovalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	approverID, ok := middleware.GetUserID(c)
	if !ok {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized: failed to get user ID", "")
		return
	}

	if req.Action == "approve" {
		err = h.svc.Manager.ApproveRequest(c.Request.Context(), uint(id), uint(approverID), req.Comment)
	} else {
		err = h.svc.Manager.RejectRequest(c.Request.Context(), uint(id), uint(approverID), req.Comment)
	}

	if err != nil {
		response.Error(c, xerrors.Internal("workflow action failed", err))
		return
	}

	response.Success(c, gin.H{"status": "processed"})
}
