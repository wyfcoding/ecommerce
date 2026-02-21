package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/wyfcoding/ecommerce/internal/address/domain"
)

// AddressService 提供地址管理的用例能力。
// 包含创建、更新、删除、设置默认、查询地址等操作。
type AddressService struct {
	repo   domain.AddressRepository
	logger *slog.Logger
}

func NewAddressService(repo domain.AddressRepository, logger *slog.Logger) *AddressService {
	return &AddressService{
		repo:   repo,
		logger: logger,
	}
}

// CreateAddress 创建新地址
func (s *AddressService) CreateAddress(ctx context.Context, addr *domain.Address) error {
	s.logger.Info("creating new address", "user_id", addr.UserID)

	// 1. 简单的业务规则：检查地址总数是否超过限制 (比如每个用户最多 20 个地址)
	count, err := s.repo.CountByUserID(ctx, addr.UserID)
	if err != nil {
		s.logger.Error("failed to count addresses", "error", err, "user_id", addr.UserID)
		return err
	}
	if count >= 20 {
		return domain.ErrAddressLimitExceeded
	}

	addr.ID = uuid.New().String()

	// 2. 如果这是用户的第一个地址，或者设置为默认地址，则需要处理默认状态
	if count == 0 {
		addr.IsDefault = true
	} else if addr.IsDefault {
		// 需将该用户现有的默认地址清空
		if err := s.repo.ClearDefaultByUserID(ctx, addr.UserID); err != nil {
			s.logger.Error("failed to clear old default address", "error", err, "user_id", addr.UserID)
			return err
		}
	}

	// 3. 保存
	if err := s.repo.Save(ctx, addr); err != nil {
		s.logger.Error("failed to save address", "error", err, "id", addr.ID)
		return err
	}

	return nil
}

// UpdateAddress 更新地址信息
func (s *AddressService) UpdateAddress(ctx context.Context, addr *domain.Address) error {
	s.logger.Info("updating address", "id", addr.ID, "user_id", addr.UserID)

	// 1. 确认地址存在且属于该用户
	existing, err := s.repo.FindByID(ctx, addr.ID)
	if err != nil {
		return err
	}
	if existing.UserID != addr.UserID {
		return domain.ErrAddressNotFound // 防越权
	}

	// 2. 如果之前不是默认，现在改成默认，需要清除其余默认
	if !existing.IsDefault && addr.IsDefault {
		if err := s.repo.ClearDefaultByUserID(ctx, addr.UserID); err != nil {
			s.logger.Error("failed to clear old default address", "error", err)
			return err
		}
	}

	// 保留必要属性
	addr.Model = existing.Model

	if err := s.repo.Update(ctx, addr); err != nil {
		s.logger.Error("failed to update address", "error", err)
		return err
	}

	return nil
}

// DeleteAddress 删除指定地址
func (s *AddressService) DeleteAddress(ctx context.Context, id string, userID int64) error {
	s.logger.Info("deleting address", "id", id, "user_id", userID)
	return s.repo.Delete(ctx, id, userID)
}

// SetDefaultAddress 设置用户的默认地址
func (s *AddressService) SetDefaultAddress(ctx context.Context, id string, userID int64) error {
	s.logger.Info("setting default address", "id", id, "user_id", userID)

	// 1. 清除当前用户所有默认地址状态
	if err := s.repo.ClearDefaultByUserID(ctx, userID); err != nil {
		s.logger.Error("failed to clear old default address", "error", err)
		return err
	}

	// 2. 将指定ID设为默认
	if err := s.repo.SetDefault(ctx, id, userID); err != nil {
		s.logger.Error("failed to set default address", "error", err)
		return err
	}

	return nil
}

// ListAddresses 获取用户的所有地址
func (s *AddressService) ListAddresses(ctx context.Context, userID int64) ([]*domain.Address, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// GetAddress 获取单条地址详情
func (s *AddressService) GetAddress(ctx context.Context, id string) (*domain.Address, error) {
	return s.repo.FindByID(ctx, id)
}
