package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/middleware"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/xerrors"
)

// WorkflowHandler 处理审批流程。
type WorkflowHandler struct {
	svc    *application.AdminService
	logger *slog.Logger
}

// NewWorkflowHandler 创建 WorkflowHandler 实例。
func NewWorkflowHandler(svc *application.AdminService, logger *slog.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *WorkflowHandler) RegisterRoutes(r *gin.RouterGroup) {
	wf := r.Group("/workflow")
	{
		// 提交审批：普通管理员即可
		wf.POST("/apply", h.Apply)

		// 执行审批：强制要求 ADMIN 角色
		wf.POST("/:id/action", middleware.HasRole("ADMIN"), h.Action)
	}
}

func (h *WorkflowHandler) Apply(c *gin.Context) {
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

func (h *WorkflowHandler) Action(c *gin.Context) {
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
