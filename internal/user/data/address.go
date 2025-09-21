package data

import (
	"context"
	"ecommerce/internal/user/biz"
	"ecommerce/internal/user/data/model"

	"gorm.io/gorm"
)

type addressRepo struct {
	*Data
}

// NewAddressRepo 是 addressRepo 的构造函数。
func NewAddressRepo(data *Data) biz.AddressRepo {
	return &addressRepo{Data: data}
}

// toBizAddress 将数据库模型 data.Address 转换为业务领域模型 biz.Address。
func (ar *addressRepo) toBizAddress(addr *model.Address) *biz.Address {
	if addr == nil {
		return nil
	}
	return &biz.Address{
		ID:              addr.ID,
		UserID:          addr.UserID,
		Name:            &addr.Name,
		Phone:           &addr.Phone,
		Province:        &addr.Province,
		City:            &addr.City,
		District:        &addr.District,
		DetailedAddress: &addr.DetailedAddress,
		IsDefault:       &addr.IsDefault,
	}
}

// unsetOldDefault 在事务中将用户的所有地址设置为非默认。
func (ar *addressRepo) unsetOldDefault(tx *gorm.DB, userID uint64) error {
	return tx.Model(&model.Address{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false).Error
}

// CreateAddress 创建一个新的收货地址。
func (ar *addressRepo) CreateAddress(ctx context.Context, addr *biz.Address) (*biz.Address, error) {
	po := &model.Address{ // 使用 model.Address
		UserID: addr.UserID,
	}
	if addr.Name != nil {
		po.Name = *addr.Name
	}
	if addr.Phone != nil {
		po.Phone = *addr.Phone
	}
	if addr.Province != nil {
		po.Province = *addr.Province
	}
	if addr.City != nil {
		po.City = *addr.City
	}
	if addr.District != nil {
		po.District = *addr.District
	}
	if addr.DetailedAddress != nil {
		po.DetailedAddress = *addr.DetailedAddress
	}
	if addr.IsDefault != nil {
		po.IsDefault = *addr.IsDefault
	}

	err := ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if po.IsDefault {
			if err := ar.unsetOldDefault(tx, addr.UserID); err != nil {
				return err
			}
		}
		return tx.Create(po).Error
	})

	if err != nil {
		return nil, err
	}
	return ar.toBizAddress(po), nil
}

// UpdateAddress 更新一个已有的收货地址。
func (ar *addressRepo) UpdateAddress(ctx context.Context, addr *biz.Address) (*biz.Address, error) {
	po := &model.Address{} // 使用 model.Address

	err := ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", addr.ID, addr.UserID).First(po).Error; err != nil {
			return err
		}

		updates := make(map[string]interface{})
		if addr.Name != nil {
			updates["name"] = *addr.Name
		}
		if addr.Phone != nil {
			updates["phone"] = *addr.Phone
		}
		if addr.Province != nil {
			updates["province"] = *addr.Province
		}
		if addr.City != nil {
			updates["city"] = *addr.City
		}
		if addr.District != nil {
			updates["district"] = *addr.District
		}
		if addr.DetailedAddress != nil {
			updates["detailed_address"] = *addr.DetailedAddress
		}
		if addr.IsDefault != nil {
			if *addr.IsDefault && !po.IsDefault {
				if err := ar.unsetOldDefault(tx, addr.UserID); err != nil {
					return err
				}
			}
			updates["is_default"] = *addr.IsDefault
		}

		if len(updates) == 0 {
			return nil // 没有需要更新的字段
		}

		return tx.Model(po).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}
	// 查询更新后的完整数据并返回
	return ar.GetAddress(ctx, addr.UserID, addr.ID)
}

// DeleteAddress 删除一个收货地址。
func (ar *addressRepo) DeleteAddress(ctx context.Context, userID, addrID uint64) error {
	return ar.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Address{}, addrID).Error // 使用 model.Address
}

// GetAddress 获取单个收货地址的详情。
func (ar *addressRepo) GetAddress(ctx context.Context, userID, addrID uint64) (*biz.Address, error) {
	var addr model.Address // 使用 model.Address
	if err := ar.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, addrID).First(&addr).Error; err != nil {
		return nil, err
	}
	return ar.toBizAddress(&addr), nil
}

// ListAddresses 获取一个用户的所有收货地址。
func (ar *addressRepo) ListAddresses(ctx context.Context, userID uint64) ([]*biz.Address, error) {
	var addrs []*model.Address // 使用 model.Address
	if err := ar.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default desc, id desc").Find(&addrs).Error; err != nil {
		return nil, err
	}
	res := make([]*biz.Address, len(addrs))
	for i, addr := range addrs {
		res[i] = ar.toBizAddress(addr)
	}
	return res, nil
}

// SetDefaultAddress 设置默认收货地址。
func (ar *addressRepo) SetDefaultAddress(ctx context.Context, userID, addrID uint64) error {
	return ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 取消旧的默认地址
		if err := ar.unsetOldDefault(tx, userID); err != nil {
			return err
		}
		// 2. 设置新的默认地址
		return tx.Model(&model.Address{}).Where("id = ? AND user_id = ?", addrID, userID).Update("is_default", true).Error
	})
}
