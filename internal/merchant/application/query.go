// Package application 商家服务查询服务
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
)

// QueryService 商家查询服务
type QueryService struct {
	merchantRepo    domain.MerchantRepository
	licenseRepo     domain.BusinessLicenseRepository
	bankAccountRepo domain.BankAccountRepository
	storeRepo       domain.StoreRepository
	settingsRepo    domain.MerchantSettingsRepository
	logger          *slog.Logger
}

// NewQueryService 创建查询服务
func NewQueryService(
	merchantRepo domain.MerchantRepository,
	licenseRepo domain.BusinessLicenseRepository,
	bankAccountRepo domain.BankAccountRepository,
	storeRepo domain.StoreRepository,
	settingsRepo domain.MerchantSettingsRepository,
	logger *slog.Logger,
) *QueryService {
	return &QueryService{
		merchantRepo:    merchantRepo,
		licenseRepo:     licenseRepo,
		bankAccountRepo: bankAccountRepo,
		storeRepo:       storeRepo,
		settingsRepo:    settingsRepo,
		logger:          logger,
	}
}

// MerchantDTO 商家查询结果DTO
type MerchantDTO struct {
	ID              uint
	UserID          uint64
	MerchantNo      string
	Name            string
	LegalName       string
	ContactName     string
	ContactPhone    string
	ContactEmail    string
	Type            domain.MerchantType
	Status          domain.MerchantStatus
	Level           domain.MerchantLevel
	LogoURL         string
	Description     string
	CommissionRate  float64
	CreditScore     float64
	TotalSales      int64
	TotalOrders     int32
	Rating          float64
	RejectReason    string
	BusinessLicense *domain.BusinessLicense
	BankAccount     *domain.BankAccount
}

// StoreDTO 店铺查询结果DTO
type StoreDTO struct {
	ID            uint
	MerchantID    uint
	StoreNo       string
	Name          string
	LogoURL       string
	BannerURL     string
	Description   string
	Announcement  string
	Categories    string
	Rating        float64
	ProductCount  int32
	MonthlySales  int64
	IsOpen        bool
	BusinessHours string
	Address       string
}

// ListResult 列表查询结果
type ListResult[T any] struct {
	Items []T
	Total int64
	Page  int
	Size  int
}

// GetMerchant 获取商家信息
func (s *QueryService) GetMerchant(ctx context.Context, merchantID uint) (*MerchantDTO, error) {
	merchant, err := s.merchantRepo.FindByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	return s.toMerchantDTO(ctx, merchant), nil
}

// GetMerchantByUserID 根据用户ID获取商家
func (s *QueryService) GetMerchantByUserID(ctx context.Context, userID uint64) (*MerchantDTO, error) {
	merchant, err := s.merchantRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.toMerchantDTO(ctx, merchant), nil
}

// ListMerchants 列出商家
func (s *QueryService) ListMerchants(ctx context.Context, filter *domain.MerchantFilter) (*ListResult[*MerchantDTO], error) {
	merchants, total, err := s.merchantRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]*MerchantDTO, 0, len(merchants))
	for _, m := range merchants {
		items = append(items, s.toMerchantDTO(ctx, m))
	}

	return &ListResult[*MerchantDTO]{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Size:  filter.PageSize,
	}, nil
}

// GetSettings 获取商家设置
func (s *QueryService) GetSettings(ctx context.Context, merchantID uint) (*domain.MerchantSettings, error) {
	return s.settingsRepo.FindByMerchantID(ctx, merchantID)
}

// GetStore 获取店铺
func (s *QueryService) GetStore(ctx context.Context, storeID uint) (*StoreDTO, error) {
	store, err := s.storeRepo.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}

	return s.toStoreDTO(store), nil
}

// ListStores 列出店铺
func (s *QueryService) ListStores(ctx context.Context, merchantID uint, page, pageSize int) (*ListResult[*StoreDTO], error) {
	stores, total, err := s.storeRepo.ListByMerchantID(ctx, merchantID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*StoreDTO, 0, len(stores))
	for _, store := range stores {
		items = append(items, s.toStoreDTO(store))
	}

	return &ListResult[*StoreDTO]{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// toMerchantDTO 转换为商家DTO
func (s *QueryService) toMerchantDTO(ctx context.Context, m *domain.Merchant) *MerchantDTO {
	dto := &MerchantDTO{
		ID:             m.ID,
		UserID:         m.UserID,
		MerchantNo:     m.MerchantNo,
		Name:           m.Name,
		LegalName:      m.LegalName,
		ContactName:    m.ContactName,
		ContactPhone:   m.ContactPhone,
		ContactEmail:   m.ContactEmail,
		Type:           m.Type,
		Status:         m.Status,
		Level:          m.Level,
		LogoURL:        m.LogoURL,
		Description:    m.Description,
		CommissionRate: m.CommissionRate,
		CreditScore:    m.CreditScore,
		TotalSales:     m.TotalSales,
		TotalOrders:    m.TotalOrders,
		Rating:         m.Rating,
		RejectReason:   m.RejectReason,
	}

	// 加载关联数据
	if license, err := s.licenseRepo.FindByMerchantID(ctx, m.ID); err == nil {
		dto.BusinessLicense = license
	}
	if account, err := s.bankAccountRepo.FindByMerchantID(ctx, m.ID); err == nil {
		dto.BankAccount = account
	}

	return dto
}

// toStoreDTO 转换为店铺DTO
func (s *QueryService) toStoreDTO(store *domain.Store) *StoreDTO {
	return &StoreDTO{
		ID:            store.ID,
		MerchantID:    store.MerchantID,
		StoreNo:       store.StoreNo,
		Name:          store.Name,
		LogoURL:       store.LogoURL,
		BannerURL:     store.BannerURL,
		Description:   store.Description,
		Announcement:  store.Announcement,
		Categories:    store.Categories,
		Rating:        store.Rating,
		ProductCount:  store.ProductCount,
		MonthlySales:  store.MonthlySales,
		IsOpen:        store.IsOpen,
		BusinessHours: store.BusinessHours,
		Address:       store.Address,
	}
}
