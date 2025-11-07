package gateway

import (
	"context"
	"time"
)

// ExpressCompany 快递公司
type ExpressCompany string

const (
	ExpressCompanySF     ExpressCompany = "SF"     // 顺丰速运
	ExpressCompanyZTO    ExpressCompany = "ZTO"    // 中通快递
	ExpressCompanyYTO    ExpressCompany = "YTO"    // 圆通速递
	ExpressCompanyYD     ExpressCompany = "YD"     // 韵达快递
	ExpressCompanySTO    ExpressCompany = "STO"    // 申通快递
	ExpressCompanyEMS    ExpressCompany = "EMS"    // EMS
	ExpressCompanyJD     ExpressCompany = "JD"     // 京东物流
	ExpressCompanyDBL    ExpressCompany = "DBL"    // 德邦物流
)

// TrackingStatus 物流状态
type TrackingStatus string

const (
	TrackingStatusCollected  TrackingStatus = "COLLECTED"   // 已揽收
	TrackingStatusInTransit  TrackingStatus = "IN_TRANSIT"  // 运输中
	TrackingStatusDelivering TrackingStatus = "DELIVERING"  // 派送中
	TrackingStatusDelivered  TrackingStatus = "DELIVERED"   // 已签收
	TrackingStatusReturning  TrackingStatus = "RETURNING"   // 退回中
	TrackingStatusReturned   TrackingStatus = "RETURNED"    // 已退回
	TrackingStatusException  TrackingStatus = "EXCEPTION"   // 异常
)

// TrackingInfo 物流跟踪信息
type TrackingInfo struct {
	Company      ExpressCompany   `json:"company"`       // 快递公司
	TrackingNo   string           `json:"trackingNo"`    // 运单号
	Status       TrackingStatus   `json:"status"`        // 当前状态
	Traces       []*TrackingTrace `json:"traces"`        // 物流轨迹
	EstimatedTime *time.Time      `json:"estimatedTime"` // 预计送达时间
	UpdatedAt    time.Time        `json:"updatedAt"`     // 更新时间
}

// TrackingTrace 物流轨迹
type TrackingTrace struct {
	Time        time.Time `json:"time"`        // 时间
	Status      string    `json:"status"`      // 状态
	Description string    `json:"description"` // 描述
	Location    string    `json:"location"`    // 地点
}

// CreateWaybillRequest 创建运单请求
type CreateWaybillRequest struct {
	Company ExpressCompany `json:"company"` // 快递公司
	
	// 寄件人信息
	SenderName    string `json:"senderName"`
	SenderPhone   string `json:"senderPhone"`
	SenderProvince string `json:"senderProvince"`
	SenderCity    string `json:"senderCity"`
	SenderDistrict string `json:"senderDistrict"`
	SenderAddress string `json:"senderAddress"`
	
	// 收件人信息
	ReceiverName    string `json:"receiverName"`
	ReceiverPhone   string `json:"receiverPhone"`
	ReceiverProvince string `json:"receiverProvince"`
	ReceiverCity    string `json:"receiverCity"`
	ReceiverDistrict string `json:"receiverDistrict"`
	ReceiverAddress string `json:"receiverAddress"`
	
	// 货物信息
	GoodsName   string  `json:"goodsName"`   // 货物名称
	GoodsWeight float64 `json:"goodsWeight"` // 重量(kg)
	GoodsValue  uint64  `json:"goodsValue"`  // 货物价值(分)
	
	// 其他信息
	Remark string `json:"remark"` // 备注
}

// CreateWaybillResponse 创建运单响应
type CreateWaybillResponse struct {
	TrackingNo   string    `json:"trackingNo"`   // 运单号
	WaybillNo    string    `json:"waybillNo"`    // 电子面单号
	PrintData    string    `json:"printData"`    // 打印数据
	EstimatedFee uint64    `json:"estimatedFee"` // 预估费用(分)
	CreatedAt    time.Time `json:"createdAt"`    // 创建时间
}

// ExpressGateway 快递网关接口
type ExpressGateway interface {
	// CreateWaybill 创建电子面单
	CreateWaybill(ctx context.Context, req *CreateWaybillRequest) (*CreateWaybillResponse, error)
	
	// QueryTracking 查询物流轨迹
	QueryTracking(ctx context.Context, company ExpressCompany, trackingNo string) (*TrackingInfo, error)
	
	// CancelWaybill 取消运单
	CancelWaybill(ctx context.Context, company ExpressCompany, trackingNo string) error
	
	// CalculateFee 计算运费
	CalculateFee(ctx context.Context, req *CalculateFeeRequest) (*CalculateFeeResponse, error)
}

// CalculateFeeRequest 计算运费请求
type CalculateFeeRequest struct {
	Company        ExpressCompany `json:"company"`
	SenderCity     string         `json:"senderCity"`
	ReceiverCity   string         `json:"receiverCity"`
	Weight         float64        `json:"weight"`         // 重量(kg)
	Volume         float64        `json:"volume"`         // 体积(cm³)
	DeclaredValue  uint64         `json:"declaredValue"`  // 声明价值(分)
}

// CalculateFeeResponse 计算运费响应
type CalculateFeeResponse struct {
	BaseFee      uint64 `json:"baseFee"`      // 基础运费(分)
	WeightFee    uint64 `json:"weightFee"`    // 重量费用(分)
	InsuranceFee uint64 `json:"insuranceFee"` // 保价费用(分)
	TotalFee     uint64 `json:"totalFee"`     // 总费用(分)
}
