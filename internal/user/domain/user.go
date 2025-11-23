package domain

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User 用户聚合根
// 包含用户的基本信息、状态以及关联的地址列表
type User struct {
	gorm.Model
	Username  string     `gorm:"column:username;type:varchar(255);uniqueIndex;not null" json:"username"` // 用户名，唯一
	Email     string     `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`       // 邮箱，唯一
	Password  string     `gorm:"column:password;type:varchar(255);not null" json:"-"`                    // 密码（加密存储）
	Phone     string     `gorm:"column:phone;type:varchar(20);index" json:"phone"`                       // 手机号
	Nickname  string     `gorm:"column:nickname;type:varchar(255)" json:"nickname"`                      // 昵称
	Avatar    string     `gorm:"column:avatar;type:varchar(1024)" json:"avatar"`                         // 头像URL
	Gender    int8       `gorm:"column:gender;type:tinyint;default:0" json:"gender"`                     // 性别 0:未知 1:男 2:女
	Birthday  *time.Time `gorm:"column:birthday;type:date" json:"birthday"`                              // 生日
	Status    int8       `gorm:"column:status;type:tinyint;default:1" json:"status"`                     // 状态 1:正常 2:禁用
	Addresses []*Address `gorm:"foreignKey:UserID" json:"addresses"`                                     // 关联地址列表
}

// Address 地址实体
// 用户的收货地址信息
type Address struct {
	gorm.Model
	UserID          uint   `gorm:"column:user_id;index;not null" json:"user_id"`                     // 所属用户ID
	RecipientName   string `gorm:"column:recipient_name;type:varchar(255);not null" json:"name"`     // 收货人姓名
	PhoneNumber     string `gorm:"column:phone_number;type:varchar(20);not null" json:"phone"`       // 收货人电话
	Province        string `gorm:"column:province;type:varchar(64);not null" json:"province"`        // 省份
	City            string `gorm:"column:city;type:varchar(64);not null" json:"city"`                // 城市
	District        string `gorm:"column:district;type:varchar(64);not null" json:"district"`        // 区/县
	DetailedAddress string `gorm:"column:detailed_address;type:varchar(255);not null" json:"detail"` // 详细地址
	PostalCode      string `gorm:"column:postal_code;type:varchar(20)" json:"postal_code"`           // 邮政编码
	IsDefault       bool   `gorm:"column:is_default;default:false" json:"is_default"`                // 是否默认地址
}

// NewUser 创建用户工厂方法
func NewUser(username, email, password, phone string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	if password == "" {
		return nil, fmt.Errorf("密码不能为空")
	}

	return &User{
		Username:  username,
		Email:     email,
		Password:  password,
		Phone:     phone,
		Status:    1,
		Addresses: []*Address{},
	}, nil
}

// UpdateProfile 更新个人信息
func (u *User) UpdateProfile(nickname, avatar string, gender int8, birthday *time.Time) {
	if nickname != "" {
		u.Nickname = nickname
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	if gender >= 0 {
		u.Gender = gender
	}
	if birthday != nil {
		u.Birthday = birthday
	}
	// GORM 会自动更新 UpdatedAt，但如果在事务中显式更新字段可能需要手动处理，
	// 这里作为领域方法，主要负责修改内存状态。
}

// ChangePassword 修改密码
func (u *User) ChangePassword(newPassword string) error {
	if newPassword == "" {
		return fmt.Errorf("新密码不能为空")
	}
	u.Password = newPassword
	return nil
}

// AddAddress 添加地址
func (u *User) AddAddress(address *Address) error {
	// 关联ID在持久化时由GORM处理，或者在此处显式赋值（如果u.ID已知）
	if u.ID != 0 {
		address.UserID = u.ID
	}

	if address.IsDefault {
		u.setDefaultAddress(address.ID)
	}

	u.Addresses = append(u.Addresses, address)
	return nil
}

// UpdateAddress 更新地址
func (u *User) UpdateAddress(addressID uint, address *Address) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			// 更新字段
			u.Addresses[i].RecipientName = address.RecipientName
			u.Addresses[i].PhoneNumber = address.PhoneNumber
			u.Addresses[i].Province = address.Province
			u.Addresses[i].City = address.City
			u.Addresses[i].District = address.District
			u.Addresses[i].DetailedAddress = address.DetailedAddress
			u.Addresses[i].PostalCode = address.PostalCode

			if address.IsDefault {
				u.setDefaultAddress(addressID)
				u.Addresses[i].IsDefault = true
			} else {
				u.Addresses[i].IsDefault = false
			}

			return nil
		}
	}
	return fmt.Errorf("地址未找到")
}

// RemoveAddress 移除地址
func (u *User) RemoveAddress(addressID uint) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("地址未找到")
}

// SetDefaultAddress 设置默认地址
func (u *User) SetDefaultAddress(addressID uint) error {
	return u.setDefaultAddress(addressID)
}

func (u *User) setDefaultAddress(addressID uint) error {
	found := false
	// 先取消所有默认地址
	for _, addr := range u.Addresses {
		if addr.ID == addressID {
			found = true
		}
		addr.IsDefault = false
	}

	if !found && addressID != 0 {
		return fmt.Errorf("地址未找到")
	}

	// 设置新的默认地址
	for _, addr := range u.Addresses {
		if addr.ID == addressID {
			addr.IsDefault = true
			return nil
		}
	}

	return nil
}

// GetDefaultAddress 获取默认地址
func (u *User) GetDefaultAddress() *Address {
	for _, addr := range u.Addresses {
		if addr.IsDefault {
			return addr
		}
	}
	return nil
}

// NewAddress 创建地址工厂方法
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
