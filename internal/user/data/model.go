package data

import (
	"gorm.io/gorm"
)

// User 是用户的数据库模型（Data Model），它直接映射到数据库中的 `users` 表结构。
type User struct {
	gorm.Model
	UserID   uint64 `gorm:"uniqueIndex;not null;comment:用户ID"`
	Username string `gorm:"uniqueIndex;not null;comment:用户名"`
	Password string `gorm:"not null;comment:加密后的密码"`

	Nickname *string `gorm:"comment:昵称"`
	Avatar   *string `gorm:"comment:头像URL"`
	Gender   *int32  `gorm:"comment:性别, 1-男, 2-女, 0-未知"`
}

func (User) TableName() string {
	return "users"
}

// Address 是用户的收货地址模型，对应数据库中的 `addresses` 表。
type Address struct {
	gorm.Model
	UserID          uint64 `gorm:"index;not null"`
	Name            string `gorm:"type:varchar(50);not null"`
	Phone           string `gorm:"type:varchar(20);not null"`
	Province        string `gorm:"type:varchar(50);not null"`
	City            string `gorm:"type:varchar(50);not null"`
	District        string `gorm:"type:varchar(50);not null"`
	DetailedAddress string `gorm:"type:varchar(255);not null"`
	IsDefault       bool   `gorm:"not null;default:false"`
}

// TableName 指定 Address 模型对应的数据库表名。
func (Address) TableName() string {
	return "addresses"
}
