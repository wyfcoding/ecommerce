package mysql

import (
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"gorm.io/gorm"
)

// OrderModel 订单写模型（持久化专用）。
type OrderModel struct {
	gorm.Model
	OrderNo              string               `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号"`
	Version              int64                `gorm:"not null;default:0;comment:事件版本号(用于事件溯源并发控制)"`
	UserID               uint64               `gorm:"index;not null;comment:用户ID"`
	Status               pb.OrderStatus       `gorm:"type:tinyint;not null;default:1;comment:订单状态"`
	PaymentStatus        pb.PaymentStatus     `gorm:"type:tinyint;not null;default:1;comment:支付状态"`
	ShippingStatus       pb.ShippingStatus    `gorm:"type:tinyint;not null;default:1;comment:物流状态"`
	TotalAmount          int64                `gorm:"not null;comment:订单总金额(分)"`
	ActualAmount         int64                `gorm:"not null;comment:实际支付金额(分)"`
	ShippingFee          int64                `gorm:"not null;default:0;comment:运费(分)"`
	DiscountAmount       int64                `gorm:"not null;default:0;comment:优惠金额(分)"`
	PaymentMethod        string               `gorm:"type:varchar(32);comment:支付方式"`
	PaymentTransactionID string               `gorm:"type:varchar(128);comment:支付流水号"`
	Remark               string               `gorm:"type:varchar(255);comment:订单备注"`
	TrackingNumber       string               `gorm:"type:varchar(128);comment:快递单号"`
	LogisticsCompany     string               `gorm:"type:varchar(128);comment:物流公司"`
	RefundAmount         int64                `gorm:"not null;default:0;comment:退款金额(分)"`
	RefundReason         string               `gorm:"type:varchar(255);comment:退款原因"`
	ShippingAddress      ShippingAddressModel `gorm:"embedded;embeddedPrefix:shipping_"`
	Items                []OrderItemModel     `gorm:"foreignKey:OrderID"`
	Logs                 []OrderLogModel      `gorm:"foreignKey:OrderID"`
	PaidAt               *time.Time           `gorm:"comment:支付时间"`
	ShippedAt            *time.Time           `gorm:"comment:发货时间"`
	DeliveredAt          *time.Time           `gorm:"comment:送达时间"`
	CompletedAt          *time.Time           `gorm:"comment:完成时间"`
	CancelledAt          *time.Time           `gorm:"comment:取消时间"`
}

func (OrderModel) TableName() string {
	return "orders"
}

// ShippingAddressModel 订单地址（持久化专用）。
type ShippingAddressModel struct {
	RecipientName   string  `gorm:"type:varchar(64);comment:收货人姓名"`
	PhoneNumber     string  `gorm:"type:varchar(20);comment:手机号"`
	Province        string  `gorm:"type:varchar(64);comment:省份"`
	City            string  `gorm:"type:varchar(64);comment:城市"`
	District        string  `gorm:"type:varchar(64);comment:区县"`
	DetailedAddress string  `gorm:"type:varchar(255);comment:详细地址"`
	PostalCode      string  `gorm:"type:varchar(20);comment:邮政编码"`
	Lat             float64 `gorm:"type:decimal(10,6);comment:纬度"`
	Lon             float64 `gorm:"type:decimal(10,6);comment:经度"`
}

// OrderItemModel 订单明细（持久化专用）。
type OrderItemModel struct {
	gorm.Model
	OrderID         uint64 `gorm:"index;not null;comment:订单ID"`
	ProductID       uint64 `gorm:"not null;comment:商品ID"`
	SkuID           uint64 `gorm:"not null;comment:SKU ID"`
	ProductName     string `gorm:"type:varchar(255);not null;comment:商品名称"`
	SkuName         string `gorm:"type:varchar(255);not null;comment:SKU名称"`
	ProductImageURL string `gorm:"type:varchar(255);comment:商品图片URL"`
	Price           int64  `gorm:"not null;comment:单价(分)"`
	Quantity        int32  `gorm:"not null;comment:数量"`
	TotalPrice      int64  `gorm:"not null;comment:总价(分)"`
}

func (OrderItemModel) TableName() string {
	return "order_items"
}

// OrderLogModel 订单日志（持久化专用）。
type OrderLogModel struct {
	gorm.Model
	OrderID   uint64 `gorm:"index;not null;comment:订单ID"`
	Operator  string `gorm:"type:varchar(64);not null;comment:操作人"`
	Action    string `gorm:"type:varchar(64);not null;comment:操作动作"`
	OldStatus string `gorm:"type:varchar(32);comment:旧状态"`
	NewStatus string `gorm:"type:varchar(32);comment:新状态"`
	Remark    string `gorm:"type:varchar(255);comment:备注"`
}

func (OrderLogModel) TableName() string {
	return "order_logs"
}

func toOrderModel(order *domain.Order) *OrderModel {
	if order == nil {
		return nil
	}

	model := &OrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		OrderNo:              order.OrderNo,
		Version:              order.Version,
		UserID:               order.UserID,
		Status:               order.Status,
		PaymentStatus:        order.PaymentStatus,
		ShippingStatus:       order.ShippingStatus,
		TotalAmount:          order.TotalAmount,
		ActualAmount:         order.ActualAmount,
		ShippingFee:          order.ShippingFee,
		DiscountAmount:       order.DiscountAmount,
		PaymentMethod:        order.PaymentMethod,
		PaymentTransactionID: order.PaymentTransactionID,
		Remark:               order.Remark,
		TrackingNumber:       order.TrackingNumber,
		LogisticsCompany:     order.LogisticsCompany,
		RefundAmount:         order.RefundAmount,
		RefundReason:         order.RefundReason,
		PaidAt:               order.PaidAt,
		ShippedAt:            order.ShippedAt,
		DeliveredAt:          order.DeliveredAt,
		CompletedAt:          order.CompletedAt,
		CancelledAt:          order.CancelledAt,
	}

	if order.ShippingAddress != nil {
		model.ShippingAddress = ShippingAddressModel{
			RecipientName:   order.ShippingAddress.RecipientName,
			PhoneNumber:     order.ShippingAddress.PhoneNumber,
			Province:        order.ShippingAddress.Province,
			City:            order.ShippingAddress.City,
			District:        order.ShippingAddress.District,
			DetailedAddress: order.ShippingAddress.DetailedAddress,
			PostalCode:      order.ShippingAddress.PostalCode,
			Lat:             order.ShippingAddress.Lat,
			Lon:             order.ShippingAddress.Lon,
		}
	}

	model.Items = make([]OrderItemModel, 0, len(order.Items))
	for _, item := range order.Items {
		if item == nil {
			continue
		}
		model.Items = append(model.Items, OrderItemModel{
			Model: gorm.Model{
				ID:        item.ID,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			},
			OrderID:         item.OrderID,
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			ProductImageURL: item.ProductImageURL,
			Price:           item.Price,
			Quantity:        item.Quantity,
			TotalPrice:      item.TotalPrice,
		})
	}

	model.Logs = make([]OrderLogModel, 0, len(order.Logs))
	for _, log := range order.Logs {
		if log == nil {
			continue
		}
		model.Logs = append(model.Logs, OrderLogModel{
			Model: gorm.Model{
				ID:        log.ID,
				CreatedAt: log.CreatedAt,
				UpdatedAt: log.UpdatedAt,
			},
			OrderID:   log.OrderID,
			Operator:  log.Operator,
			Action:    log.Action,
			OldStatus: log.OldStatus,
			NewStatus: log.NewStatus,
			Remark:    log.Remark,
		})
	}

	return model
}

func toDomainOrder(model *OrderModel) *domain.Order {
	if model == nil {
		return nil
	}

	order := &domain.Order{
		ID:                   model.ID,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
		OrderNo:              model.OrderNo,
		Version:              model.Version,
		UserID:               model.UserID,
		Status:               model.Status,
		PaymentStatus:        model.PaymentStatus,
		ShippingStatus:       model.ShippingStatus,
		TotalAmount:          model.TotalAmount,
		ActualAmount:         model.ActualAmount,
		ShippingFee:          model.ShippingFee,
		DiscountAmount:       model.DiscountAmount,
		PaymentMethod:        model.PaymentMethod,
		PaymentTransactionID: model.PaymentTransactionID,
		Remark:               model.Remark,
		TrackingNumber:       model.TrackingNumber,
		LogisticsCompany:     model.LogisticsCompany,
		RefundAmount:         model.RefundAmount,
		RefundReason:         model.RefundReason,
		PaidAt:               model.PaidAt,
		ShippedAt:            model.ShippedAt,
		DeliveredAt:          model.DeliveredAt,
		CompletedAt:          model.CompletedAt,
		CancelledAt:          model.CancelledAt,
	}

	if hasShippingAddress(model.ShippingAddress) {
		order.ShippingAddress = &domain.ShippingAddress{
			RecipientName:   model.ShippingAddress.RecipientName,
			PhoneNumber:     model.ShippingAddress.PhoneNumber,
			Province:        model.ShippingAddress.Province,
			City:            model.ShippingAddress.City,
			District:        model.ShippingAddress.District,
			DetailedAddress: model.ShippingAddress.DetailedAddress,
			PostalCode:      model.ShippingAddress.PostalCode,
			Lat:             model.ShippingAddress.Lat,
			Lon:             model.ShippingAddress.Lon,
		}
	}

	order.Items = make([]*domain.OrderItem, 0, len(model.Items))
	for _, item := range model.Items {
		itemCopy := item
		order.Items = append(order.Items, &domain.OrderItem{
			ID:              itemCopy.ID,
			CreatedAt:       itemCopy.CreatedAt,
			UpdatedAt:       itemCopy.UpdatedAt,
			OrderID:         itemCopy.OrderID,
			ProductID:       itemCopy.ProductID,
			SkuID:           itemCopy.SkuID,
			ProductName:     itemCopy.ProductName,
			SkuName:         itemCopy.SkuName,
			ProductImageURL: itemCopy.ProductImageURL,
			Price:           itemCopy.Price,
			Quantity:        itemCopy.Quantity,
			TotalPrice:      itemCopy.TotalPrice,
		})
	}

	order.Logs = make([]*domain.OrderLog, 0, len(model.Logs))
	for _, log := range model.Logs {
		logCopy := log
		order.Logs = append(order.Logs, &domain.OrderLog{
			ID:        logCopy.ID,
			CreatedAt: logCopy.CreatedAt,
			UpdatedAt: logCopy.UpdatedAt,
			OrderID:   logCopy.OrderID,
			Operator:  logCopy.Operator,
			Action:    logCopy.Action,
			OldStatus: logCopy.OldStatus,
			NewStatus: logCopy.NewStatus,
			Remark:    logCopy.Remark,
		})
	}

	fillDerivedStatuses(order)
	order.InitFSM()
	return order
}

func hasShippingAddress(addr ShippingAddressModel) bool {
	return addr.RecipientName != "" ||
		addr.PhoneNumber != "" ||
		addr.Province != "" ||
		addr.City != "" ||
		addr.District != "" ||
		addr.DetailedAddress != "" ||
		addr.PostalCode != "" ||
		addr.Lat != 0 ||
		addr.Lon != 0
}

func fillDerivedStatuses(order *domain.Order) {
	if order == nil {
		return
	}
	if order.PaymentStatus == pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		switch order.Status {
		case pb.OrderStatus_REFUND_REQUESTED:
			order.PaymentStatus = pb.PaymentStatus_REFUNDING
		case pb.OrderStatus_REFUNDED:
			order.PaymentStatus = pb.PaymentStatus_REFUND_SUCCESS
		case pb.OrderStatus_CANCELLED:
			if order.PaidAt != nil {
				order.PaymentStatus = pb.PaymentStatus_REFUNDING
			} else {
				order.PaymentStatus = pb.PaymentStatus_UNPAID
			}
		case pb.OrderStatus_PAID, pb.OrderStatus_SHIPPED, pb.OrderStatus_DELIVERED, pb.OrderStatus_COMPLETED:
			order.PaymentStatus = pb.PaymentStatus_SUCCESS
		case pb.OrderStatus_PENDING_PAYMENT, pb.OrderStatus_ALLOCATING:
			order.PaymentStatus = pb.PaymentStatus_UNPAID
		case pb.OrderStatus_CLOSED:
			order.PaymentStatus = pb.PaymentStatus_FAILED
		}
	}

	if order.ShippingStatus == pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED {
		switch order.Status {
		case pb.OrderStatus_SHIPPED:
			order.ShippingStatus = pb.ShippingStatus_SHIPPING_SHIPPED
		case pb.OrderStatus_DELIVERED, pb.OrderStatus_COMPLETED:
			order.ShippingStatus = pb.ShippingStatus_SHIPPING_DELIVERED
		case pb.OrderStatus_CANCELLED, pb.OrderStatus_REFUND_REQUESTED, pb.OrderStatus_REFUNDED:
			order.ShippingStatus = pb.ShippingStatus_EXCEPTION
		case pb.OrderStatus_PENDING_PAYMENT, pb.OrderStatus_ALLOCATING, pb.OrderStatus_PAID:
			order.ShippingStatus = pb.ShippingStatus_PENDING_SHIPMENT
		}
	}
}
