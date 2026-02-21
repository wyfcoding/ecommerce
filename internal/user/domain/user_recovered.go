package domain

import (
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// RecoveredUser 聚合根（救回版本：包含地址管理级联逻辑）
type RecoveredUser struct {
	gorm.Model
	database.BaseEntity
	Username  string             `gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	Email     string             `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Addresses []RecoveredAddress `gorm:"foreignKey:UserID"`
}

type RecoveredAddress struct {
	gorm.Model
	UserID    uint   `gorm:"column:user_id;index;not null"`
	Contact   string `gorm:"column:contact;type:varchar(64)"`
	IsDefault bool   `gorm:"column:is_default;default:false"`
}

func (u *RecoveredUser) AddAddress(addr RecoveredAddress) {
	if addr.IsDefault {
		for i := range u.Addresses {
			u.Addresses[i].IsDefault = false
		}
	}
	u.Addresses = append(u.Addresses, addr)
}
