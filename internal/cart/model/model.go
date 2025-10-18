package model

import (
	"time"
)

// Cart 购物车模型
// gorm.Model 包含了 CreatedAt, UpdatedAt, DeletedAt, ID 字段
type Cart struct {
	ID        uint      `gorm:"primarykey" json:"id"` // 购物车唯一ID
	UserID    uint      `gorm:"not null;index" json:"user_id"` // 用户ID，建立索引以加速查询
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"` // 购物车中的商品项列表
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// CartItem 购物车商品项模型
type CartItem struct {
	ID          uint    `gorm:"primarykey" json:"id"` // 购物车项唯一ID
	CartID      uint    `gorm:"not null;index" json:"cart_id"` // 所属购物车的ID
	ProductID   uint    `gorm:"not null;index" json:"product_id"` // 商品ID
	Quantity    int     `gorm:"not null" json:"quantity"` // 商品数量
	Price       float64 `gorm:"type:decimal(10,2)" json:"price"` // 商品加入购物车时的单价
	ProductName string  `gorm:"type:varchar(255)" json:"product_name"` // 商品名称 (冗余字段，用于展示)
	ProductImg  string  `gorm:"type:varchar(255)" json:"product_img"` // 商品图片URL (冗余字段，用于展示)
	AddedAt     time.Time `gorm:"not null" json:"added_at"` // 商品添加时间
}

// TableName 自定义Cart模型的表名
func (Cart) TableName() string {
	return "carts"
}

// TableName 自定义CartItem模型的表名
func (CartItem) TableName() string {
	return "cart_items"
}