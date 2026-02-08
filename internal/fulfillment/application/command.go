// Package application 履约服务应用层
// 生成摘要：
// 1) 实现 CQRS 模式的命令服务
// 2) 处理完整履约流程：创建→拣货→打包→发货
package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// CommandService 履约命令服务
type CommandService struct {
	fulfillmentRepo domain.FulfillmentRepository
	eventPublisher  messagequeue.EventPublisher
	logger          *slog.Logger
}

// NewCommandService 创建命令服务
func NewCommandService(
	fulfillmentRepo domain.FulfillmentRepository,
	eventPublisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *CommandService {
	return &CommandService{
		fulfillmentRepo: fulfillmentRepo,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

// CreateFulfillmentCommand 创建履约单命令
type CreateFulfillmentCommand struct {
	OrderNo          string
	MerchantID       uint64
	StoreID          uint64
	WarehouseID      uint64
	Type             domain.FulfillmentType
	ExpectedShipTime *time.Time
	Remark           string
	Address          ShippingAddress
	Items            []FulfillmentItemDTO
}

// ShippingAddress 收货地址 DTO
type ShippingAddress struct {
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Address       string
	PostalCode    string
}

// FulfillmentItemDTO 履约商品项 DTO
type FulfillmentItemDTO struct {
	SKUID       string
	ProductName string
	SKUName     string
	ImageURL    string
	Quantity    int32
	Location    string
	BatchNo     string
}

// CreateFulfillmentResult 创建履约单结果
type CreateFulfillmentResult struct {
	FulfillmentID uint
	FulfillmentNo string
}

// CreateFulfillment 创建履约单
func (s *CommandService) CreateFulfillment(ctx context.Context, cmd CreateFulfillmentCommand) (*CreateFulfillmentResult, error) {
	start := time.Now()

	fulfillment := domain.NewFulfillment(
		cmd.OrderNo,
		cmd.MerchantID,
		cmd.StoreID,
		cmd.WarehouseID,
		cmd.Type,
	)

	fulfillment.SetShippingAddress(
		cmd.Address.ReceiverName,
		cmd.Address.ReceiverPhone,
		cmd.Address.Province,
		cmd.Address.City,
		cmd.Address.District,
		cmd.Address.Address,
		cmd.Address.PostalCode,
	)

	fulfillment.ExpectedShipTime = cmd.ExpectedShipTime
	fulfillment.Remark = cmd.Remark

	for _, item := range cmd.Items {
		fulfillment.AddItem(
			item.SKUID,
			item.ProductName,
			item.SKUName,
			item.ImageURL,
			item.Location,
			item.BatchNo,
			item.Quantity,
		)
	}

	if err := s.fulfillmentRepo.Save(ctx, fulfillment); err != nil {
		s.logger.ErrorContext(ctx, "failed to save fulfillment",
			"order_no", cmd.OrderNo,
			"error", err,
			"duration", time.Since(start))
		return nil, err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "fulfillment created",
		"fulfillment_id", fulfillment.ID,
		"fulfillment_no", fulfillment.FulfillmentNo,
		"order_no", cmd.OrderNo,
		"duration", time.Since(start))

	return &CreateFulfillmentResult{
		FulfillmentID: fulfillment.ID,
		FulfillmentNo: fulfillment.FulfillmentNo,
	}, nil
}

// AssignPickingCommand 分配拣货命令
type AssignPickingCommand struct {
	FulfillmentID uint
	PickerID      uint64
	PickerName    string
}

// AssignPicking 分配拣货员
func (s *CommandService) AssignPicking(ctx context.Context, cmd AssignPickingCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.AssignPicker(cmd.PickerID, cmd.PickerName); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "picker assigned",
		"fulfillment_id", cmd.FulfillmentID,
		"picker_id", cmd.PickerID)

	return nil
}

// StartPickingCommand 开始拣货命令
type StartPickingCommand struct {
	FulfillmentID uint
	PickerID      uint64
}

// StartPicking 开始拣货
func (s *CommandService) StartPicking(ctx context.Context, cmd StartPickingCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.StartPicking(cmd.PickerID); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "picking started",
		"fulfillment_id", cmd.FulfillmentID)

	return nil
}

// CompletePickingCommand 完成拣货命令
type CompletePickingCommand struct {
	FulfillmentID uint
	Items         map[string]int32 // SKU ID -> 已拣数量
}

// CompletePicking 完成拣货
func (s *CommandService) CompletePicking(ctx context.Context, cmd CompletePickingCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.CompletePicking(cmd.Items); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "picking completed",
		"fulfillment_id", cmd.FulfillmentID)

	return nil
}

// StartPackingCommand 开始打包命令
type StartPackingCommand struct {
	FulfillmentID uint
	PackerID      uint64
	PackerName    string
}

// StartPacking 开始打包
func (s *CommandService) StartPacking(ctx context.Context, cmd StartPackingCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.StartPacking(cmd.PackerID, cmd.PackerName); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "packing started",
		"fulfillment_id", cmd.FulfillmentID)

	return nil
}

// CompletePackingCommand 完成打包命令
type CompletePackingCommand struct {
	FulfillmentID uint
	Packages      []PackageDTO
}

// PackageDTO 包裹 DTO
type PackageDTO struct {
	Weight float64
	Length float64
	Width  float64
	Height float64
	SKUIDs []string
}

// CompletePacking 完成打包
func (s *CommandService) CompletePacking(ctx context.Context, cmd CompletePackingCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	packages := make([]domain.Package, 0, len(cmd.Packages))
	for i, p := range cmd.Packages {
		packages = append(packages, domain.Package{
			PackageNo: fulfillment.FulfillmentNo + "-P" + string(rune('1'+i)),
			Weight:    p.Weight,
			Length:    p.Length,
			Width:     p.Width,
			Height:    p.Height,
		})
	}

	if err := fulfillment.CompletePacking(packages); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "packing completed",
		"fulfillment_id", cmd.FulfillmentID,
		"package_count", len(packages))

	return nil
}

// ArrangeShipmentCommand 安排发货命令
type ArrangeShipmentCommand struct {
	FulfillmentID uint
	CarrierCode   string
	CarrierName   string
	ShippingFee   int64
}

// ArrangeShipment 安排发货
func (s *CommandService) ArrangeShipment(ctx context.Context, cmd ArrangeShipmentCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.ArrangeShipment(cmd.CarrierCode, cmd.CarrierName, cmd.ShippingFee); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "shipment arranged",
		"fulfillment_id", cmd.FulfillmentID,
		"carrier_code", cmd.CarrierCode)

	return nil
}

// ConfirmShipmentCommand 确认发货命令
type ConfirmShipmentCommand struct {
	FulfillmentID uint
	TrackingNo    string
}

// ConfirmShipment 确认发货
func (s *CommandService) ConfirmShipment(ctx context.Context, cmd ConfirmShipmentCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.ConfirmShipment(cmd.TrackingNo); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "shipment confirmed",
		"fulfillment_id", cmd.FulfillmentID,
		"tracking_no", cmd.TrackingNo)

	return nil
}

// CancelFulfillmentCommand 取消履约命令
type CancelFulfillmentCommand struct {
	FulfillmentID uint
	Reason        string
	Operator      string
}

// CancelFulfillment 取消履约单
func (s *CommandService) CancelFulfillment(ctx context.Context, cmd CancelFulfillmentCommand) error {
	fulfillment, err := s.fulfillmentRepo.FindByID(ctx, cmd.FulfillmentID)
	if err != nil {
		return err
	}

	if err := fulfillment.Cancel(cmd.Reason, cmd.Operator); err != nil {
		return err
	}

	if err := s.fulfillmentRepo.Update(ctx, fulfillment); err != nil {
		return err
	}

	s.publishEvents(ctx, fulfillment.GetDomainEvents())
	fulfillment.ClearDomainEvents()

	s.logger.InfoContext(ctx, "fulfillment cancelled",
		"fulfillment_id", cmd.FulfillmentID,
		"reason", cmd.Reason)

	return nil
}

// publishEvents 发布领域事件
func (s *CommandService) publishEvents(ctx context.Context, events []domain.DomainEvent) {
	for _, event := range events {
		if err := s.eventPublisher.Publish(ctx, event.EventName(), "", event); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish event",
				"event", event.EventName(),
				"error", err)
		}
	}
}
