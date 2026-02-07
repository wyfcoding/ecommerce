package domain

import (
	"fmt"
	"time"
)

// User 实体是用户模块的聚合根。
// 包含了用户的基本信息、认证凭据、个人资料、状态以及关联的地址列表。
type User struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Username  string     `json:"username"`   // 用户名，唯一索引，不允许为空。
	Email     string     `json:"email"`      // 邮箱，唯一索引，不允许为空。
	FullName  string     `json:"full_name"`  // 全名。
	Password  string     `json:"-"`          // 密码（加密存储），JSON序列化时忽略。
	Phone     string     `json:"phone"`      // 手机号，索引字段。
	Nickname  string     `json:"nickname"`   // 昵称。
	Avatar    string     `json:"avatar"`     // 头像URL。
	Gender    int8       `json:"gender"`     // 性别 0:未知 1:男 2:女，默认为未知。
	Birthday  *time.Time `json:"birthday"`   // 生日。
	Status    int8       `json:"status"`     // 状态 1:正常 2:禁用，默认为正常。
	Addresses []*Address `json:"addresses"`  // 关联的地址列表，一对多关系。
	Roles     []*Role    `json:"roles"`      // 关联的角色列表。
	Tier      *UserTier  `json:"user_tier"`  // 关联的用户等级信息。
}

// Address 实体代表用户的收货地址信息。
// 它是User聚合根的一部分。
type Address struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UserID          uint      `json:"user_id"`      // 所属用户ID，索引字段，不允许为空。
	RecipientName   string    `json:"name"`         // 收货人姓名，不允许为空。
	PhoneNumber     string    `json:"phone"`        // 收货人电话，不允许为空。
	Province        string    `json:"province"`     // 省份，不允许为空。
	City            string    `json:"city"`         // 城市，不允许为空。
	District        string    `json:"district"`     // 区/县，不允许为空。
	DetailedAddress string    `json:"detail"`       // 详细地址，不允许为空。
	PostalCode      string    `json:"postal_code"`  // 邮政编码。
	IsDefault       bool      `json:"is_default"`   // 是否默认地址，默认为否。
}

// NewUser 是一个工厂方法，用于创建并返回一个新的 User 实体实例。
func NewUser(username, email, password, phone string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	return &User{
		Username:  username,
		Email:     email,
		Password:  password,
		Phone:     phone,
		Status:    1,            // 默认状态为正常。
		Addresses: []*Address{}, // 初始化地址列表。
	}, nil
}

// UpdateProfile 更新用户个人信息。
// nickname: 昵称。
// avatar: 头像URL。
// gender: 性别。
// birthday: 生日。
func (u *User) UpdateProfile(nickname, avatar string, gender int8, birthday *time.Time) {
	if nickname != "" {
		u.Nickname = nickname
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	// 只有当gender >= 0时才更新，允许0表示未知。
	if gender >= 0 {
		u.Gender = gender
	}
	if birthday != nil {
		u.Birthday = birthday
	}
	// 备注：GORM 会自动更新 UpdatedAt 字段。
	// 此领域方法主要负责修改内存中的实体状态，持久化由仓储完成。
}

// ChangePassword 修改用户密码。
// newPassword: 新密码（应为哈希后的密码）。
func (u *User) ChangePassword(newPassword string) error {
	if newPassword == "" {
		return fmt.Errorf("new password cannot be empty")
	}
	u.Password = newPassword // 更新密码。
	return nil
}

// AddAddress 为用户添加一个新地址。
// address: 待添加的地址实体。
func (u *User) AddAddress(address *Address) error {
	// 关联用户ID。在持久化时由GORM处理，或者在此处显式赋值（如果u.ID已知）。
	if u.ID != 0 {
		address.UserID = u.ID
	}

	u.Addresses = append(u.Addresses, address) // 先添加到列表中

	// 如果新地址设置为默认，则调用私有方法处理默认地址逻辑。
	if address.IsDefault {
		if err := u.setDefaultAddress(address.ID); err != nil {
			// 如果是新创建的地址且ID还未分配（ID为0），setDefaultAddress 可能会因为找不到ID而报错。
			// 此时我们直接遍历列表，通过对象指针匹配来设置默认状态。
			for _, addr := range u.Addresses {
				// 取消其他所有地址的默认状态
				addr.IsDefault = (addr == address)
			}
		}
	}

	return nil
}

// UpdateAddress 更新用户的一个地址。
// addressID: 待更新地址的ID。
// address: 包含更新信息的地址实体。
func (u *User) UpdateAddress(addressID uint, address *Address) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			// 更新地址字段。
			u.Addresses[i].RecipientName = address.RecipientName
			u.Addresses[i].PhoneNumber = address.PhoneNumber
			u.Addresses[i].Province = address.Province
			u.Addresses[i].City = address.City
			u.Addresses[i].District = address.District
			u.Addresses[i].DetailedAddress = address.DetailedAddress
			u.Addresses[i].PostalCode = address.PostalCode

			// 处理默认地址逻辑。
			if address.IsDefault {
				if err := u.setDefaultAddress(addressID); err != nil {
					return err
				}
				u.Addresses[i].IsDefault = true
			} else {
				u.Addresses[i].IsDefault = false
			}
			return nil
		}
	}
	return fmt.Errorf("address not found")
}

// RemoveAddress 移除用户的地址。
// addressID: 待移除地址的ID。
func (u *User) RemoveAddress(addressID uint) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...) // 从列表中移除地址。
			return nil
		}
	}
	return fmt.Errorf("address not found")
}

// SetDefaultAddress 设置用户默认地址。
// addressID: 设为默认地址的ID。
func (u *User) SetDefaultAddress(addressID uint) error {
	return u.setDefaultAddress(addressID)
}

// setDefaultAddress 是一个私有辅助方法，用于处理设置默认地址的逻辑。
func (u *User) setDefaultAddress(addressID uint) error {
	found := false
	// 1. 先取消所有现有地址的默认状态。
	for _, addr := range u.Addresses {
		if addr.ID == addressID {
			found = true
		}
		addr.IsDefault = false
	}

	// 如果要设置的地址ID不存在于用户地址列表中，且不是设置为空默认，则报错。
	if !found && addressID != 0 {
		return fmt.Errorf("address not found")
	}

	// 2. 设置新的默认地址。
	for _, addr := range u.Addresses {
		if addr.ID == addressID {
			addr.IsDefault = true
			return nil
		}
	}
	return nil // 如果addressID为0（表示没有默认地址），则会走到这里。
}

// GetDefaultAddress 获取用户的默认地址。
func (u *User) GetDefaultAddress() *Address {
	for _, addr := range u.Addresses {
		if addr.IsDefault {
			return addr
		}
	}
	return nil // 如果没有默认地址，返回nil。
}

// NewAddress 是一个工厂方法，用于创建并返回一个新的 Address 实体实例。
func NewAddress(userID uint, recipientName, phoneNumber, province, city, district, detailedAddress, postalCode string, isDefault bool) *Address {
	return &Address{
		UserID:          userID,
		RecipientName:   recipientName,
		PhoneNumber:     phoneNumber,
		Province:        province,
		City:            city,
		District:        district,
		DetailedAddress: detailedAddress,
		PostalCode:      postalCode,
		IsDefault:       isDefault,
	}
}
