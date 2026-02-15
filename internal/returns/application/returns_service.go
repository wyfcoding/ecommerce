package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/returns/v1"
	"github.com/wyfcoding/ecommerce/internal/returns/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type ReturnService struct {
	repo   domain.ReturnRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

func NewReturnService(repo domain.ReturnRepository, idGen idgen.Generator, logger *slog.Logger) *ReturnService {
	return &ReturnService{
		repo:   repo,
		idGen:  idGen,
		logger: logger.With("service", "returns_application"),
	}
}

func (s *ReturnService) CreateRequest(ctx context.Context, userID, orderID string, items []domain.ReturnItem) (*domain.ReturnRequest, error) {
	// TODO: 校验订单是否属于用户且可退货（调用 order gRPC）

	req := &domain.ReturnRequest{
		ID:        fmt.Sprintf("ret_%d", s.idGen.Generate()),
		OrderID:   orderID,
		UserID:    userID,
		Items:     items,
		Status:    pb.ReturnStatus_RETURN_STATUS_PENDING_REVIEW,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ReturnService) ApproveRequest(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rma := fmt.Sprintf("RMA-%d", s.idGen.Generate())
	req.Approve(rma)

	if err := s.repo.Save(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ReturnService) ReceiveItem(ctx context.Context, rma string) (*domain.ReturnRequest, error) {
	req, err := s.repo.GetByRMA(ctx, rma)
	if err != nil {
		return nil, err
	}

	req.Receive()

	if err := s.repo.Save(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ReturnService) LogQC(ctx context.Context, id string, passed bool, notes string) (*domain.ReturnRequest, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	req.SetQCResult(passed, notes)

	if err := s.repo.Save(ctx, req); err != nil {
		return nil, err
	}

	// 触发库存逻辑（若 passed 为 true，调用 inventory 增加 stock）
	// TODO: 发送 ReturnQCResultEvent 到 Kafka

	return req, nil
}

func (s *ReturnService) ListMyReturns(ctx context.Context, userID string, page, pageSize int32) ([]*domain.ReturnRequest, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByUserID(ctx, userID, int(offset), int(pageSize))
}

func (s *ReturnService) GetDetail(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	return s.repo.GetByID(ctx, id)
}
