package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"gorm.io/gorm"
)

// FlashsaleModel 秒杀活动写模型。
type FlashsaleModel struct {
	gorm.Model
	Name          string                 `gorm:"column:name;type:varchar(255);not null;comment:活动名称"`
	ProductID     uint64                 `gorm:"column:product_id;not null;index;comment:商品ID"`
	SkuID         uint64                 `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	OriginalPrice int64                  `gorm:"column:original_price;not null;comment:原价(分)"`
	FlashPrice    int64                  `gorm:"column:flash_price;not null;comment:秒杀价(分)"`
	TotalStock    int32                  `gorm:"column:total_stock;not null;comment:总库存"`
	SoldCount     int32                  `gorm:"column:sold_count;not null;default:0;comment:已售数量"`
	LimitPerUser  int32                  `gorm:"column:limit_per_user;not null;default:1;comment:每人限购数量"`
	StartTime     time.Time              `gorm:"column:start_time;not null;comment:开始时间"`
	EndTime       time.Time              `gorm:"column:end_time;not null;comment:结束时间"`
	Status        domain.FlashsaleStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
	Description   string                 `gorm:"column:description;type:text;comment:描述"`
	Version       int64                  `gorm:"column:version;not null;default:1;comment:乐观锁版本"`
}

// FlashsaleOrderModel 秒杀订单写模型。
type FlashsaleOrderModel struct {
	gorm.Model
	FlashsaleID uint64                      `gorm:"column:flashsale_id;not null;index;comment:秒杀活动ID"`
	UserID      uint64                      `gorm:"column:user_id;not null;index;comment:用户ID"`
	ProductID   uint64                      `gorm:"column:product_id;not null;comment:商品ID"`
	SkuID       uint64                      `gorm:"column:sku_id;not null;comment:SKU ID"`
	Quantity    int32                       `gorm:"column:quantity;not null;comment:数量"`
	Price       int64                       `gorm:"column:price;not null;comment:单价(分)"`
	TotalAmount int64                       `gorm:"column:total_amount;not null;comment:总金额(分)"`
	Status      domain.FlashsaleOrderStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
}

func (FlashsaleModel) TableName() string      { return "flashsales" }
func (FlashsaleOrderModel) TableName() string { return "flashsale_orders" }

func toFlashsaleModel(flashsale *domain.Flashsale) *FlashsaleModel {
	if flashsale == nil {
		return nil
	}
	return &FlashsaleModel{
		Model: gorm.Model{
			ID:        flashsale.ID,
			CreatedAt: flashsale.CreatedAt,
			UpdatedAt: flashsale.UpdatedAt,
		},
		Name:          flashsale.Name,
		ProductID:     flashsale.ProductID,
		SkuID:         flashsale.SkuID,
		OriginalPrice: flashsale.OriginalPrice,
		FlashPrice:    flashsale.FlashPrice,
		TotalStock:    flashsale.TotalStock,
		SoldCount:     flashsale.SoldCount,
		LimitPerUser:  flashsale.LimitPerUser,
		StartTime:     flashsale.StartTime,
		EndTime:       flashsale.EndTime,
		Status:        flashsale.Status,
		Description:   flashsale.Description,
		Version:       flashsale.Version,
	}
}

func toFlashsale(model *FlashsaleModel) *domain.Flashsale {
	if model == nil {
		return nil
	}
	return &domain.Flashsale{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Name:          model.Name,
		ProductID:     model.ProductID,
		SkuID:         model.SkuID,
		OriginalPrice: model.OriginalPrice,
		FlashPrice:    model.FlashPrice,
		TotalStock:    model.TotalStock,
		SoldCount:     model.SoldCount,
		LimitPerUser:  model.LimitPerUser,
		StartTime:     model.StartTime,
		EndTime:       model.EndTime,
		Status:        model.Status,
		Description:   model.Description,
		Version:       model.Version,
	}
}

func toFlashsaleOrderModel(order *domain.FlashsaleOrder) *FlashsaleOrderModel {
	if order == nil {
		return nil
	}
	return &FlashsaleOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		FlashsaleID: order.FlashsaleID,
		UserID:      order.UserID,
		ProductID:   order.ProductID,
		SkuID:       order.SkuID,
		Quantity:    order.Quantity,
		Price:       order.Price,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
	}
}

func toFlashsaleOrder(model *FlashsaleOrderModel) *domain.FlashsaleOrder {
	if model == nil {
		return nil
	}
	return &domain.FlashsaleOrder{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		FlashsaleID: model.FlashsaleID,
		UserID:      model.UserID,
		ProductID:   model.ProductID,
		SkuID:       model.SkuID,
		Quantity:    model.Quantity,
		Price:       model.Price,
		TotalAmount: model.TotalAmount,
		Status:      model.Status,
	}
}
