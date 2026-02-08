// Package application 履约服务查询服务
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
)

// QueryService 履约查询服务
type QueryService struct {
	fulfillmentRepo domain.FulfillmentRepository
	logger          *slog.Logger
}

// NewQueryService 创建查询服务
func NewQueryService(
	fulfillmentRepo domain.FulfillmentRepository,
	logger *slog.Logger,
) *QueryService {
	return &QueryService{
		fulfillmentRepo: fulfillmentRepo,
		logger:          logger,
	}
}

// FulfillmentDTO 履约单查询结果 DTO
type FulfillmentDTO struct {
	ID            uint
	FulfillmentNo string
	OrderNo       string
	MerchantID    uint64
	StoreID       uint64
	WarehouseID   uint64
	Type          domain.FulfillmentType
	Status        domain.FulfillmentStatus

	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	District      string
	Address       string

	PickerID   uint64
	PickerName string
	PackerID   uint64
	PackerName string

	CarrierCode string
	CarrierName string
	TrackingNo  string
	ShippingFee int64

	Items    []domain.FulfillmentItem
	Packages []domain.Package
}

// ListResult 列表结果
type ListResult struct {
	Items []FulfillmentDTO
	Total int64
	Page  int
	Size  int
}

// GetFulfillment 获取履约单详情
func (s *QueryService) GetFulfillment(ctx context.Context, id uint) (*FulfillmentDTO, error) {
	f, err := s.fulfillmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toDTO(f), nil
}

// GetByFulfillmentNo 根据履约单号获取
func (s *QueryService) GetByFulfillmentNo(ctx context.Context, no string) (*FulfillmentDTO, error) {
	f, err := s.fulfillmentRepo.FindByFulfillmentNo(ctx, no)
	if err != nil {
		return nil, err
	}
	return s.toDTO(f), nil
}

// ListByOrderNo 根据订单号列出履约单
func (s *QueryService) ListByOrderNo(ctx context.Context, orderNo string) ([]FulfillmentDTO, error) {
	fulfillments, err := s.fulfillmentRepo.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	result := make([]FulfillmentDTO, 0, len(fulfillments))
	for _, f := range fulfillments {
		result = append(result, *s.toDTO(f))
	}
	return result, nil
}

// ListFulfillments 列出履约单
func (s *QueryService) ListFulfillments(ctx context.Context, filter *domain.FulfillmentFilter) (*ListResult, error) {
	fulfillments, total, err := s.fulfillmentRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]FulfillmentDTO, 0, len(fulfillments))
	for _, f := range fulfillments {
		items = append(items, *s.toDTO(f))
	}

	return &ListResult{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Size:  filter.PageSize,
	}, nil
}

// toDTO 转换为 DTO
func (s *QueryService) toDTO(f *domain.Fulfillment) *FulfillmentDTO {
	return &FulfillmentDTO{
		ID:            f.ID,
		FulfillmentNo: f.FulfillmentNo,
		OrderNo:       f.OrderNo,
		MerchantID:    f.MerchantID,
		StoreID:       f.StoreID,
		WarehouseID:   f.WarehouseID,
		Type:          f.Type,
		Status:        f.Status,
		ReceiverName:  f.ReceiverName,
		ReceiverPhone: f.ReceiverPhone,
		Province:      f.Province,
		City:          f.City,
		District:      f.District,
		Address:       f.Address,
		PickerID:      f.PickerID,
		PickerName:    f.PickerName,
		PackerID:      f.PackerID,
		PackerName:    f.PackerName,
		CarrierCode:   f.CarrierCode,
		CarrierName:   f.CarrierName,
		TrackingNo:    f.TrackingNo,
		ShippingFee:   f.ShippingFee,
		Items:         f.Items,
		Packages:      f.Packages,
	}
}
