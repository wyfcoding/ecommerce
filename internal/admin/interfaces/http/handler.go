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

// Login 处理管理员登录，签发 JWT 令牌。
func (h *AdminHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	// 执行登录验证逻辑
	token, user, err := h.svc.Command.Login(
		c.Request.Context(),
		req.Username,
		req.Password,
		"ecommerce-secret-key",
		"ecommerce",
		24*time.Hour,
	)
	if err != nil {
		h.logger.WarnContext(c.Request.Context(), "admin login failed", "username", req.Username, "error", err)
		response.ErrorWithStatus(c, http.StatusUnauthorized, "invalid username or password", "")
		return
	}

	h.logger.InfoContext(c.Request.Context(), "admin logged in", "user_id", user.ID, "username", user.Username)
	response.Success(c, gin.H{
		"token": token,
		"user": application.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
		},
	})
}

// Register 处理新管理员账号的注册（通常仅限内部管理）。
func (h *AdminHandler) Register(c *gin.Context) {
	var req application.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request data", err.Error())
		return
	}

	user, err := h.svc.Command.RegisterAdmin(c.Request.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "admin registration failed", "username", req.Username, "error", err)
		response.Error(c, err)
		return
	}

	h.logger.InfoContext(c.Request.Context(), "new admin registered", "user_id", user.ID, "username", user.Username)
	response.Success(c, gin.H{"message": "admin user registered successfully"})
}

// Me 返回当前登录管理员的基础个人资料。
func (h *AdminHandler) Me(c *gin.Context) {
	// 此处仅为占位实现，实际应从 Context 或 DB 中拉取完整资料
	response.Success(c, gin.H{"msg": "profile info"})
}

// --- Workflow Handlers ---

// Apply 提交一个新的高风险操作审批申请。
func (h *AdminHandler) Apply(c *gin.Context) {
	var req application.ApprovalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid approval data", err.Error())
		return
	}

	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized: missing user_id", "")
		return
	}

	domainReq := &domain.ApprovalRequest{
		RequesterID: uint(requesterID),
		ActionType:  req.ActionType,
		Description: req.Description,
		Payload:     req.Payload,
	}

	if err := h.svc.Command.CreateRequest(c.Request.Context(), domainReq); err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to create approval request", "user_id", requesterID, "type", req.ActionType, "error", err)
		response.Error(c, err)
		return
	}

	h.logger.InfoContext(c.Request.Context(), "approval request submitted", "req_id", domainReq.ID, "user_id", requesterID)
	response.Success(c, gin.H{"id": domainReq.ID})
}

// Action 处理对审批申请的“通过”或“拒绝”动作。
func (h *AdminHandler) Action(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid workflow id format", "")
		return
	}

	var req application.ApprovalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, "invalid action data", err.Error())
		return
	}

	approverID, ok := middleware.GetUserID(c)
	if !ok {
		response.ErrorWithStatus(c, http.StatusUnauthorized, "unauthorized: missing user_id", "")
		return
	}

	if req.Action == "approve" {
		err = h.svc.Command.ApproveRequest(c.Request.Context(), uint(id), uint(approverID), req.Comment)
	} else {
		err = h.svc.Command.RejectRequest(c.Request.Context(), uint(id), uint(approverID), req.Comment)
	}

	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "workflow action failed", "req_id", id, "action", req.Action, "error", err)
		response.Error(c, err)
		return
	}

	h.logger.InfoContext(c.Request.Context(), "workflow action processed", "req_id", id, "action", req.Action, "approver", approverID)
	response.Success(c, gin.H{"status": "processed"})
}
