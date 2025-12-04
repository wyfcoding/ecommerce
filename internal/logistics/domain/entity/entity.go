package entity

import (
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Logistics模块的业务错误。
var (
	ErrLogisticsNotFound = errors.New("物流信息不存在") // 物流记录未找到。
	ErrInvalidStatus     = errors.New("无效的状态")   // 无效的物流状态转换。
)

// LogisticsStatus 定义了物流单的生命周期状态。
type LogisticsStatus int8

const (
	LogisticsStatusPending    LogisticsStatus = 0 // 待发货：订单已创建，等待商品出库发货。
	LogisticsStatusPickedUp   LogisticsStatus = 1 // 已揽收：商品已由快递公司揽收。
	LogisticsStatusInTransit  LogisticsStatus = 2 // 运输中：商品正在运输途中。
	LogisticsStatusDelivering LogisticsStatus = 3 // 派送中：商品正在派送给收件人。
	LogisticsStatusDelivered  LogisticsStatus = 4 // 已签收：商品已被收件人签收。
	LogisticsStatusReturning  LogisticsStatus = 5 // 退回中：商品正在退回发件人途中。
	LogisticsStatusReturned   LogisticsStatus = 6 // 已退回：商品已退回发件人。
	LogisticsStatusException  LogisticsStatus = 7 // 异常：物流过程中出现异常情况。
)

// Logistics 实体是物流模块的聚合根。
// 它代表一个物流单的完整信息，包含了订单、商品发收件人信息、物流状态、轨迹和预计时间。
type Logistics struct {
	gorm.Model                        // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	OrderID         uint64            `gorm:"not null;index;comment:订单ID" json:"order_id"`                  // 关联的订单ID，索引字段。
	OrderNo         string            `gorm:"type:varchar(64);not null;comment:订单号" json:"order_no"`        // 关联的订单号。
	TrackingNo      string            `gorm:"type:varchar(64);uniqueIndex;comment:物流单号" json:"tracking_no"` // 物流单号，唯一索引。
	Carrier         string            `gorm:"type:varchar(64);comment:承运商" json:"carrier"`                  // 承运物流公司的名称。
	CarrierCode     string            `gorm:"type:varchar(32);comment:承运商编码" json:"carrier_code"`           // 承运物流公司的编码。
	SenderName      string            `gorm:"type:varchar(64);comment:发件人姓名" json:"sender_name"`            // 发件人姓名。
	SenderPhone     string            `gorm:"type:varchar(20);comment:发件人电话" json:"sender_phone"`           // 发件人电话。
	SenderAddress   string            `gorm:"type:varchar(255);comment:发件人地址" json:"sender_address"`        // 发件人地址。
	ReceiverName    string            `gorm:"type:varchar(64);comment:收件人姓名" json:"receiver_name"`          // 收件人姓名。
	ReceiverPhone   string            `gorm:"type:varchar(20);comment:收件人电话" json:"receiver_phone"`         // 收件人电话。
	ReceiverAddress string            `gorm:"type:varchar(255);comment:收件人地址" json:"receiver_address"`      // 收件人地址。
	SenderLat       float64           `gorm:"type:decimal(10,6);comment:发件人纬度" json:"sender_lat"`           // 发件人地址的纬度。
	SenderLon       float64           `gorm:"type:decimal(10,6);comment:发件人经度" json:"sender_lon"`           // 发件人地址的经度。
	ReceiverLat     float64           `gorm:"type:decimal(10,6);comment:收件人纬度" json:"receiver_lat"`         // 收件人地址的纬度。
	ReceiverLon     float64           `gorm:"type:decimal(10,6);comment:收件人经度" json:"receiver_lon"`         // 收件人地址的经度。
	Status          LogisticsStatus   `gorm:"default:0;comment:状态" json:"status"`                           // 物流单状态，默认为待发货。
	CurrentLocation string            `gorm:"type:varchar(255);comment:当前位置" json:"current_location"`       // 物流单当前所在位置。
	EstimatedTime   *time.Time        `gorm:"comment:预计送达时间" json:"estimated_time"`                         // 预计送达时间。
	DeliveredAt     *time.Time        `gorm:"comment:签收时间" json:"delivered_at"`                             // 实际签收时间。
	Traces          []*LogisticsTrace `gorm:"foreignKey:LogisticsID" json:"traces"`                         // 关联的物流轨迹记录列表，一对多关系。
	Route           *DeliveryRoute    `gorm:"foreignKey:LogisticsID" json:"route"`                          // 关联的配送路线信息，一对一关系。
}

// LogisticsTrace 实体代表物流单的一条轨迹记录。
type LogisticsTrace struct {
	gorm.Model         // 嵌入gorm.Model。
	LogisticsID uint64 `gorm:"not null;index;comment:物流ID" json:"logistics_id"`           // 关联的物流单ID，索引字段。
	TrackingNo  string `gorm:"type:varchar(64);not null;comment:物流单号" json:"tracking_no"` // 关联的物流单号。
	Location    string `gorm:"type:varchar(255);comment:位置" json:"location"`              // 轨迹发生的位置。
	Description string `gorm:"type:text;comment:描述" json:"description"`                   // 轨迹描述，例如“您的包裹已出库”。
	Status      string `gorm:"type:varchar(32);comment:状态描述" json:"status"`               // 轨迹发生时的物流状态描述。
}

// DeliveryRoute 实体代表一个物流单的配送路线规划信息。
type DeliveryRoute struct {
	gorm.Model          // 嵌入gorm.Model。
	LogisticsID uint64  `gorm:"not null;uniqueIndex;comment:物流ID" json:"logistics_id"` // 关联的物流单ID，唯一索引。
	RouteData   string  `gorm:"type:text;comment:路线数据(JSON)" json:"route_data"`        // 配送路线数据，通常为JSON格式的地理坐标点序列。
	Distance    float64 `gorm:"type:decimal(10,2);comment:总距离(米)" json:"distance"`     // 配送路线的总距离（米）。
}

// NewLogistics 创建并返回一个新的 Logistics 实体实例。
// orderID, orderNo: 订单信息。
// trackingNo, carrier, carrierCode: 运单和承运商信息。
// senderName, senderPhone, senderAddress, senderLat, senderLon: 发件人信息。
// receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon: 收件人信息。
func NewLogistics(orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64) *Logistics {
	return &Logistics{
		OrderID:         orderID,
		OrderNo:         orderNo,
		TrackingNo:      trackingNo,
		Carrier:         carrier,
		CarrierCode:     carrierCode,
		SenderName:      senderName,
		SenderPhone:     senderPhone,
		SenderAddress:   senderAddress,
		SenderLat:       senderLat,
		SenderLon:       senderLon,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ReceiverAddress: receiverAddress,
		ReceiverLat:     receiverLat,
		ReceiverLon:     receiverLon,
		Status:          LogisticsStatusPending, // 新创建的物流单默认为待发货状态。
	}
}

// PickUp 更新物流状态为“已揽收”。
func (l *Logistics) PickUp() {
	l.Status = LogisticsStatusPickedUp
}

// Transit 更新物流状态为“运输中”，并记录当前位置。
func (l *Logistics) Transit(location string) {
	l.Status = LogisticsStatusInTransit
	l.CurrentLocation = location
}

// Deliver 更新物流状态为“派送中”。
func (l *Logistics) Deliver() {
	l.Status = LogisticsStatusDelivering
}

// Complete 更新物流状态为“已签收”，并记录签收时间。
func (l *Logistics) Complete() {
	l.Status = LogisticsStatusDelivered
	now := time.Now()
	l.DeliveredAt = &now
}

// Return 更新物流状态为“退回中”。
func (l *Logistics) Return() {
	l.Status = LogisticsStatusReturning
}

// ReturnComplete 更新物流状态为“已退回”。
func (l *Logistics) ReturnComplete() {
	l.Status = LogisticsStatusReturned
}

// Exception 更新物流状态为“异常”，并记录异常原因。
func (l *Logistics) Exception(reason string) {
	l.Status = LogisticsStatusException
	l.CurrentLocation = reason // 在此场景下，CurrentLocation存储异常原因。
}

// UpdateLocation 更新物流单的当前位置。
func (l *Logistics) UpdateLocation(location string) {
	l.CurrentLocation = location
}

// SetEstimatedTime 设置物流单的预计送达时间。
func (l *Logistics) SetEstimatedTime(estimatedTime time.Time) {
	l.EstimatedTime = &estimatedTime
}

// AddTrace 添加一条物流轨迹记录。
// location: 轨迹发生的位置。
// description: 轨迹描述。
// status: 轨迹发生时的状态描述。
func (l *Logistics) AddTrace(location, description, status string) {
	trace := &LogisticsTrace{
		TrackingNo:  l.TrackingNo,
		Location:    location,
		Description: description,
		Status:      status,
	}
	l.Traces = append(l.Traces, trace) // 将轨迹添加到关联的Traces切片。
}
