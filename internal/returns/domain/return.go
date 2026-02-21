package domain

import (
	"errors"

	"gorm.io/gorm"
	"github.com/wyfcoding/pkg/database"
)

type ReturnStatus string

const (
	ReturnCreated   ReturnStatus = "CREATED"
	ReturnApproved  ReturnStatus = "APPROVED"
	ReturnReceived  ReturnStatus = "RECEIVED"
	ReturnInspected ReturnStatus = "INSPECTED"
	ReturnRefunded  ReturnStatus = "REFUNDED"
	ReturnRejected  ReturnStatus = "REJECTED"
)

// ReturnRequest 退货申请聚合根。
type ReturnRequest struct {
	gorm.Model
	database.BaseEntity
	OrderID  string       `gorm:"column:order_id;index;not null"`
	ItemID   string       `gorm:"column:item_id;not null"`
	Reason   string       `gorm:"column:reason;type:text"`
	Status   ReturnStatus `gorm:"column:status;default:'CREATED'"`
}

// Approve 审核通过。
func (r *ReturnRequest) Approve() error {
	if r.Status != ReturnCreated {
		return errors.New("invalid status transition")
	}
	r.Status = ReturnApproved
	return nil
}

// MarkReceived 确认收到货物。
func (r *ReturnRequest) MarkReceived() error {
	if r.Status != ReturnApproved {
		return errors.New("invalid status transition")
	}
	r.Status = ReturnReceived
	return nil
}

type Repository interface {
	Save(r *ReturnRequest) error
	FindByID(id uint) (*ReturnRequest, error)
}
