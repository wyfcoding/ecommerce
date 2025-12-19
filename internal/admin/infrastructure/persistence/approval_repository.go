package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type approvalRepository struct {
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) domain.ApprovalRepository {
	return &approvalRepository{db: db}
}

func (r *approvalRepository) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *approvalRepository) GetRequestByID(ctx context.Context, id uint) (*domain.ApprovalRequest, error) {
	var req domain.ApprovalRequest
	if err := r.db.WithContext(ctx).Preload("Logs").First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *approvalRepository) UpdateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *approvalRepository) ListPendingRequests(ctx context.Context, roleLimit string) ([]*domain.ApprovalRequest, error) {
	var reqs []*domain.ApprovalRequest
	db := r.db.WithContext(ctx).Where("status = ?", domain.ApprovalStatusPending)

	if roleLimit != "" {
		db = db.Where("approver_role = ?", roleLimit)
	}

	if err := db.Order("created_at asc").Find(&reqs).Error; err != nil {
		return nil, err
	}
	return reqs, nil
}

func (r *approvalRepository) AddLog(ctx context.Context, log *domain.ApprovalLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
