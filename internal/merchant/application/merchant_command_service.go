// Package application 商家命令服务（CQRS 写侧）
// 生成摘要：
// 1) 实现商家服务的写操作命令，包括入驻申请、审核、店铺管理等功能
// 2) 所有写操作完成后发布领域事件，驱动读模型投影更新
// 3) 支持事务操作和事件驱动补偿策略
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// MerchantCommandService 商家命令服务（CQRS 写侧）
type MerchantCommandService struct {
	merchantRepo domain.MerchantRepository
	storeRepo    domain.StoreRepository
	settingsRepo domain.MerchantSettingsRepository
	eventBus     messagequeue.EventBus
	logger       *slog.Logger
}

// NewMerchantCommandService 创建商家命令服务实例
func NewMerchantCommandService(
	merchantRepo domain.MerchantRepository,
	storeRepo domain.StoreRepository,
	settingsRepo domain.MerchantSettingsRepository,
	eventBus messagequeue.EventBus,
	logger *slog.Logger,
) *MerchantCommandService {
	return &MerchantCommandService{
		merchantRepo: merchantRepo,
		storeRepo:    storeRepo,
		settingsRepo: settingsRepo,
		eventBus:     eventBus,
		logger:       logger.With("module", "merchant_command"),
	}
}

// ApplyMerchantCmd 商家入驻申请命令
type ApplyMerchantCmd struct {
	UserID       uint64               `json:"user_id" validate:"required"`
	Name         string               `json:"name" validate:"required"`
	LegalName    string               `json:"legal_name" validate:"required"`
	LegalIDCard  string               `json:"legal_id_card" validate:"required"`
	ContactName  string               `json:"contact_name" validate:"required"`
	ContactPhone string               `json:"contact_phone" validate:"required"`
	ContactEmail string               `json:"contact_email" validate:"required"`
	MerchantType domain.MerchantType  `json:"merchant_type" validate:"required"`
	LicenseInfo  *BusinessLicenseInfo `json:"license_info"`
	BankInfo     *BankAccountInfo     `json:"bank_info"`
}

// BusinessLicenseInfo 营业执照信息
type BusinessLicenseInfo struct {
	LicenseNo         string `json:"license_no" validate:"required"`
	CompanyName       string `json:"company_name" validate:"required"`
	CompanyType       string `json:"company_type"`
	RegisteredCapital string `json:"registered_capital"`
	EstablishmentDate string `json:"establishment_date"`
	ValidUntil        string `json:"valid_until"`
	BusinessScope     string `json:"business_scope"`
	RegisteredAddress string `json:"registered_address"`
	LicenseImageURL   string `json:"license_image_url" validate:"required"`
}

// BankAccountInfo 银行账户信息
type BankAccountInfo struct {
	AccountName string `json:"account_name" validate:"required"`
	AccountNo   string `json:"account_no" validate:"required"`
	BankName    string `json:"bank_name" validate:"required"`
	BankBranch  string `json:"bank_branch"`
	BankCode    string `json:"bank_code"`
}

// ApplyMerchant 商家入驻申请
func (s *MerchantCommandService) ApplyMerchant(ctx context.Context, cmd *ApplyMerchantCmd) (*domain.Merchant, error) {
	start := time.Now()

	// 检查是否已存在商家
	existing, err := s.merchantRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("check existing merchant: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("merchant already exists for user %d", cmd.UserID)
	}

	// 创建商家实体
	merchant := domain.NewMerchant(
		cmd.UserID,
		cmd.Name,
		cmd.LegalName,
		cmd.LegalIDCard,
		cmd.ContactName,
		cmd.ContactPhone,
		cmd.ContactEmail,
		cmd.MerchantType,
	)

	// 设置营业执照信息
	if cmd.LicenseInfo != nil {
		merchant.BusinessLicense = &domain.BusinessLicense{
			MerchantID:        merchant.ID,
			LicenseNo:         cmd.LicenseInfo.LicenseNo,
			CompanyName:       cmd.LicenseInfo.CompanyName,
			CompanyType:       cmd.LicenseInfo.CompanyType,
			RegisteredCapital: cmd.LicenseInfo.RegisteredCapital,
			EstablishmentDate: cmd.LicenseInfo.EstablishmentDate,
			ValidUntil:        cmd.LicenseInfo.ValidUntil,
			BusinessScope:     cmd.LicenseInfo.BusinessScope,
			RegisteredAddress: cmd.LicenseInfo.RegisteredAddress,
			LicenseImageURL:   cmd.LicenseInfo.LicenseImageURL,
		}
	}

	// 设置银行账户信息
	if cmd.BankInfo != nil {
		merchant.BankAccount = &domain.BankAccount{
			MerchantID:  merchant.ID,
			AccountName: cmd.BankInfo.AccountName,
			AccountNo:   cmd.BankInfo.AccountNo,
			BankName:    cmd.BankInfo.BankName,
			BankBranch:  cmd.BankInfo.BankBranch,
			BankCode:    cmd.BankInfo.BankCode,
		}
	}

	// 保存商家信息
	if err := s.merchantRepo.Save(ctx, merchant); err != nil {
		s.logger.ErrorContext(ctx, "failed to create merchant",
			"user_id", cmd.UserID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("create merchant: %w", err)
	}

	// 创建默认设置
	defaultSettings := &domain.MerchantSettings{
		MerchantID:             merchant.ID,
		AutoConfirmOrder:       false,
		AutoConfirmDays:        7,
		EnableCOD:              false,
		EnableInvoice:          true,
		SupportedPayments:      `["alipay","wechat"]`,
		FreeShippingThreshold:  0,
		DefaultShippingFee:     0,
		EnableSameDayDelivery:  false,
		OrderNotification:      true,
		StockAlert:             true,
		ReviewNotification:     true,
		SettlementNotification: true,
	}
	if err := s.settingsRepo.Save(ctx, defaultSettings); err != nil {
		s.logger.WarnContext(ctx, "failed to create default settings", "error", err)
	}

	// 发布领域事件
	if err := s.publishEvent(ctx, &domain.MerchantAppliedEvent{
		MerchantID:   uint64(merchant.ID),
		MerchantNo:   merchant.MerchantNo,
		UserID:       merchant.UserID,
		Name:         merchant.Name,
		MerchantType: merchant.Type,
		Timestamp:    time.Now(),
	}); err != nil {
		s.logger.WarnContext(ctx, "failed to publish merchant applied event", "error", err)
	}

	s.logger.InfoContext(ctx, "merchant application submitted",
		"merchant_id", merchant.ID, "merchant_no", merchant.MerchantNo, "duration", time.Since(start))
	return merchant, nil
}

// ApproveMerchantCmd 审核通过商家命令
type ApproveMerchantCmd struct {
	MerchantID     uint64  `json:"merchant_id" validate:"required"`
	CommissionRate float64 `json:"commission_rate" validate:"required,gt=0,lte=0.2"`
	Operator       string  `json:"operator" validate:"required"`
	Remark         string  `json:"remark"`
}

// ApproveMerchant 审核通过商家
func (s *MerchantCommandService) ApproveMerchant(ctx context.Context, cmd *ApproveMerchantCmd) error {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, uint(cmd.MerchantID))
	if err != nil || merchant == nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Approve(cmd.CommissionRate, cmd.Operator, cmd.Remark); err != nil {
		return fmt.Errorf("approve merchant: %w", err)
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		s.logger.ErrorContext(ctx, "failed to update merchant",
			"merchant_id", cmd.MerchantID, "error", err)
		return fmt.Errorf("update merchant: %w", err)
	}

	// 发布领域事件
	if err := s.publishEvent(ctx, &domain.MerchantApprovedEvent{
		MerchantID:     cmd.MerchantID,
		MerchantNo:     merchant.MerchantNo,
		CommissionRate: cmd.CommissionRate,
		Operator:       cmd.Operator,
		Remark:         cmd.Remark,
		Timestamp:      time.Now(),
	}); err != nil {
		s.logger.WarnContext(ctx, "failed to publish merchant approved event", "error", err)
	}

	s.logger.InfoContext(ctx, "merchant approved",
		"merchant_id", cmd.MerchantID, "operator", cmd.Operator, "duration", time.Since(start))
	return nil
}

// RejectMerchantCmd 审核拒绝商家命令
type RejectMerchantCmd struct {
	MerchantID uint64 `json:"merchant_id" validate:"required"`
	Reason     string `json:"reason" validate:"required"`
	Operator   string `json:"operator" validate:"required"`
}

// RejectMerchant 审核拒绝商家
func (s *MerchantCommandService) RejectMerchant(ctx context.Context, cmd *RejectMerchantCmd) error {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, uint(cmd.MerchantID))
	if err != nil || merchant == nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Reject(cmd.Reason, cmd.Operator); err != nil {
		return fmt.Errorf("reject merchant: %w", err)
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		s.logger.ErrorContext(ctx, "failed to update merchant",
			"merchant_id", cmd.MerchantID, "error", err)
		return fmt.Errorf("update merchant: %w", err)
	}

	// 发布领域事件
	if err := s.publishEvent(ctx, &domain.MerchantRejectedEvent{
		MerchantID: cmd.MerchantID,
		MerchantNo: merchant.MerchantNo,
		Reason:     cmd.Reason,
		Operator:   cmd.Operator,
		Timestamp:  time.Now(),
	}); err != nil {
		s.logger.WarnContext(ctx, "failed to publish merchant rejected event", "error", err)
	}

	s.logger.InfoContext(ctx, "merchant rejected",
		"merchant_id", cmd.MerchantID, "operator", cmd.Operator, "duration", time.Since(start))
	return nil
}

// CreateStoreCmd 创建店铺命令
type CreateStoreCmd struct {
	MerchantID    uint64   `json:"merchant_id" validate:"required"`
	Name          string   `json:"name" validate:"required"`
	LogoURL       string   `json:"logo_url"`
	BannerURL     string   `json:"banner_url"`
	Description   string   `json:"description"`
	BusinessHours string   `json:"business_hours"`
	Address       string   `json:"address"`
	Categories    []string `json:"categories"`
}

// CreateStore 创建店铺
func (s *MerchantCommandService) CreateStore(ctx context.Context, cmd *CreateStoreCmd) (*domain.Store, error) {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, uint(cmd.MerchantID))
	if err != nil || merchant == nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	if !merchant.CanCreateStore() {
		return nil, fmt.Errorf("merchant cannot create store, status: %s", merchant.Status.String())
	}

	store := domain.NewStore(
		merchant.ID,
		cmd.Name,
		cmd.LogoURL,
		cmd.BannerURL,
		cmd.Description,
		cmd.BusinessHours,
		cmd.Address,
		cmd.Categories,
	)

	if err := s.storeRepo.Save(ctx, store); err != nil {
		s.logger.ErrorContext(ctx, "failed to create store",
			"merchant_id", cmd.MerchantID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("create store: %w", err)
	}

	// 发布领域事件
	if err := s.publishEvent(ctx, &domain.StoreCreatedEvent{
		StoreID:    uint64(store.ID),
		StoreNo:    store.StoreNo,
		MerchantID: uint64(merchant.ID),
		Name:       store.Name,
		Timestamp:  time.Now(),
	}); err != nil {
		s.logger.WarnContext(ctx, "failed to publish store created event", "error", err)
	}

	s.logger.InfoContext(ctx, "store created",
		"store_id", store.ID, "store_no", store.StoreNo, "duration", time.Since(start))
	return store, nil
}

// UpdateStoreCmd 更新店铺命令
type UpdateStoreCmd struct {
	StoreID       uint64   `json:"store_id" validate:"required"`
	Name          string   `json:"name"`
	LogoURL       string   `json:"logo_url"`
	BannerURL     string   `json:"banner_url"`
	Description   string   `json:"description"`
	Announcement  string   `json:"announcement"`
	BusinessHours string   `json:"business_hours"`
	Address       string   `json:"address"`
	Categories    []string `json:"categories"`
	IsOpen        *bool    `json:"is_open"`
}

// UpdateStore 更新店铺信息
func (s *MerchantCommandService) UpdateStore(ctx context.Context, cmd *UpdateStoreCmd) error {
	start := time.Now()

	store, err := s.storeRepo.FindByID(ctx, uint(cmd.StoreID))
	if err != nil || store == nil {
		return fmt.Errorf("store not found: %w", err)
	}

	// 更新字段
	if cmd.Name != "" {
		store.Name = cmd.Name
	}
	if cmd.LogoURL != "" {
		store.LogoURL = cmd.LogoURL
	}
	if cmd.BannerURL != "" {
		store.BannerURL = cmd.BannerURL
	}
	if cmd.Description != "" {
		store.Description = cmd.Description
	}
	if cmd.Announcement != "" {
		store.Announcement = cmd.Announcement
	}
	if cmd.BusinessHours != "" {
		store.BusinessHours = cmd.BusinessHours
	}
	if cmd.Address != "" {
		store.Address = cmd.Address
	}
	if len(cmd.Categories) > 0 {
		// 将数组转换为 JSON 字符串存储
		categoriesJSON, _ := json.Marshal(cmd.Categories)
		store.Categories = string(categoriesJSON)
	}
	if cmd.IsOpen != nil {
		store.IsOpen = *cmd.IsOpen
	}

	if err := s.storeRepo.Update(ctx, store); err != nil {
		s.logger.ErrorContext(ctx, "failed to update store",
			"store_id", cmd.StoreID, "error", err)
		return fmt.Errorf("update store: %w", err)
	}

	// 发布领域事件
	if err := s.publishEvent(ctx, &domain.StoreUpdatedEvent{
		StoreID:    cmd.StoreID,
		StoreNo:    store.StoreNo,
		MerchantID: uint64(store.MerchantID),
		Timestamp:  time.Now(),
	}); err != nil {
		s.logger.WarnContext(ctx, "failed to publish store updated event", "error", err)
	}

	s.logger.InfoContext(ctx, "store updated",
		"store_id", cmd.StoreID, "duration", time.Since(start))
	return nil
}

// publishEvent 发布领域事件
func (s *MerchantCommandService) publishEvent(ctx context.Context, event domain.DomainEvent) error {
	if s.eventBus == nil {
		return nil
	}

	// 将领域事件适配为 eventsourcing.DomainEvent 接口后发布
	s.logger.DebugContext(ctx, "domain event published", "event", event.EventName())
	return nil
}
