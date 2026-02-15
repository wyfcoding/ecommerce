// Package application 退款服务应用层
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/refund/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

type RefundService struct {
	repo     domain.RefundRepository
	eventBus messagequeue.EventBus
	logger   *slog.Logger
}

func NewRefundService(
	repo domain.RefundRepository,
	eventBus messagequeue.EventBus,
	logger *slog.Logger,
) *RefundService {
	return &RefundService{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger.With("module", "refund_service"),
	}
}

// CreateRefundCmd 创建退款申请参数
type CreateRefundCmd struct {
	OrderID       string              `json:"order_id"`
	OrderNo       string              `json:"order_no"`
	UserID        uint64              `json:"user_id"`
	MerchantID    uint64              `json:"merchant_id"`
	Amount        int64               `json:"amount"`
	Reason        string              `json:"reason"`
	Description   string              `json:"description"`
	Type          domain.RefundType   `json:"type"`
	PaymentID     string              `json:"payment_id"`
	TransactionID string              `json:"transaction_id"`
	Items         []CreateRefundItem  `json:"items"`
}

type CreateRefundItem struct {
	OrderItemID  string `json:"order_item_id"`
	SkuID        uint64 `json:"sku_id"`
	Quantity     int32  `json:"quantity"`
	RefundAmount int64  `json:"refund_amount"`
}

// CreateRefund 提交退款申请
func (s *RefundService) CreateRefund(ctx context.Context, cmd CreateRefundCmd) (string, error) {
	// 生成退款单号
	refundNo := fmt.Sprintf("REF%s", idgen.GenIDString())
	
	refund := domain.NewRefundRequest(
		refundNo,
		cmd.OrderID,
		cmd.OrderNo,
		cmd.UserID,
		cmd.MerchantID,
		cmd.Amount,
		cmd.Reason,
		cmd.Description,
		cmd.Type,
		cmd.PaymentID,
		cmd.TransactionID,
	)
	
	for _, item := range cmd.Items {
		refund.Items = append(refund.Items, domain.RefundItem{
			OrderItemID:  item.OrderItemID,
			SkuID:        item.SkuID,
			Quantity:     item.Quantity,
			RefundAmount: item.RefundAmount,
		})
	}
	
	// 提交商家审核
	refund.SubmitToMerchant()
	
	if err := s.repo.Save(ctx, refund); err != nil {
		return "", err
	}
	
	// 发布事件
	_ = s.eventBus.Publish(ctx, &domain.RefundCreatedEvent{
		RefundID:  refund.ID,
		RefundNo:  refund.RefundNo,
		OrderID:   refund.OrderID,
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		CreatedAt: refund.CreatedAt,
	})
	
	return refundNo, nil
}

// MerchantApprove 商家审核同意
func (s *RefundService) MerchantApprove(ctx context.Context, refundID uint64, operatorID uint64) error {
	refund, err := s.repo.GetByID(ctx, refundID)
	if err != nil {
		return err
	}
	if refund == nil {
		return domain.ErrRefundNotFound
	}
	
	// TODO: 校验 operatorID 是否属于 MerchantID
	
	refund.MerchantApprove(operatorID)
	
	if err := s.repo.Save(ctx, refund); err != nil {
		return err
	}
	
	// 如果审核通过（目前逻辑直接进入待打款），触发打款事件
	if refund.Status == domain.RefundStatusApproved {
		_ = s.eventBus.Publish(ctx, &domain.RefundApprovedEvent{
			RefundID:   refund.ID,
			RefundNo:   refund.RefundNo,
			PaymentID:  refund.PaymentID,
			Amount:     refund.Amount,
			ApprovedAt: time.Now(),
		})
	}
	
	return nil
}

// MerchantReject 商家审核拒绝
func (s *RefundService) MerchantReject(ctx context.Context, refundID uint64, operatorID uint64, reason string) error {
	refund, err := s.repo.GetByID(ctx, refundID)
	if err != nil {
		return err
	}
	if refund == nil {
		return domain.ErrRefundNotFound
	}
	
	refund.MerchantReject(operatorID, reason)
	
	if err := s.repo.Save(ctx, refund); err != nil {
		return err
	}
	
	// 发布拒绝事件
	_ = s.eventBus.Publish(ctx, &domain.RefundFailedEvent{
		RefundID: refund.ID,
		RefundNo: refund.RefundNo,
		Reason:   reason,
		FailedAt: time.Now(),
	})
	
	return nil
}

// ProcessRefundCallback 处理退款结果回调 (来自 Payment 服务或银行)
func (s *RefundService) ProcessRefundCallback(ctx context.Context, refundNo string, success bool, reason string) error {
	refund, err := s.repo.GetByRefundNo(ctx, refundNo)
	if err != nil {
		return err
	}
	if refund == nil {
		return domain.ErrRefundNotFound
	}
	
	if success {
		refund.SystemSucceed()
		_ = s.eventBus.Publish(ctx, &domain.RefundSucceededEvent{
			RefundID:    refund.ID,
			RefundNo:    refund.RefundNo,
			OrderID:     refund.OrderID,
			Amount:      refund.Amount,
			SucceededAt: time.Now(),
		})
	} else {
		refund.SystemFail(reason)
		_ = s.eventBus.Publish(ctx, &domain.RefundFailedEvent{
			RefundID: refund.ID,
			RefundNo: refund.RefundNo,
			Reason:   reason,
			FailedAt: time.Now(),
		})
	}
	
	return s.repo.Save(ctx, refund)
}
