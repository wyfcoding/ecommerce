package domain

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// 地址类型常数
const (
	AddressTypeUnspecified = 0
	AddressTypeHome        = 1
	AddressTypeWork        = 2
	AddressTypeOther       = 3
)

// Address 用户收货地址实体
// 使用 GORM 的 BaseModel 进行统一字段管理
type Address struct {
	gorm.Model
	ID            string  `gorm:"type:varchar(64);uniqueIndex;not null" json:"id"`        // 业务主键
	UserID        int64   `gorm:"index:idx_user_default;not null" json:"user_id"`         // 用户ID
	RecipientName string  `gorm:"type:varchar(128);not null" json:"recipient_name"`       // 收件人姓名
	PhoneNumber   string  `gorm:"type:varchar(32);not null" json:"phone_number"`          // 手机号/联系电话
	Country       string  `gorm:"type:varchar(64)" json:"country"`                        // 国家
	Province      string  `gorm:"type:varchar(64)" json:"province"`                       // 省份/州
	City          string  `gorm:"type:varchar(64)" json:"city"`                           // 城市
	District      string  `gorm:"type:varchar(64)" json:"district"`                       // 区/县
	DetailAddress string  `gorm:"type:varchar(256);not null" json:"detail_address"`       // 详细地址
	PostalCode    string  `gorm:"type:varchar(32)" json:"postal_code"`                    // 邮政编码
	IsDefault     bool    `gorm:"index:idx_user_default;default:false" json:"is_default"` // 是否为默认地址
	Type          int32   `gorm:"type:int;default:0" json:"type"`                         // 地址类型 (家, 公司等)
	Latitude      float64 `gorm:"type:decimal(10,8)" json:"latitude"`                     // 纬度 (便于后续物流分派)
	Longitude     float64 `gorm:"type:decimal(11,8)" json:"longitude"`                    // 经度
}

// 领域错误
var (
	ErrAddressNotFound      = errors.New("address not found")
	ErrAddressLimitExceeded = errors.New("address limit exceeded")
	ErrInvalidAddress       = errors.New("invalid address data")
)

// AddressRepository 地址仓储接口
type AddressRepository interface {
	Save(ctx context.Context, addr *Address) error
	Update(ctx context.Context, addr *Address) error
	Delete(ctx context.Context, id string, userID int64) error
	FindByID(ctx context.Context, id string) (*Address, error)
	FindByUserID(ctx context.Context, userID int64) ([]*Address, error)
	ClearDefaultByUserID(ctx context.Context, userID int64) error
	SetDefault(ctx context.Context, id string, userID int64) error
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}
