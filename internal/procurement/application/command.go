package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/procurement/domain"
)

type ProcurementCommandService struct {
	repo domain.ProcurementRepository
}

func NewProcurementCommandService(repo domain.ProcurementRepository) *ProcurementCommandService {
	return &ProcurementCommandService{repo: repo}
}

func (s *ProcurementCommandService) CreatePurchaseRequest(ctx context.Context, applicantID, reason string, items []domain.PurchaseRequestItem) (string, error) {
	id := fmt.Sprintf("PR%d", time.Now().UnixNano())
	pr := &domain.PurchaseRequest{
		RequestID:   id,
		ApplicantID: applicantID,
		Reason:      reason,
		Status:      domain.PRStatusPending,
		Items:       items,
	}
	// fix item references
	for i := range pr.Items {
		pr.Items[i].RequestID = id
	}

	if err := s.repo.SavePurchaseRequest(ctx, pr); err != nil {
		return "", err
	}
	return id, nil
}

func (s *ProcurementCommandService) ApprovePurchaseRequest(ctx context.Context, requestID, approverID string, approved bool, comment string) error {
	pr, err := s.repo.GetPurchaseRequest(ctx, requestID)
	if err != nil {
		return err
	}

	if approved {
		pr.Status = domain.PRStatusApproved
		now := time.Now()
		pr.ApprovedAt = &now
	} else {
		pr.Status = domain.PRStatusRejected
	}
	pr.ApproverID = approverID
	pr.Comment = comment

	return s.repo.SavePurchaseRequest(ctx, pr)
}

func (s *ProcurementCommandService) CreatePurchaseOrder(ctx context.Context, reqID, supplierID, warehouseID, remark string, items []struct {
	SKUID string
	Name  string
	Qty   int32
	Price float64
}) (string, error) {
	id := fmt.Sprintf("PO%d", time.Now().UnixNano())
	po := domain.NewPurchaseOrder(id, reqID, supplierID, warehouseID, remark)

	for _, item := range items {
		po.AddItem(item.SKUID, item.Name, item.Qty, decimal.NewFromFloat(item.Price))
	}

	if err := s.repo.SavePurchaseOrder(ctx, po); err != nil {
		return "", err
	}

	// specialized logic: if linked to PR, update PR status
	if reqID != "" {
		pr, err := s.repo.GetPurchaseRequest(ctx, reqID)
		if err == nil {
			pr.Status = domain.PRStatusOrdered
			_ = s.repo.SavePurchaseRequest(ctx, pr)
		}
	}

	return id, nil
}

func (s *ProcurementCommandService) UpdatePurchaseOrderStatus(ctx context.Context, orderID, status string) error {
	po, err := s.repo.GetPurchaseOrder(ctx, orderID)
	if err != nil {
		return err
	}

	switch status {
	case "CONFIRMED":
		_ = po.Confirm()
	case "SHIPPED":
		_ = po.Ship()
	case "RECEIVED":
		_ = po.Receive()
	case "COMPLETED":
		po.Status = domain.POStatusCompleted
	case "CANCELLED":
		po.Status = domain.POStatusCancelled
	}

	return s.repo.SavePurchaseOrder(ctx, po)
}
