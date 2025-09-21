package biz

import (
	"context"
	"errors"
	"fmt" // 导入 fmt 包
	"regexp"
)

// AddressUsecase 封装了地址相关的业务逻辑。
type AddressUsecase struct {
	repo AddressRepo
}

// NewAddressUsecase 是 AddressUsecase 的构造函数。
func NewAddressUsecase(repo AddressRepo) *AddressUsecase {
	return &AddressUsecase{repo: repo}
}

// validatePhoneNumber 校验手机号格式。
func (uc *AddressUsecase) validatePhoneNumber(phone string) bool {
	// 一个简单的手机号校验规则 (11位数字，以1开头)
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(phone)
}

// CreateAddress 负责创建收货地址的业务逻辑。
func (uc *AddressUsecase) CreateAddress(ctx context.Context, addr *Address) (*Address, error) {
	// 业务校验：例如手机号格式
	if addr.Phone == nil || !uc.validatePhoneNumber(*addr.Phone) {
		return nil, errors.New("手机号格式不正确")
	}
	// 其他校验，如地址库校验、字段非空等可在此处添加
	if addr.Name == nil || *addr.Name == "" {
		return nil, errors.New("收货人姓名不能为空")
	}

	return uc.repo.CreateAddress(ctx, addr)
}

// UpdateAddress 负责更新收货地址的业务逻辑。
func (uc *AddressUsecase) UpdateAddress(ctx context.Context, addr *Address) (*Address, error) {
	if addr.Phone != nil && !uc.validatePhoneNumber(*addr.Phone) {
		return nil, fmt.Errorf("手机号格式不正确")
	}
	return uc.repo.UpdateAddress(ctx, addr)
}

// DeleteAddress 负责删除收货地址的业务逻辑。
func (uc *AddressUsecase) DeleteAddress(ctx context.Context, userID, addrID uint64) error {
	return uc.repo.DeleteAddress(ctx, userID, addrID)
}

// GetAddress 负责获取单个地址的业务逻辑。
func (uc *AddressUsecase) GetAddress(ctx context.Context, userID, addrID uint64) (*Address, error) {
	return uc.repo.GetAddress(ctx, userID, addrID)
}

// ListAddresses 负责获取地址列表的业务逻辑。
func (uc *AddressUsecase) ListAddresses(ctx context.Context, userID uint64) ([]*Address, error) {
	return uc.repo.ListAddresses(ctx, userID)
}

// SetDefaultAddress 负责设置默认地址的业务逻辑。
func (uc *AddressUsecase) SetDefaultAddress(ctx context.Context, userID, addrID uint64) error {
	return uc.repo.SetDefaultAddress(ctx, userID, addrID)
}