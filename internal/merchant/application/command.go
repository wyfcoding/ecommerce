// Package application 商家服务应用层
// 生成摘要：
// 1) 实现 CQRS 模式的命令服务
// 2) 处理商家入驻、审核、店铺管理等用例
// 假设：
// - 所有操作需要事务支持
// - 领域事件通过消息队列异步发送
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

// CommandService 商家命令服务
type CommandService struct {
	merchantRepo    domain.MerchantRepository
	licenseRepo     domain.BusinessLicenseRepository
	bankAccountRepo domain.BankAccountRepository
	storeRepo       domain.StoreRepository
	settingsRepo    domain.MerchantSettingsRepository
	eventPublisher  messagequeue.EventPublisher
	logger          *slog.Logger
}

// NewCommandService 创建命令服务
func NewCommandService(
	merchantRepo domain.MerchantRepository,
	licenseRepo domain.BusinessLicenseRepository,
	bankAccountRepo domain.BankAccountRepository,
	storeRepo domain.StoreRepository,
	settingsRepo domain.MerchantSettingsRepository,
	eventPublisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *CommandService {
	return &CommandService{
		merchantRepo:    merchantRepo,
		licenseRepo:     licenseRepo,
		bankAccountRepo: bankAccountRepo,
		storeRepo:       storeRepo,
		settingsRepo:    settingsRepo,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

// ApplyCommand 入驻申请命令
type ApplyCommand struct {
	UserID       uint64
	Name         string
	LegalName    string
	LegalIDCard  string
	ContactName  string
	ContactPhone string
	ContactEmail string
	Type         domain.MerchantType
	License      *LicenseDTO
	BankAccount  *BankAccountDTO
	LogoURL      string
	Description  string
}

// LicenseDTO 营业执照DTO
type LicenseDTO struct {
	LicenseNo         string
	CompanyName       string
	CompanyType       string
	RegisteredCapital string
	EstablishmentDate string
	ValidUntil        string
	BusinessScope     string
	RegisteredAddress string
	LicenseImageURL   string
}

// BankAccountDTO 银行账户DTO
type BankAccountDTO struct {
	AccountName string
	AccountNo   string
	BankName    string
	BankBranch  string
	BankCode    string
}

// ApplyResult 入驻申请结果
type ApplyResult struct {
	MerchantID uint
	MerchantNo string
}

// Apply 处理入驻申请
func (s *CommandService) Apply(ctx context.Context, cmd ApplyCommand) (*ApplyResult, error) {
	start := time.Now()

	// 检查用户是否已申请
	existing, _ := s.merchantRepo.FindByUserID(ctx, cmd.UserID)
	if existing != nil {
		return nil, fmt.Errorf("user has already applied for merchant, status: %s", existing.Status.String())
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
		cmd.Type,
	)
	merchant.LogoURL = cmd.LogoURL
	merchant.Description = cmd.Description

	// 保存商家
	if err := s.merchantRepo.Save(ctx, merchant); err != nil {
		s.logger.ErrorContext(ctx, "failed to save merchant",
			"user_id", cmd.UserID,
			"error", err,
			"duration", time.Since(start))
		return nil, fmt.Errorf("failed to save merchant: %w", err)
	}

	// 保存营业执照
	if cmd.License != nil {
		license := &domain.BusinessLicense{
			MerchantID:        merchant.ID,
			LicenseNo:         cmd.License.LicenseNo,
			CompanyName:       cmd.License.CompanyName,
			CompanyType:       cmd.License.CompanyType,
			RegisteredCapital: cmd.License.RegisteredCapital,
			EstablishmentDate: cmd.License.EstablishmentDate,
			ValidUntil:        cmd.License.ValidUntil,
			BusinessScope:     cmd.License.BusinessScope,
			RegisteredAddress: cmd.License.RegisteredAddress,
			LicenseImageURL:   cmd.License.LicenseImageURL,
		}
		if err := s.licenseRepo.Save(ctx, license); err != nil {
			s.logger.ErrorContext(ctx, "failed to save business license",
				"merchant_id", merchant.ID,
				"error", err)
			// 不阻塞主流程
		}
	}

	// 保存银行账户
	if cmd.BankAccount != nil {
		account := &domain.BankAccount{
			MerchantID:  merchant.ID,
			AccountName: cmd.BankAccount.AccountName,
			AccountNo:   cmd.BankAccount.AccountNo,
			BankName:    cmd.BankAccount.BankName,
			BankBranch:  cmd.BankAccount.BankBranch,
			BankCode:    cmd.BankAccount.BankCode,
		}
		if err := s.bankAccountRepo.Save(ctx, account); err != nil {
			s.logger.ErrorContext(ctx, "failed to save bank account",
				"merchant_id", merchant.ID,
				"error", err)
		}
	}

	// 创建默认商家设置
	settings := &domain.MerchantSettings{
		MerchantID:             merchant.ID,
		AutoConfirmDays:        7,
		EnableInvoice:          true,
		OrderNotification:      true,
		StockAlert:             true,
		ReviewNotification:     true,
		SettlementNotification: true,
		NotificationPhone:      cmd.ContactPhone,
		NotificationEmail:      cmd.ContactEmail,
	}
	if err := s.settingsRepo.Save(ctx, settings); err != nil {
		s.logger.ErrorContext(ctx, "failed to save merchant settings",
			"merchant_id", merchant.ID,
			"error", err)
	}

	// 发布领域事件
	s.publishEvents(ctx, merchant.GetDomainEvents())
	merchant.ClearDomainEvents()

	s.logger.InfoContext(ctx, "merchant application submitted",
		"merchant_id", merchant.ID,
		"merchant_no", merchant.MerchantNo,
		"user_id", cmd.UserID,
		"duration", time.Since(start))

	return &ApplyResult{
		MerchantID: merchant.ID,
		MerchantNo: merchant.MerchantNo,
	}, nil
}

// ApproveCommand 审核通过命令
type ApproveCommand struct {
	MerchantID     uint
	CommissionRate float64
	Operator       string
	Remark         string
}

// Approve 审核通过
func (s *CommandService) Approve(ctx context.Context, cmd ApproveCommand) error {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Approve(cmd.CommissionRate, cmd.Operator, cmd.Remark); err != nil {
		return err
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	s.publishEvents(ctx, merchant.GetDomainEvents())
	merchant.ClearDomainEvents()

	s.logger.InfoContext(ctx, "merchant approved",
		"merchant_id", cmd.MerchantID,
		"operator", cmd.Operator,
		"commission_rate", cmd.CommissionRate,
		"duration", time.Since(start))

	return nil
}

// RejectCommand 审核拒绝命令
type RejectCommand struct {
	MerchantID uint
	Reason     string
	Operator   string
}

// Reject 审核拒绝
func (s *CommandService) Reject(ctx context.Context, cmd RejectCommand) error {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Reject(cmd.Reason, cmd.Operator); err != nil {
		return err
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	s.publishEvents(ctx, merchant.GetDomainEvents())
	merchant.ClearDomainEvents()

	s.logger.InfoContext(ctx, "merchant rejected",
		"merchant_id", cmd.MerchantID,
		"operator", cmd.Operator,
		"reason", cmd.Reason,
		"duration", time.Since(start))

	return nil
}

// DisableCommand 禁用商家命令
type DisableCommand struct {
	MerchantID uint
	Reason     string
	Operator   string
}

// Disable 禁用商家
func (s *CommandService) Disable(ctx context.Context, cmd DisableCommand) error {
	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Disable(cmd.Reason, cmd.Operator); err != nil {
		return err
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	s.publishEvents(ctx, merchant.GetDomainEvents())
	merchant.ClearDomainEvents()

	s.logger.InfoContext(ctx, "merchant disabled",
		"merchant_id", cmd.MerchantID,
		"operator", cmd.Operator)

	return nil
}

// EnableCommand 启用商家命令
type EnableCommand struct {
	MerchantID uint
	Operator   string
}

// Enable 启用商家
func (s *CommandService) Enable(ctx context.Context, cmd EnableCommand) error {
	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if err := merchant.Enable(cmd.Operator); err != nil {
		return err
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	s.publishEvents(ctx, merchant.GetDomainEvents())
	merchant.ClearDomainEvents()

	s.logger.InfoContext(ctx, "merchant enabled",
		"merchant_id", cmd.MerchantID,
		"operator", cmd.Operator)

	return nil
}

// UpdateMerchantCommand 更新商家命令
type UpdateMerchantCommand struct {
	MerchantID   uint
	Name         *string
	ContactName  *string
	ContactPhone *string
	ContactEmail *string
	LogoURL      *string
	Description  *string
	BankAccount  *BankAccountDTO
}

// UpdateMerchant 更新商家信息
func (s *CommandService) UpdateMerchant(ctx context.Context, cmd UpdateMerchantCommand) error {
	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return fmt.Errorf("merchant not found: %w", err)
	}

	if cmd.Name != nil {
		merchant.Name = *cmd.Name
	}
	if cmd.ContactName != nil {
		merchant.ContactName = *cmd.ContactName
	}
	if cmd.ContactPhone != nil {
		merchant.ContactPhone = *cmd.ContactPhone
	}
	if cmd.ContactEmail != nil {
		merchant.ContactEmail = *cmd.ContactEmail
	}
	if cmd.LogoURL != nil {
		merchant.LogoURL = *cmd.LogoURL
	}
	if cmd.Description != nil {
		merchant.Description = *cmd.Description
	}

	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	// 更新银行账户
	if cmd.BankAccount != nil {
		account, _ := s.bankAccountRepo.FindByMerchantID(ctx, cmd.MerchantID)
		if account != nil {
			account.AccountName = cmd.BankAccount.AccountName
			account.AccountNo = cmd.BankAccount.AccountNo
			account.BankName = cmd.BankAccount.BankName
			account.BankBranch = cmd.BankAccount.BankBranch
			account.BankCode = cmd.BankAccount.BankCode
			if err := s.bankAccountRepo.Update(ctx, account); err != nil {
				s.logger.ErrorContext(ctx, "failed to update bank account",
					"merchant_id", cmd.MerchantID,
					"error", err)
			}
		}
	}

	s.logger.InfoContext(ctx, "merchant updated",
		"merchant_id", cmd.MerchantID)

	return nil
}

// CreateStoreCommand 创建店铺命令
type CreateStoreCommand struct {
	MerchantID    uint
	Name          string
	LogoURL       string
	BannerURL     string
	Description   string
	Categories    []string
	BusinessHours string
	Address       string
}

// CreateStore 创建店铺
func (s *CommandService) CreateStore(ctx context.Context, cmd CreateStoreCommand) (*domain.Store, error) {
	start := time.Now()

	merchant, err := s.merchantRepo.FindByID(ctx, cmd.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	if !merchant.CanCreateStore() {
		return nil, fmt.Errorf("merchant cannot create store, status: %s", merchant.Status.String())
	}

	store := domain.NewStore(
		cmd.MerchantID,
		cmd.Name,
		cmd.LogoURL,
		cmd.BannerURL,
		cmd.Description,
		cmd.BusinessHours,
		cmd.Address,
		cmd.Categories,
	)

	// 序列化类目
	if len(cmd.Categories) > 0 {
		categoriesJSON, _ := json.Marshal(cmd.Categories)
		store.Categories = string(categoriesJSON)
	}

	if err := s.storeRepo.Save(ctx, store); err != nil {
		return nil, fmt.Errorf("failed to save store: %w", err)
	}

	s.logger.InfoContext(ctx, "store created",
		"store_id", store.ID,
		"store_no", store.StoreNo,
		"merchant_id", cmd.MerchantID,
		"duration", time.Since(start))

	return store, nil
}

// UpdateStoreCommand 更新店铺命令
type UpdateStoreCommand struct {
	StoreID       uint
	Name          *string
	LogoURL       *string
	BannerURL     *string
	Description   *string
	Announcement  *string
	Categories    []string
	IsOpen        *bool
	BusinessHours *string
	Address       *string
}

// UpdateStore 更新店铺
func (s *CommandService) UpdateStore(ctx context.Context, cmd UpdateStoreCommand) (*domain.Store, error) {
	store, err := s.storeRepo.FindByID(ctx, cmd.StoreID)
	if err != nil {
		return nil, fmt.Errorf("store not found: %w", err)
	}

	if cmd.Name != nil {
		store.Name = *cmd.Name
	}
	if cmd.LogoURL != nil {
		store.LogoURL = *cmd.LogoURL
	}
	if cmd.BannerURL != nil {
		store.BannerURL = *cmd.BannerURL
	}
	if cmd.Description != nil {
		store.Description = *cmd.Description
	}
	if cmd.Announcement != nil {
		store.Announcement = *cmd.Announcement
	}
	if len(cmd.Categories) > 0 {
		categoriesJSON, _ := json.Marshal(cmd.Categories)
		store.Categories = string(categoriesJSON)
	}
	if cmd.IsOpen != nil {
		store.IsOpen = *cmd.IsOpen
	}
	if cmd.BusinessHours != nil {
		store.BusinessHours = *cmd.BusinessHours
	}
	if cmd.Address != nil {
		store.Address = *cmd.Address
	}

	if err := s.storeRepo.Update(ctx, store); err != nil {
		return nil, fmt.Errorf("failed to update store: %w", err)
	}

	s.logger.InfoContext(ctx, "store updated",
		"store_id", cmd.StoreID)

	return store, nil
}

// UpdateSettingsCommand 更新设置命令
type UpdateSettingsCommand struct {
	MerchantID uint
	Settings   *SettingsDTO
}

// SettingsDTO 设置DTO
type SettingsDTO struct {
	AutoConfirmOrder       *bool
	AutoConfirmDays        *int32
	EnableCOD              *bool
	EnableInvoice          *bool
	SupportedPayments      []string
	FreeShippingThreshold  *int64
	DefaultShippingFee     *int64
	EnableSameDayDelivery  *bool
	OrderNotification      *bool
	StockAlert             *bool
	ReviewNotification     *bool
	SettlementNotification *bool
	NotificationPhone      *string
	NotificationEmail      *string
}

// UpdateSettings 更新商家设置
func (s *CommandService) UpdateSettings(ctx context.Context, cmd UpdateSettingsCommand) (*domain.MerchantSettings, error) {
	settings, err := s.settingsRepo.FindByMerchantID(ctx, cmd.MerchantID)
	if err != nil {
		// 不存在则创建
		settings = &domain.MerchantSettings{
			MerchantID: cmd.MerchantID,
		}
	}

	dto := cmd.Settings
	if dto.AutoConfirmOrder != nil {
		settings.AutoConfirmOrder = *dto.AutoConfirmOrder
	}
	if dto.AutoConfirmDays != nil {
		settings.AutoConfirmDays = *dto.AutoConfirmDays
	}
	if dto.EnableCOD != nil {
		settings.EnableCOD = *dto.EnableCOD
	}
	if dto.EnableInvoice != nil {
		settings.EnableInvoice = *dto.EnableInvoice
	}
	if len(dto.SupportedPayments) > 0 {
		paymentsJSON, _ := json.Marshal(dto.SupportedPayments)
		settings.SupportedPayments = string(paymentsJSON)
	}
	if dto.FreeShippingThreshold != nil {
		settings.FreeShippingThreshold = *dto.FreeShippingThreshold
	}
	if dto.DefaultShippingFee != nil {
		settings.DefaultShippingFee = *dto.DefaultShippingFee
	}
	if dto.EnableSameDayDelivery != nil {
		settings.EnableSameDayDelivery = *dto.EnableSameDayDelivery
	}
	if dto.OrderNotification != nil {
		settings.OrderNotification = *dto.OrderNotification
	}
	if dto.StockAlert != nil {
		settings.StockAlert = *dto.StockAlert
	}
	if dto.ReviewNotification != nil {
		settings.ReviewNotification = *dto.ReviewNotification
	}
	if dto.SettlementNotification != nil {
		settings.SettlementNotification = *dto.SettlementNotification
	}
	if dto.NotificationPhone != nil {
		settings.NotificationPhone = *dto.NotificationPhone
	}
	if dto.NotificationEmail != nil {
		settings.NotificationEmail = *dto.NotificationEmail
	}

	if settings.ID == 0 {
		if err := s.settingsRepo.Save(ctx, settings); err != nil {
			return nil, fmt.Errorf("failed to save settings: %w", err)
		}
	} else {
		if err := s.settingsRepo.Update(ctx, settings); err != nil {
			return nil, fmt.Errorf("failed to update settings: %w", err)
		}
	}

	s.logger.InfoContext(ctx, "merchant settings updated",
		"merchant_id", cmd.MerchantID)

	return settings, nil
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
