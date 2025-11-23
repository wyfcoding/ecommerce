package entity

import (
	"gorm.io/gorm"
)

// Cart 购物车聚合根
type Cart struct {
	gorm.Model
	UserID uint64      `gorm:"not null;uniqueIndex;comment:用户ID" json:"user_id"`
	Items  []*CartItem `gorm:"foreignKey:CartID" json:"items"`
}

// CartItem 购物车项实体
type CartItem struct {
	gorm.Model
	CartID          uint64  `gorm:"not null;index;comment:购物车ID" json:"cart_id"`
	ProductID       uint64  `gorm:"not null;comment:商品ID" json:"product_id"`
	SkuID           uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	ProductName     string  `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`
	SkuName         string  `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`
	Price           float64 `gorm:"type:decimal(10,2);not null;comment:价格" json:"price"`
	Quantity        int32   `gorm:"not null;comment:数量" json:"quantity"`
	ProductImageURL string  `gorm:"type:varchar(255);comment:商品图片URL" json:"product_image_url"`
}

// NewCart 创建购物车
func NewCart(userID uint64) *Cart {
	return &Cart{
		UserID: userID,
		Items:  []*CartItem{},
	}
}

// AddItem 添加商品到购物车
func (c *Cart) AddItem(productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) {
	// 检查是否已存在相同的SKU
	for _, item := range c.Items {
		if item.SkuID == skuID {
			item.Quantity += quantity
			return
		}
	}

	// 添加新项
	item := &CartItem{
		ProductID:       productID,
		SkuID:           skuID,
		ProductName:     productName,
		SkuName:         skuName,
		Price:           price,
		Quantity:        quantity,
		ProductImageURL: imageURL,
	}
	c.Items = append(c.Items, item)
}

// UpdateItemQuantity 更新购物车项数量
func (c *Cart) UpdateItemQuantity(skuID uint64, quantity int32) {
	for _, item := range c.Items {
		if item.SkuID == skuID {
			item.Quantity = quantity
			return
		}
	}
}

// RemoveItem 移除购物车项
func (c *Cart) RemoveItem(skuID uint64) {
	for i, item := range c.Items {
		if item.SkuID == skuID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return
		}
	}
}

// Clear 清空购物车
func (c *Cart) Clear() {
	c.Items = []*CartItem{}
}

// GetTotalPrice 获取总价格
func (c *Cart) GetTotalPrice() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

// GetTotalQuantity 获取总数量
func (c *Cart) GetTotalQuantity() int32 {
	var total int32
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}
