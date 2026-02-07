package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"gorm.io/gorm"
)

// PointsProductModel 积分商品写模型。
type PointsProductModel struct {
	gorm.Model
	Name         string                 `gorm:"column:name;type:varchar(255);not null;comment:商品名称"`
	Description  string                 `gorm:"column:description;type:text;comment:商品描述"`
	ImageURL     string                 `gorm:"column:image_url;type:varchar(255);comment:图片URL"`
	Points       int64                  `gorm:"column:points;not null;comment:所需积分"`
	Stock        int32                  `gorm:"column:stock;not null;default:0;comment:库存"`
	SoldCount    int32                  `gorm:"column:sold_count;not null;default:0;comment:已售数量"`
	LimitPerUser int32                  `gorm:"column:limit_per_user;not null;default:0;comment:每人限购"`
	Status       domain.PointsProductStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
}

// PointsOrderModel 积分订单写模型。
type PointsOrderModel struct {
	gorm.Model
	OrderNo     string                 `gorm:"column:order_no;type:varchar(64);uniqueIndex;not null;comment:订单编号"`
	UserID      uint64                 `gorm:"column:user_id;not null;index;comment:用户ID"`
	ProductID   uint64                 `gorm:"column:product_id;not null;index;comment:商品ID"`
	ProductName string                 `gorm:"column:product_name;type:varchar(255);not null;comment:商品名称"`
	Quantity    int32                  `gorm:"column:quantity;not null;comment:数量"`
	Points      int64                  `gorm:"column:points;not null;comment:单价积分"`
	TotalPoints int64                  `gorm:"column:total_points;not null;comment:总积分"`
	Status      domain.PointsOrderStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
	Address     string                 `gorm:"column:address;type:varchar(255);comment:收货地址"`
	Phone       string                 `gorm:"column:phone;type:varchar(20);comment:联系电话"`
	Receiver    string                 `gorm:"column:receiver;type:varchar(64);comment:收货人"`
	ShippedAt   *time.Time             `gorm:"column:shipped_at;comment:发货时间"`
	CompletedAt *time.Time             `gorm:"column:completed_at;comment:完成时间"`
}

// PointsAccountModel 积分账户写模型。
type PointsAccountModel struct {
	gorm.Model
	UserID      uint64 `gorm:"column:user_id;uniqueIndex;not null;comment:用户ID"`
	TotalPoints int64  `gorm:"column:total_points;not null;default:0;comment:总积分"`
	UsedPoints  int64  `gorm:"column:used_points;not null;default:0;comment:已用积分"`
}

// PointsTransactionModel 积分流水写模型。
type PointsTransactionModel struct {
	gorm.Model
	UserID      uint64 `gorm:"column:user_id;not null;index;comment:用户ID"`
	Type        string `gorm:"column:type;type:varchar(32);not null;comment:类型"`
	Points      int64  `gorm:"column:points;not null;comment:变动积分"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
	RefID       string `gorm:"column:ref_id;type:varchar(64);index;comment:关联ID"`
}

func (PointsProductModel) TableName() string     { return "points_products" }
func (PointsOrderModel) TableName() string       { return "points_orders" }
func (PointsAccountModel) TableName() string     { return "points_accounts" }
func (PointsTransactionModel) TableName() string { return "points_transactions" }

func toProductModel(product *domain.PointsProduct) *PointsProductModel {
	if product == nil {
		return nil
	}
	return &PointsProductModel{
		Model: gorm.Model{
			ID:        product.ID,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
		},
		Name:         product.Name,
		Description:  product.Description,
		ImageURL:     product.ImageURL,
		Points:       product.Points,
		Stock:        product.Stock,
		SoldCount:    product.SoldCount,
		LimitPerUser: product.LimitPerUser,
		Status:       product.Status,
	}
}

func toProduct(model *PointsProductModel) *domain.PointsProduct {
	if model == nil {
		return nil
	}
	return &domain.PointsProduct{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		Name:         model.Name,
		Description:  model.Description,
		ImageURL:     model.ImageURL,
		Points:       model.Points,
		Stock:        model.Stock,
		SoldCount:    model.SoldCount,
		LimitPerUser: model.LimitPerUser,
		Status:       model.Status,
	}
}

func toOrderModel(order *domain.PointsOrder) *PointsOrderModel {
	if order == nil {
		return nil
	}
	return &PointsOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		OrderNo:     order.OrderNo,
		UserID:      order.UserID,
		ProductID:   order.ProductID,
		ProductName: order.ProductName,
		Quantity:    order.Quantity,
		Points:      order.Points,
		TotalPoints: order.TotalPoints,
		Status:      order.Status,
		Address:     order.Address,
		Phone:       order.Phone,
		Receiver:    order.Receiver,
		ShippedAt:   order.ShippedAt,
		CompletedAt: order.CompletedAt,
	}
}

func toOrder(model *PointsOrderModel) *domain.PointsOrder {
	if model == nil {
		return nil
	}
	return &domain.PointsOrder{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		OrderNo:     model.OrderNo,
		UserID:      model.UserID,
		ProductID:   model.ProductID,
		ProductName: model.ProductName,
		Quantity:    model.Quantity,
		Points:      model.Points,
		TotalPoints: model.TotalPoints,
		Status:      model.Status,
		Address:     model.Address,
		Phone:       model.Phone,
		Receiver:    model.Receiver,
		ShippedAt:   model.ShippedAt,
		CompletedAt: model.CompletedAt,
	}
}

func toAccountModel(account *domain.PointsAccount) *PointsAccountModel {
	if account == nil {
		return nil
	}
	return &PointsAccountModel{
		Model: gorm.Model{
			ID:        account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		UserID:      account.UserID,
		TotalPoints: account.TotalPoints,
		UsedPoints:  account.UsedPoints,
	}
}

func toAccount(model *PointsAccountModel) *domain.PointsAccount {
	if model == nil {
		return nil
	}
	return &domain.PointsAccount{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		UserID:      model.UserID,
		TotalPoints: model.TotalPoints,
		UsedPoints:  model.UsedPoints,
	}
}

func toTransactionModel(tx *domain.PointsTransaction) *PointsTransactionModel {
	if tx == nil {
		return nil
	}
	return &PointsTransactionModel{
		Model: gorm.Model{
			ID:        tx.ID,
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
		},
		UserID:      tx.UserID,
		Type:        tx.Type,
		Points:      tx.Points,
		Description: tx.Description,
		RefID:       tx.RefID,
	}
}

func toTransaction(model *PointsTransactionModel) *domain.PointsTransaction {
	if model == nil {
		return nil
	}
	return &domain.PointsTransaction{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		UserID:      model.UserID,
		Type:        model.Type,
		Points:      model.Points,
		Description: model.Description,
		RefID:       model.RefID,
	}
}
