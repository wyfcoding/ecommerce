package repository

import (
	"context"
	"ecommerce/internal/aftersales/biz"
	"ecommerce/internal/aftersales/model"
	"time"

	"gorm.io/gorm"
)

// Data 结构体持有数据库连接。
type Data struct {
	db *gorm.DB
}

// NewData 是 Data 结构体的构造函数。
func NewData(db *gorm.DB) *Data {
	return &Data{db: db}
}

// transaction 实现了 biz.Transaction 接口。
type transaction struct {
	db *gorm.DB
}

// NewTransaction 是 transaction 的构造函数。
func NewTransaction(data *Data) biz.Transaction {
	return &transaction{db: data.db}
}

// InTx 在一个数据库事务中执行函数。
func (t *transaction) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}

type aftersalesRepo struct {
	data *Data
}

func NewAftersalesRepo(data *Data) biz.AftersalesRepo {
	return &aftersalesRepo{data: data}
}

func (r *aftersalesRepo) CreateReturnOrder(ctx context.Context, order *biz.ReturnOrder) (*biz.ReturnOrder, error) {
	po := &model.ReturnOrder{
		OrderID:    order.OrderID,
		UserID:     order.UserID,
		Reason:     order.Reason,
		Status:     order.Status,
		RequestDate: order.RequestDate,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	order.ID = po.ID
	return order, nil
}

func (r *aftersalesRepo) GetReturnOrderByID(ctx context.Context, id uint) (*biz.ReturnOrder, error) {
	var po model.ReturnOrder
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.ReturnOrder{
		ID:          po.ID,
		OrderID:     po.OrderID,
		UserID:      po.UserID,
		Reason:      po.Reason,
		Status:      po.Status,
		RequestDate: po.RequestDate,
	}, nil
}

func (r *aftersalesRepo) UpdateReturnOrderStatus(ctx context.Context, id uint, status string) error {
	return r.data.db.WithContext(ctx).Model(&model.ReturnOrder{}).Where("id = ?", id).Update("status", status).Error
}

// ReturnOrder 是退货订单的数据库模型。
type ReturnOrder struct {
	gorm.Model
	OrderID     uint      `gorm:"not null;comment:订单ID" json:"orderId"`
	UserID      uint      `gorm:"not null;comment:用户ID" json:"userId"`
	Reason      string    `gorm:"type:varchar(255);comment:退货原因" json:"reason"`
	Status      string    `gorm:"type:varchar(50);comment:退货状态" json:"status"` // 例如: "pending", "approved", "rejected", "completed"
	RequestDate time.Time `gorm:"comment:申请日期" json:"requestDate"`
}

func (ReturnOrder) TableName() string {
	return "return_orders"
}