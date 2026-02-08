// Package domain 履约服务领域层
// 生成摘要：
// 1) 定义履约聚合根和状态机
// 2) 定义拣货任务、打包任务实体
// 假设：
// - 履约单按状态机严格流转
// - 支持部分拣货和异常处理
package domain

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// FulfillmentStatus 履约单状态
type FulfillmentStatus int8

const (
	FulfillmentStatusPending     FulfillmentStatus = 1  // 待处理
	FulfillmentStatusPicking     FulfillmentStatus = 2  // 拣货中
	FulfillmentStatusPicked      FulfillmentStatus = 3  // 已拣货
	FulfillmentStatusPacking     FulfillmentStatus = 4  // 打包中
	FulfillmentStatusPacked      FulfillmentStatus = 5  // 已打包
	FulfillmentStatusReadyToShip FulfillmentStatus = 6  // 待发货
	FulfillmentStatusShipped     FulfillmentStatus = 7  // 已发货
	FulfillmentStatusCompleted   FulfillmentStatus = 8  // 已完成
	FulfillmentStatusCancelled   FulfillmentStatus = 9  // 已取消
	FulfillmentStatusException   FulfillmentStatus = 10 // 异常
)

// String 返回状态字符串
func (s FulfillmentStatus) String() string {
	switch s {
	case FulfillmentStatusPending:
		return "pending"
	case FulfillmentStatusPicking:
		return "picking"
	case FulfillmentStatusPicked:
		return "picked"
	case FulfillmentStatusPacking:
		return "packing"
	case FulfillmentStatusPacked:
		return "packed"
	case FulfillmentStatusReadyToShip:
		return "ready_to_ship"
	case FulfillmentStatusShipped:
		return "shipped"
	case FulfillmentStatusCompleted:
		return "completed"
	case FulfillmentStatusCancelled:
		return "cancelled"
	case FulfillmentStatusException:
		return "exception"
	default:
		return "unknown"
	}
}

// FulfillmentType 履约类型
type FulfillmentType int8

const (
	FulfillmentTypeNormal    FulfillmentType = 1 // 普通履约
	FulfillmentTypeExpress   FulfillmentType = 2 // 加急履约
	FulfillmentTypeScheduled FulfillmentType = 3 // 预约履约
	FulfillmentTypeSameDay   FulfillmentType = 4 // 当日达
	FulfillmentTypeNextDay   FulfillmentType = 5 // 次日达
)

// Fulfillment 履约单聚合根
type Fulfillment struct {
	gorm.Model
	// FulfillmentNo 履约单号，唯一标识
	FulfillmentNo string `gorm:"column:fulfillment_no;type:varchar(32);unique_index;not null" json:"fulfillment_no"`
	// OrderNo 关联订单号
	OrderNo string `gorm:"column:order_no;type:varchar(32);index;not null" json:"order_no"`
	// MerchantID 商家ID
	MerchantID uint64 `gorm:"column:merchant_id;index;not null" json:"merchant_id"`
	// StoreID 店铺ID
	StoreID uint64 `gorm:"column:store_id;index" json:"store_id"`
	// WarehouseID 仓库ID
	WarehouseID uint64 `gorm:"column:warehouse_id;index" json:"warehouse_id"`
	// Type 履约类型
	Type FulfillmentType `gorm:"column:type;type:tinyint;not null;default:1" json:"type"`
	// Status 履约状态
	Status FulfillmentStatus `gorm:"column:status;type:tinyint;not null;default:1" json:"status"`

	// 收货地址
	ReceiverName  string `gorm:"column:receiver_name;type:varchar(64)" json:"receiver_name"`
	ReceiverPhone string `gorm:"column:receiver_phone;type:varchar(20)" json:"receiver_phone"`
	Province      string `gorm:"column:province;type:varchar(32)" json:"province"`
	City          string `gorm:"column:city;type:varchar(32)" json:"city"`
	District      string `gorm:"column:district;type:varchar(32)" json:"district"`
	Address       string `gorm:"column:address;type:varchar(255)" json:"address"`
	PostalCode    string `gorm:"column:postal_code;type:varchar(10)" json:"postal_code"`

	// 拣货信息
	PickerID       uint64     `gorm:"column:picker_id" json:"picker_id"`
	PickerName     string     `gorm:"column:picker_name;type:varchar(64)" json:"picker_name"`
	PickAssignAt   *time.Time `gorm:"column:pick_assign_at" json:"pick_assign_at"`
	PickStartAt    *time.Time `gorm:"column:pick_start_at" json:"pick_start_at"`
	PickCompleteAt *time.Time `gorm:"column:pick_complete_at" json:"pick_complete_at"`

	// 打包信息
	PackerID       uint64     `gorm:"column:packer_id" json:"packer_id"`
	PackerName     string     `gorm:"column:packer_name;type:varchar(64)" json:"packer_name"`
	PackStartAt    *time.Time `gorm:"column:pack_start_at" json:"pack_start_at"`
	PackCompleteAt *time.Time `gorm:"column:pack_complete_at" json:"pack_complete_at"`

	// 物流信息
	CarrierCode string     `gorm:"column:carrier_code;type:varchar(32)" json:"carrier_code"`
	CarrierName string     `gorm:"column:carrier_name;type:varchar(64)" json:"carrier_name"`
	TrackingNo  string     `gorm:"column:tracking_no;type:varchar(64);index" json:"tracking_no"`
	ShippingFee int64      `gorm:"column:shipping_fee;not null;default:0" json:"shipping_fee"`
	ShippedAt   *time.Time `gorm:"column:shipped_at" json:"shipped_at"`

	// 时间信息
	ExpectedShipTime *time.Time `gorm:"column:expected_ship_time" json:"expected_ship_time"`

	// 备注
	Remark       string `gorm:"column:remark;type:text" json:"remark"`
	CancelReason string `gorm:"column:cancel_reason;type:varchar(255)" json:"cancel_reason"`
	CancelBy     string `gorm:"column:cancel_by;type:varchar(64)" json:"cancel_by"`

	// 关联数据
	Items      []FulfillmentItem  `gorm:"foreignKey:FulfillmentID" json:"items"`
	Packages   []Package          `gorm:"foreignKey:FulfillmentID" json:"packages"`
	Exceptions []PickingException `gorm:"foreignKey:FulfillmentID" json:"exceptions"`

	// 领域事件（不持久化）
	domainEvents []DomainEvent `gorm:"-" json:"-"`
}

// TableName 表名
func (Fulfillment) TableName() string {
	return "fulfillments"
}

// FulfillmentItem 履约商品项
type FulfillmentItem struct {
	gorm.Model
	// FulfillmentID 履约单ID
	FulfillmentID uint `gorm:"column:fulfillment_id;index;not null" json:"fulfillment_id"`
	// SKUID SKU标识
	SKUID string `gorm:"column:sku_id;type:varchar(64);not null" json:"sku_id"`
	// ProductName 商品名称
	ProductName string `gorm:"column:product_name;type:varchar(255)" json:"product_name"`
	// SKUName SKU名称
	SKUName string `gorm:"column:sku_name;type:varchar(255)" json:"sku_name"`
	// ImageURL 图片URL
	ImageURL string `gorm:"column:image_url;type:varchar(512)" json:"image_url"`
	// Quantity 应发数量
	Quantity int32 `gorm:"column:quantity;not null;default:0" json:"quantity"`
	// PickedQuantity 已拣数量
	PickedQuantity int32 `gorm:"column:picked_quantity;not null;default:0" json:"picked_quantity"`
	// PackedQuantity 已打包数量
	PackedQuantity int32 `gorm:"column:packed_quantity;not null;default:0" json:"packed_quantity"`
	// Location 库位
	Location string `gorm:"column:location;type:varchar(64)" json:"location"`
	// BatchNo 批次号
	BatchNo string `gorm:"column:batch_no;type:varchar(64)" json:"batch_no"`
}

// TableName 表名
func (FulfillmentItem) TableName() string {
	return "fulfillment_items"
}

// Package 包裹
type Package struct {
	gorm.Model
	// FulfillmentID 履约单ID
	FulfillmentID uint `gorm:"column:fulfillment_id;index;not null" json:"fulfillment_id"`
	// PackageNo 包裹号
	PackageNo string `gorm:"column:package_no;type:varchar(32);unique_index;not null" json:"package_no"`
	// Weight 重量(kg)
	Weight float64 `gorm:"column:weight;type:decimal(10,3);not null;default:0" json:"weight"`
	// Length 长(cm)
	Length float64 `gorm:"column:length;type:decimal(10,2);not null;default:0" json:"length"`
	// Width 宽(cm)
	Width float64 `gorm:"column:width;type:decimal(10,2);not null;default:0" json:"width"`
	// Height 高(cm)
	Height float64 `gorm:"column:height;type:decimal(10,2);not null;default:0" json:"height"`
	// SKUIDs 包含的SKU（JSON数组）
	SKUIDs string `gorm:"column:sku_ids;type:text" json:"sku_ids"`
}

// TableName 表名
func (Package) TableName() string {
	return "fulfillment_packages"
}

// PickingException 拣货异常
type PickingException struct {
	gorm.Model
	// FulfillmentID 履约单ID
	FulfillmentID uint `gorm:"column:fulfillment_id;index;not null" json:"fulfillment_id"`
	// SKUID SKU标识
	SKUID string `gorm:"column:sku_id;type:varchar(64);not null" json:"sku_id"`
	// ExpectedQuantity 期望数量
	ExpectedQuantity int32 `gorm:"column:expected_quantity;not null" json:"expected_quantity"`
	// ActualQuantity 实际数量
	ActualQuantity int32 `gorm:"column:actual_quantity;not null" json:"actual_quantity"`
	// Reason 异常原因
	Reason string `gorm:"column:reason;type:varchar(255)" json:"reason"`
	// ReportedAt 报告时间
	ReportedAt time.Time `gorm:"column:reported_at;not null" json:"reported_at"`
}

// TableName 表名
func (PickingException) TableName() string {
	return "picking_exceptions"
}

// NewFulfillment 创建履约单
func NewFulfillment(orderNo string, merchantID, storeID, warehouseID uint64, fulfillType FulfillmentType) *Fulfillment {
	now := time.Now()
	fulfillmentNo := fmt.Sprintf("FO%s%04d", now.Format("20060102150405"), now.UnixNano()%10000)

	f := &Fulfillment{
		FulfillmentNo: fulfillmentNo,
		OrderNo:       orderNo,
		MerchantID:    merchantID,
		StoreID:       storeID,
		WarehouseID:   warehouseID,
		Type:          fulfillType,
		Status:        FulfillmentStatusPending,
		domainEvents:  make([]DomainEvent, 0),
	}

	f.addEvent(&FulfillmentCreatedEvent{
		FulfillmentID: uint64(f.ID),
		FulfillmentNo: f.FulfillmentNo,
		OrderNo:       orderNo,
		MerchantID:    merchantID,
		Timestamp:     now,
	})

	return f
}

// SetShippingAddress 设置收货地址
func (f *Fulfillment) SetShippingAddress(name, phone, province, city, district, address, postalCode string) {
	f.ReceiverName = name
	f.ReceiverPhone = phone
	f.Province = province
	f.City = city
	f.District = district
	f.Address = address
	f.PostalCode = postalCode
}

// AddItem 添加履约商品
func (f *Fulfillment) AddItem(skuID, productName, skuName, imageURL, location, batchNo string, quantity int32) {
	f.Items = append(f.Items, FulfillmentItem{
		SKUID:       skuID,
		ProductName: productName,
		SKUName:     skuName,
		ImageURL:    imageURL,
		Quantity:    quantity,
		Location:    location,
		BatchNo:     batchNo,
	})
}

// AssignPicker 分配拣货员
func (f *Fulfillment) AssignPicker(pickerID uint64, pickerName string) error {
	if f.Status != FulfillmentStatusPending {
		return errors.New("can only assign picker when status is pending")
	}

	now := time.Now()
	f.PickerID = pickerID
	f.PickerName = pickerName
	f.PickAssignAt = &now

	f.addEvent(&PickerAssignedEvent{
		FulfillmentID: uint64(f.ID),
		PickerID:      pickerID,
		PickerName:    pickerName,
		Timestamp:     now,
	})

	return nil
}

// StartPicking 开始拣货
func (f *Fulfillment) StartPicking(pickerID uint64) error {
	if f.Status != FulfillmentStatusPending {
		return errors.New("can only start picking when status is pending")
	}
	if f.PickerID != pickerID {
		return errors.New("picker mismatch")
	}

	now := time.Now()
	f.Status = FulfillmentStatusPicking
	f.PickStartAt = &now

	f.addEvent(&PickingStartedEvent{
		FulfillmentID: uint64(f.ID),
		PickerID:      pickerID,
		Timestamp:     now,
	})

	return nil
}

// CompletePicking 完成拣货
func (f *Fulfillment) CompletePicking(pickedItems map[string]int32) error {
	if f.Status != FulfillmentStatusPicking {
		return errors.New("can only complete picking when status is picking")
	}

	now := time.Now()

	// 更新已拣数量
	for i := range f.Items {
		if qty, ok := pickedItems[f.Items[i].SKUID]; ok {
			f.Items[i].PickedQuantity = qty
		}
	}

	f.Status = FulfillmentStatusPicked
	f.PickCompleteAt = &now

	f.addEvent(&PickingCompletedEvent{
		FulfillmentID: uint64(f.ID),
		PickerID:      f.PickerID,
		Timestamp:     now,
	})

	return nil
}

// ReportException 报告拣货异常
func (f *Fulfillment) ReportException(skuID string, expected, actual int32, reason string) error {
	if f.Status != FulfillmentStatusPicking && f.Status != FulfillmentStatusPicked {
		return errors.New("can only report exception during picking")
	}

	now := time.Now()
	f.Exceptions = append(f.Exceptions, PickingException{
		SKUID:            skuID,
		ExpectedQuantity: expected,
		ActualQuantity:   actual,
		Reason:           reason,
		ReportedAt:       now,
	})

	return nil
}

// StartPacking 开始打包
func (f *Fulfillment) StartPacking(packerID uint64, packerName string) error {
	if f.Status != FulfillmentStatusPicked {
		return errors.New("can only start packing when status is picked")
	}

	now := time.Now()
	f.Status = FulfillmentStatusPacking
	f.PackerID = packerID
	f.PackerName = packerName
	f.PackStartAt = &now

	f.addEvent(&PackingStartedEvent{
		FulfillmentID: uint64(f.ID),
		PackerID:      packerID,
		Timestamp:     now,
	})

	return nil
}

// CompletePacking 完成打包
func (f *Fulfillment) CompletePacking(packages []Package) error {
	if f.Status != FulfillmentStatusPacking {
		return errors.New("can only complete packing when status is packing")
	}

	now := time.Now()
	f.Packages = packages
	f.Status = FulfillmentStatusPacked
	f.PackCompleteAt = &now

	// 更新已打包数量
	for i := range f.Items {
		f.Items[i].PackedQuantity = f.Items[i].PickedQuantity
	}

	f.addEvent(&PackingCompletedEvent{
		FulfillmentID: uint64(f.ID),
		PackerID:      f.PackerID,
		PackageCount:  len(packages),
		Timestamp:     now,
	})

	return nil
}

// ArrangeShipment 安排发货
func (f *Fulfillment) ArrangeShipment(carrierCode, carrierName string, shippingFee int64) error {
	if f.Status != FulfillmentStatusPacked {
		return errors.New("can only arrange shipment when status is packed")
	}

	f.CarrierCode = carrierCode
	f.CarrierName = carrierName
	f.ShippingFee = shippingFee
	f.Status = FulfillmentStatusReadyToShip

	return nil
}

// ConfirmShipment 确认发货
func (f *Fulfillment) ConfirmShipment(trackingNo string) error {
	if f.Status != FulfillmentStatusReadyToShip {
		return errors.New("can only confirm shipment when status is ready to ship")
	}

	now := time.Now()
	f.TrackingNo = trackingNo
	f.Status = FulfillmentStatusShipped
	f.ShippedAt = &now

	f.addEvent(&ShipmentConfirmedEvent{
		FulfillmentID: uint64(f.ID),
		TrackingNo:    trackingNo,
		CarrierCode:   f.CarrierCode,
		Timestamp:     now,
	})

	return nil
}

// Cancel 取消履约单
func (f *Fulfillment) Cancel(reason, operator string) error {
	if f.Status == FulfillmentStatusShipped || f.Status == FulfillmentStatusCompleted {
		return errors.New("cannot cancel shipped or completed fulfillment")
	}
	if f.Status == FulfillmentStatusCancelled {
		return errors.New("fulfillment already cancelled")
	}

	now := time.Now()
	f.Status = FulfillmentStatusCancelled
	f.CancelReason = reason
	f.CancelBy = operator

	f.addEvent(&FulfillmentCancelledEvent{
		FulfillmentID: uint64(f.ID),
		Reason:        reason,
		Operator:      operator,
		Timestamp:     now,
	})

	return nil
}

// addEvent 添加领域事件
func (f *Fulfillment) addEvent(event DomainEvent) {
	f.domainEvents = append(f.domainEvents, event)
}

// GetDomainEvents 获取领域事件
func (f *Fulfillment) GetDomainEvents() []DomainEvent {
	return f.domainEvents
}

// ClearDomainEvents 清除领域事件
func (f *Fulfillment) ClearDomainEvents() {
	f.domainEvents = nil
}
