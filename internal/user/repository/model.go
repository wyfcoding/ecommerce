package data

import (
	"time"

	"gorm.io/gorm"
)

// User 是用户的数据库模型（Data Model），它直接映射到数据库中的 `users` 表结构。
type User struct {
	gorm.Model
	Username string    `gorm:"uniqueIndex;not null;type:varchar(64);comment:用户名" json:"username"`
	Password string    `gorm:"not null;type:varchar(255);comment:密码" json:"password"`
	Nickname string    `gorm:"type:varchar(64);comment:昵称" json:"nickname"`
	Avatar   string    `gorm:"type:varchar(255);comment:头像URL" json:"avatar"`
	Gender   int32     `gorm:"type:tinyint;comment:性别 0:未知 1:男 2:女" json:"gender"`
	Birthday time.Time `gorm:"comment:生日" json:"birthday"`
	Phone    string    `gorm:"uniqueIndex;type:varchar(20);comment:手机号" json:"phone"`
	Email    string    `gorm:"uniqueIndex;type:varchar(100);comment:邮箱" json:"email"`
}

func (User) TableName() string {
	return "users"
}

// Address 是用户的收货地址模型，对应数据库中的 `addresses` 表。
type Address struct {
	gorm.Model
	UserID    uint   `json:"userId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	IsDefault bool   `json:"isDefault"`
}

// TableName 指定 Address 模型对应的数据库表名。
func (Address) TableName() string {
	return "addresses"
}
