package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// WorkflowService 处理审批流程
type WorkflowService struct {
	repo       domain.ApprovalRepository
	opsService *SystemOpsService // 用于执行通过后的逻辑
	audit      *AuditService
	logger     *slog.Logger
}

// NewWorkflowService 定义了 NewWorkflow 相关的服务逻辑。
func NewWorkflowService(repo domain.ApprovalRepository, ops *SystemOpsService, audit *AuditService, logger *slog.Logger) *WorkflowService {
	return &WorkflowService{
		repo:       repo,
		opsService: ops,
		audit:      audit,
		logger:     logger,
	}
}

// CreateRequest 提交新的审批申请
func (s *WorkflowService) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	req.Status = domain.ApprovalStatusPending
	req.CurrentStep = 1
	// 简单逻辑：假设所有流程都是1步审批，需要 "SUPER_ADMIN" 角色
	req.TotalSteps = 1
	req.ApproverRole = "SUPER_ADMIN"

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return err
	}

	s.audit.LogAction(ctx, &domain.AuditLog{
		UserID:   req.RequesterID,
		Action:   "workflow:create",
		Resource: "approval_request",
		TargetID: fmt.Sprintf("%d", req.ID),
		Status:   1,
	})
	return nil
}

// ApproveRequest 审批通过
func (s *WorkflowService) ApproveRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}

	// 记录审批Log
	logEntry := &domain.ApprovalLog{
		RequestID:  req.ID,
		ApproverID: approverID,
		Action:     domain.ApprovalActionApprove,
		Comment:    comment,
	}
	if err := s.repo.AddLog(ctx, logEntry); err != nil {
		return err
	}

	// 更新状态
	req.Status = domain.ApprovalStatusApproved
	now := time.Now()
	req.FinalizedAt = &now
	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return err
	}

	// 自动执行操作
	go func() {
		bgCtx := context.Background()
		if err := s.opsService.ExecuteOperation(bgCtx, req); err != nil {
			s.logger.Error("failed to execute operation after approval", "reqID", req.ID, "error", err)
			// TODO: 更新状态为"执行失败"？目前模型里没这个状态，暂记日志
		}
	}()

	return nil
}

// RejectRequest 审批拒绝
func (s *WorkflowService) RejectRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}

	// 记录日志
	logEntry := &domain.ApprovalLog{
		RequestID:  req.ID,
		ApproverID: approverID,
		Action:     domain.ApprovalActionReject,
		Comment:    comment,
	}
	if err := s.repo.AddLog(ctx, logEntry); err != nil {
		return err
	}

	// 更新状态
	req.Status = domain.ApprovalStatusRejected
	now := time.Now()
	req.FinalizedAt = &now
	return s.repo.UpdateRequest(ctx, req)
}
