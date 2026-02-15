package domain

import (
	"context"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/returns/v1"
)

// ReturnRequest 聚合根，代表一个退货申请.
type ReturnRequest struct {
	ID             string
	OrderID        string
	UserID         string
	Items          []ReturnItem
	Status         pb.ReturnStatus
	RMANumber      string
	TrackingNumber string
	WarehouseNotes string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ReturnItem 退货明细.
type ReturnItem struct {
	SKUID    string
	Quantity int32
	Reason   string
	Images   []string
}

// ReturnRepository 仓储接口.
type ReturnRepository interface {
	Save(ctx context.Context, req *ReturnRequest) error
	GetByID(ctx context.Context, id string) (*ReturnRequest, error)
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*ReturnRequest, int, error)
	GetByRMA(ctx context.Context, rma string) (*ReturnRequest, error)
}

// 领域方法
func (r *ReturnRequest) Approve(rma string) {
	r.Status = pb.ReturnStatus_RETURN_STATUS_APPROVED
	r.RMANumber = rma
	r.UpdatedAt = time.Now()
}

func (r *ReturnRequest) Ship(tracking string) {
	r.Status = pb.ReturnStatus_RETURN_STATUS_SHIPPED
	r.TrackingNumber = tracking
	r.UpdatedAt = time.Now()
}

func (r *ReturnRequest) Receive() {
	r.Status = pb.ReturnStatus_RETURN_STATUS_RECEIVED
	r.UpdatedAt = time.Now()
}

func (r *ReturnRequest) SetQCResult(passed bool, notes string) {
	if passed {
		r.Status = pb.ReturnStatus_RETURN_STATUS_QC_PASSED
	} else {
		r.Status = pb.ReturnStatus_RETURN_STATUS_QC_FAILED
	}
	r.WarehouseNotes = notes
	r.UpdatedAt = time.Now()
}

func (r *ReturnRequest) FinalizeRefund() {
	r.Status = pb.ReturnStatus_RETURN_STATUS_REFUNDED
	r.UpdatedAt = time.Now()
}

// Done
