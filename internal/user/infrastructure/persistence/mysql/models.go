package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
	"gorm.io/gorm"
)

// UserModel 用户表模型
type UserModel struct {
	gorm.Model
	Username  string         `gorm:"column:username;type:varchar(255);uniqueIndex;not null"`
	Email     string         `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	FullName  string         `gorm:"column:full_name;type:varchar(100)"`
	Password  string         `gorm:"column:password;type:varchar(255);not null"`
	Phone     string         `gorm:"column:phone;type:varchar(20);index"`
	Nickname  string         `gorm:"column:nickname;type:varchar(255)"`
	Avatar    string         `gorm:"column:avatar;type:varchar(1024)"`
	Gender    int8           `gorm:"column:gender;type:tinyint;default:0"`
	Birthday  *time.Time     `gorm:"column:birthday;type:date"`
	Status    int8           `gorm:"column:status;type:tinyint;default:1"`
	Addresses []AddressModel `gorm:"foreignKey:UserID"`
}

func (UserModel) TableName() string { return "users" }

// AddressModel 地址表模型
type AddressModel struct {
	gorm.Model
	UserID          uint   `gorm:"column:user_id;index;not null"`
	RecipientName   string `gorm:"column:recipient_name;type:varchar(255);not null"`
	PhoneNumber     string `gorm:"column:phone_number;type:varchar(20);not null"`
	Province        string `gorm:"column:province;type:varchar(64);not null"`
	City            string `gorm:"column:city;type:varchar(64);not null"`
	District        string `gorm:"column:district;type:varchar(64);not null"`
	DetailedAddress string `gorm:"column:detailed_address;type:varchar(255);not null"`
	PostalCode      string `gorm:"column:postal_code;type:varchar(20)"`
	IsDefault       bool   `gorm:"column:is_default;default:false"`
}

func (AddressModel) TableName() string { return "user_addresses" }

// mapping helpers

func toUserModel(u *domain.User) *UserModel {
	if u == nil {
		return nil
	}
	password := u.Password
	if password == "" {
		password = u.PasswordHash
	}
	return &UserModel{
		Model: gorm.Model{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Password: password,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Gender:   int8(u.Gender),
		Birthday: u.Birthday,
		Status:   int8(u.Status),
	}
}

func toUser(m *UserModel) *domain.User {
	if m == nil {
		return nil
	}
	user := &domain.User{
		Model: gorm.Model{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		Username:     m.Username,
		Email:        m.Email,
		FullName:     m.FullName,
		Password:     m.Password,
		PasswordHash: m.Password,
		Phone:        m.Phone,
		Nickname:     m.Nickname,
		Avatar:       m.Avatar,
		Gender:       domain.Gender(m.Gender),
		Birthday:     m.Birthday,
		Status:       domain.UserStatus(m.Status),
	}
	if len(m.Addresses) > 0 {
		user.Addresses = make([]*domain.Address, len(m.Addresses))
		for i := range m.Addresses {
			user.Addresses[i] = toAddress(&m.Addresses[i])
		}
	}
	return user
}

func toAddressModel(a *domain.Address) *AddressModel {
	if a == nil {
		return nil
	}
	recipientName := a.RecipientName
	if recipientName == "" {
		recipientName = a.Contact
	}
	phoneNumber := a.PhoneNumber
	if phoneNumber == "" {
		phoneNumber = a.Phone
	}
	detailedAddress := a.DetailedAddress
	if detailedAddress == "" {
		detailedAddress = a.Detail
	}
	postalCode := a.PostalCode
	if postalCode == "" {
		postalCode = a.ZipCode
	}
	return &AddressModel{
		Model: gorm.Model{
			ID:        a.ID,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		},
		UserID:          a.UserID,
		RecipientName:   recipientName,
		PhoneNumber:     phoneNumber,
		Province:        a.Province,
		City:            a.City,
		District:        a.District,
		DetailedAddress: detailedAddress,
		PostalCode:      postalCode,
		IsDefault:       a.IsDefault,
	}
}

func toAddress(m *AddressModel) *domain.Address {
	if m == nil {
		return nil
	}
	return &domain.Address{
		Model: gorm.Model{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		UserID:          m.UserID,
		Contact:         m.RecipientName,
		Phone:           m.PhoneNumber,
		RecipientName:   m.RecipientName,
		PhoneNumber:     m.PhoneNumber,
		Province:        m.Province,
		City:            m.City,
		District:        m.District,
		Detail:          m.DetailedAddress,
		DetailedAddress: m.DetailedAddress,
		ZipCode:         m.PostalCode,
		PostalCode:      m.PostalCode,
		IsDefault:       m.IsDefault,
	}
}
