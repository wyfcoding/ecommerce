package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrLogisticsNotFound = errors.New("物流信息不存在")
	ErrInvalidStatus     = errors.New("无效的状态")
)

// LogisticsStatus 物流状态
type LogisticsStatus int8

const (
	LogisticsStatusPending    LogisticsStatus = 0 // 待发货
	LogisticsStatusPickedUp   LogisticsStatus = 1 // 已揽收
	LogisticsStatusInTransit  LogisticsStatus = 2 // 运输中
	LogisticsStatusDelivering LogisticsStatus = 3 // 派送中
	LogisticsStatusDelivered  LogisticsStatus = 4 // 已签收
	LogisticsStatusReturning  LogisticsStatus = 5 // 退回中
	LogisticsStatusReturned   LogisticsStatus = 6 // 已退回
	LogisticsStatusException  LogisticsStatus = 7 // 异常
)

// Logistics 物流实体
type Logistics struct {
	gorm.Model
	OrderID         uint64            `gorm:"not null;index;comment:订单ID" json:"order_id"`
	OrderNo         string            `gorm:"type:varchar(64);not null;comment:订单号" json:"order_no"`
	TrackingNo      string            `gorm:"type:varchar(64);uniqueIndex;comment:物流单号" json:"tracking_no"`
	Carrier         string            `gorm:"type:varchar(64);comment:承运商" json:"carrier"`
	CarrierCode     string            `gorm:"type:varchar(32);comment:承运商编码" json:"carrier_code"`
	SenderName      string            `gorm:"type:varchar(64);comment:发件人姓名" json:"sender_name"`
	SenderPhone     string            `gorm:"type:varchar(20);comment:发件人电话" json:"sender_phone"`
	SenderAddress   string            `gorm:"type:varchar(255);comment:发件人地址" json:"sender_address"`
	ReceiverName    string            `gorm:"type:varchar(64);comment:收件人姓名" json:"receiver_name"`
	ReceiverPhone   string            `gorm:"type:varchar(20);comment:收件人电话" json:"receiver_phone"`
	ReceiverAddress string            `gorm:"type:varchar(255);comment:收件人地址" json:"receiver_address"`
	Status          LogisticsStatus   `gorm:"default:0;comment:状态" json:"status"`
	CurrentLocation string            `gorm:"type:varchar(255);comment:当前位置" json:"current_location"`
	EstimatedTime   *time.Time        `gorm:"comment:预计送达时间" json:"estimated_time"`
	DeliveredAt     *time.Time        `gorm:"comment:签收时间" json:"delivered_at"`
	Traces          []*LogisticsTrace `gorm:"foreignKey:LogisticsID" json:"traces"`
}

// LogisticsTrace 物流轨迹实体
type LogisticsTrace struct {
	gorm.Model
	LogisticsID uint64 `gorm:"not null;index;comment:物流ID" json:"logistics_id"`
	TrackingNo  string `gorm:"type:varchar(64);not null;comment:物流单号" json:"tracking_no"`
	Location    string `gorm:"type:varchar(255);comment:位置" json:"location"`
	Description string `gorm:"type:text;comment:描述" json:"description"`
	Status      string `gorm:"type:varchar(32);comment:状态描述" json:"status"`
}

// NewLogistics 创建物流
func NewLogistics(orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string,
	receiverName, receiverPhone, receiverAddress string) *Logistics {
	return &Logistics{
		OrderID:         orderID,
		OrderNo:         orderNo,
		TrackingNo:      trackingNo,
		Carrier:         carrier,
		CarrierCode:     carrierCode,
		SenderName:      senderName,
		SenderPhone:     senderPhone,
		SenderAddress:   senderAddress,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ReceiverAddress: receiverAddress,
		Status:          LogisticsStatusPending,
	}
}

// PickUp 揽收
func (l *Logistics) PickUp() {
	l.Status = LogisticsStatusPickedUp
}

// Transit 运输中
func (l *Logistics) Transit(location string) {
	l.Status = LogisticsStatusInTransit
	l.CurrentLocation = location
}

// Deliver 派送中
func (l *Logistics) Deliver() {
	l.Status = LogisticsStatusDelivering
}

// Complete 签收
func (l *Logistics) Complete() {
	l.Status = LogisticsStatusDelivered
	now := time.Now()
	l.DeliveredAt = &now
}

// Return 退回
func (l *Logistics) Return() {
	l.Status = LogisticsStatusReturning
}

// ReturnComplete 退回完成
func (l *Logistics) ReturnComplete() {
	l.Status = LogisticsStatusReturned
}

// Exception 异常
func (l *Logistics) Exception(reason string) {
	l.Status = LogisticsStatusException
	l.CurrentLocation = reason
}

// UpdateLocation 更新位置
func (l *Logistics) UpdateLocation(location string) {
	l.CurrentLocation = location
}

// SetEstimatedTime 设置预计送达时间
func (l *Logistics) SetEstimatedTime(estimatedTime time.Time) {
	l.EstimatedTime = &estimatedTime
}

// AddTrace 添加轨迹
func (l *Logistics) AddTrace(location, description, status string) {
	trace := &LogisticsTrace{
		TrackingNo:  l.TrackingNo,
		Location:    location,
		Description: description,
		Status:      status,
	}
	l.Traces = append(l.Traces, trace)
}
