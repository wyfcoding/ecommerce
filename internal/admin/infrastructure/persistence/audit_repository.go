package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 定义了数据持久层接口。
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) Find(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	if uid, ok := filter["user_id"]; ok {
		db = db.Where("user_id = ?", uid)
	}
	if action, ok := filter["action"]; ok {
		db = db.Where("action = ?", action)
	}
	if res, ok := filter["resource"]; ok {
		db = db.Where("resource = ?", res)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
