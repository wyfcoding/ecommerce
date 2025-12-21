package application

import (
	"context"
	"errors"

	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// AddressService 定义了 Address 相关的服务逻辑。
type AddressService struct {
	userRepo    domain.UserRepository
	addressRepo domain.AddressRepository
	logger      *slog.Logger
}

// NewAddressService 创建 Address 服务实例。
func NewAddressService(userRepo domain.UserRepository, addressRepo domain.AddressRepository, logger *slog.Logger) *AddressService {
	return &AddressService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		logger:      logger,
	}
}

// AddAddress 添加用户地址。
func (s *AddressService) AddAddress(ctx context.Context, userID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	address := domain.NewAddress(uint(userID), name, phone, province, city, district, detail, "", isDefault)
	address.ID = uint(idgen.GenID())

	if err := s.addressRepo.Save(ctx, address); err != nil {
		s.logger.ErrorContext(ctx, "failed to add address", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "address added successfully", "user_id", userID, "address_id", address.ID)

	if isDefault {
		if err := s.addressRepo.SetDefault(ctx, uint(userID), address.ID); err != nil {
			return nil, err
		}
		address.IsDefault = true
	}

	return address, nil
}

// ListAddresses 列出用户的所有地址。
func (s *AddressService) ListAddresses(ctx context.Context, userID uint64) ([]*domain.Address, error) {
	return s.addressRepo.FindByUserID(ctx, uint(userID))
}

// UpdateAddress 更新用户地址信息。
func (s *AddressService) UpdateAddress(ctx context.Context, userID, addressID uint64, name, phone, province, city, district, detail string, isDefault bool) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found or not owned by user")
	}

	if name != "" {
		address.RecipientName = name
	}
	if phone != "" {
		address.PhoneNumber = phone
	}
	if province != "" {
		address.Province = province
	}
	if city != "" {
		address.City = city
	}
	if district != "" {
		address.District = district
	}
	if detail != "" {
		address.DetailedAddress = detail
	}

	if err := s.addressRepo.Update(ctx, address); err != nil {
		s.logger.ErrorContext(ctx, "failed to update address", "user_id", userID, "address_id", addressID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "address updated successfully", "user_id", userID, "address_id", addressID)

	if isDefault {
		if err := s.addressRepo.SetDefault(ctx, uint(userID), uint(addressID)); err != nil {
			return nil, err
		}
		address.IsDefault = true
	}

	return address, nil
}

// DeleteAddress 删除用户地址。
func (s *AddressService) DeleteAddress(ctx context.Context, userID, addressID uint64) error {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return err
	}
	if address == nil || address.UserID != uint(userID) {
		return errors.New("address not found or not owned by user")
	}

	if err := s.addressRepo.Delete(ctx, uint(addressID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete address", "user_id", userID, "address_id", addressID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "address deleted successfully", "user_id", userID, "address_id", addressID)
	return nil
}

// GetAddress 获取地址详情。
func (s *AddressService) GetAddress(ctx context.Context, userID, addressID uint64) (*domain.Address, error) {
	address, err := s.addressRepo.FindByID(ctx, uint(addressID))
	if err != nil {
		return nil, err
	}
	if address == nil || address.UserID != uint(userID) {
		return nil, errors.New("address not found or not owned by user")
	}
	return address, nil
}
