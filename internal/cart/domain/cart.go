package domain

import (
	"gorm.io/gorm" // 导入GORM库。
)

// Cart 实体是购物车模块的聚合根。
// 它代表一个用户的购物车，包含了用户ID和购物车中的所有商品项。
type Cart struct {
	gorm.Model                    // 嵌入gorm.Model,包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	UserID            uint64      `gorm:"not null;uniqueIndex;comment:用户ID" json:"user_id"`              // 用户ID,唯一索引,不允许为空,确保每个用户只有一个购物车。
	AppliedCouponCode string      `gorm:"type:varchar(100);comment:已应用的优惠券码" json:"applied_coupon_code"` // 已应用的优惠券码。
	Items             []*CartItem `gorm:"foreignKey:CartID" json:"items"`                                // 购物车中的商品项列表,一对多关系。
}

// CartItem 实体代表购物车中的一个商品项。
type CartItem struct {
	gorm.Model              // 嵌入gorm.Model。
	CartID          uint64  `gorm:"not null;index;comment:购物车ID" json:"cart_id"`                 // 关联的购物车ID，索引字段，不允许为空。
	ProductID       uint64  `gorm:"not null;comment:商品ID" json:"product_id"`                     // 商品ID。
	SkuID           uint64  `gorm:"not null;index;comment:SKU ID" json:"sku_id"`                 // SKU ID，索引字段。
	ProductName     string  `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"` // 商品名称。
	SkuName         string  `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`    // SKU名称（例如，颜色、尺码等属性组合）。
	Price           float64 `gorm:"type:decimal(10,2);not null;comment:价格" json:"price"`         // 商品单价。
	Quantity        int32   `gorm:"not null;comment:数量" json:"quantity"`                         // 购买数量。
	ProductImageURL string  `gorm:"type:varchar(255);comment:商品图片URL" json:"product_image_url"`  // 商品图片URL。
}

// NewCart 创建并返回一个新的 Cart 实体实例。
// userID: 购物车所属的用户ID。
func NewCart(userID uint64) *Cart {
	return &Cart{
		UserID: userID,
		Items:  []*CartItem{}, // 初始化购物车商品项列表。
	}
}

// AddItem 将商品添加到购物车。
// 如果购物车中已存在相同SKU的商品，则更新其数量；否则添加新的商品项。
// productID: 商品ID。
// skuID: SKU ID。
// productName: 商品名称。
// skuName: SKU名称。
// price: 商品单价。
// quantity: 待添加的数量。
// imageURL: 商品图片URL。
func (c *Cart) AddItem(productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) {
	// 检查购物车中是否已存在该SKU的商品。
	for _, item := range c.Items {
		if item.SkuID == skuID {
			item.Quantity += quantity // 如果存在，则增加数量。
			return
		}
	}

	// 如果不存在，则添加一个新的商品项。
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

// UpdateItemQuantity 更新购物车中指定SKU商品的数量。
// skuID: 待更新商品的SKU ID。
// quantity: 更新后的商品数量。
func (c *Cart) UpdateItemQuantity(skuID uint64, quantity int32) {
	for _, item := range c.Items {
		if item.SkuID == skuID {
			item.Quantity = quantity // 找到商品并更新数量。
			return
		}
	}
}

// RemoveItem 从购物车中移除指定SKU的商品。
// skuID: 待移除商品的SKU ID。
func (c *Cart) RemoveItem(skuID uint64) {
	for i, item := range c.Items {
		if item.SkuID == skuID {
			// 从切片中移除元素。
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return
		}
	}
}

// Clear 清空购物车中的所有商品项。
func (c *Cart) Clear() {
	c.Items = []*CartItem{} // 将商品项切片重置为空。
}

// GetTotalPrice 计算购物车中所有商品的总价格。
func (c *Cart) GetTotalPrice() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity) // 累加每个商品项的总价。
	}
	return total
}

// GetTotalQuantity 计算购物车中所有商品的总数量。
func (c *Cart) GetTotalQuantity() int32 {
	var total int32
	for _, item := range c.Items {
		total += item.Quantity // 累加每个商品项的数量。
	}
	return total
}
