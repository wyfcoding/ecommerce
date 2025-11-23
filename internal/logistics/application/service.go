package application

import (
	"context"
	"ecommerce/internal/logistics/domain/entity"
	"ecommerce/internal/logistics/domain/repository"
	"time"

	"log/slog"
)

type LogisticsService struct {
	repo   repository.LogisticsRepository
	logger *slog.Logger
}

func NewLogisticsService(repo repository.LogisticsRepository, logger *slog.Logger) *LogisticsService {
	return &LogisticsService{
		repo:   repo,
		logger: logger,
	}
}

// CreateLogistics 创建物流单
func (s *LogisticsService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string,
	receiverName, receiverPhone, receiverAddress string) (*entity.Logistics, error) {

	logistics := entity.NewLogistics(orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress,
		receiverName, receiverPhone, receiverAddress)

	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.Error("failed to save logistics", "error", err)
		return nil, err
	}
	return logistics, nil
}

// GetLogistics 获取物流信息
func (s *LogisticsService) GetLogistics(ctx context.Context, id uint64) (*entity.Logistics, error) {
	return s.repo.GetByID(ctx, id)
}

// GetLogisticsByTrackingNo 根据运单号获取物流信息
func (s *LogisticsService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error) {
	return s.repo.GetByTrackingNo(ctx, trackingNo)
}

// UpdateStatus 更新物流状态
func (s *LogisticsService) UpdateStatus(ctx context.Context, id uint64, status entity.LogisticsStatus, location, description string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	switch status {
	case entity.LogisticsStatusPickedUp:
		logistics.PickUp()
	case entity.LogisticsStatusInTransit:
		logistics.Transit(location)
	case entity.LogisticsStatusDelivering:
		logistics.Deliver()
	case entity.LogisticsStatusDelivered:
		logistics.Complete()
	case entity.LogisticsStatusReturning:
		logistics.Return()
	case entity.LogisticsStatusReturned:
		logistics.ReturnComplete()
	case entity.LogisticsStatusException:
		logistics.Exception(description)
	default:
		return entity.ErrInvalidStatus
	}

	logistics.AddTrace(location, description, "") // Status string could be mapped from enum if needed

	return s.repo.Save(ctx, logistics)
}

// AddTrace 添加物流轨迹
func (s *LogisticsService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.AddTrace(location, description, status)
	logistics.UpdateLocation(location)

	return s.repo.Save(ctx, logistics)
}

// SetEstimatedTime 设置预计送达时间
func (s *LogisticsService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.SetEstimatedTime(estimatedTime)
	return s.repo.Save(ctx, logistics)
}

// ListLogistics 获取物流列表
func (s *LogisticsService) ListLogistics(ctx context.Context, page, pageSize int) ([]*entity.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}
