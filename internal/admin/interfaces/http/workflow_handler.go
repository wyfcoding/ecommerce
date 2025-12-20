package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// WorkflowHandler 处理 HTTP 或 gRPC 请求。
type WorkflowHandler struct {
	workflowService *application.WorkflowService
	logger          *slog.Logger
}

// NewWorkflowHandler 处理 HTTP 或 gRPC 请求。
func NewWorkflowHandler(workflowService *application.WorkflowService, logger *slog.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		logger:          logger,
	}
}

func (h *WorkflowHandler) RegisterRoutes(r *gin.RouterGroup) {
	wf := r.Group("/workflow")
	{
		wf.POST("/apply", h.Apply)
		wf.POST("/:id/action", h.Action) // approve/reject
	}
}

func (h *WorkflowHandler) Apply(c *gin.Context) {
	var req application.ApprovalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get Requester ID from Auth Middleware
	requesterID := uint(1) // Mock

	domainReq := &domain.ApprovalRequest{
		RequesterID: requesterID,
		ActionType:  req.ActionType,
		Description: req.Description,
		Payload:     req.Payload,
	}

	if err := h.workflowService.CreateRequest(c.Request.Context(), domainReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": domainReq.ID})
}

func (h *WorkflowHandler) Action(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req application.ApprovalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get Approver ID from Auth Middleware
	approverID := uint(2) // Mock

	var err error
	if req.Action == "approve" {
		err = h.workflowService.ApproveRequest(c.Request.Context(), uint(id), approverID, req.Comment)
	} else {
		err = h.workflowService.RejectRequest(c.Request.Context(), uint(id), approverID, req.Comment)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
