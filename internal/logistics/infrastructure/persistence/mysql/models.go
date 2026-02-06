package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"gorm.io/gorm"
)

// LogisticsModel 物流写模型（持久化专用）。
type LogisticsModel struct {
	gorm.Model
	OrderID         uint64                 `gorm:"not null;index;comment:订单ID"`
	OrderNo         string                 `gorm:"type:varchar(64);not null;comment:订单号"`
	TrackingNo      string                 `gorm:"type:varchar(64);uniqueIndex;comment:物流单号"`
	Carrier         string                 `gorm:"type:varchar(64);comment:承运商"`
	CarrierCode     string                 `gorm:"type:varchar(32);comment:承运商编码"`
	SenderName      string                 `gorm:"type:varchar(64);comment:发件人姓名"`
	SenderPhone     string                 `gorm:"type:varchar(20);comment:发件人电话"`
	SenderAddress   string                 `gorm:"type:varchar(255);comment:发件人地址"`
	ReceiverName    string                 `gorm:"type:varchar(64);comment:收件人姓名"`
	ReceiverPhone   string                 `gorm:"type:varchar(20);comment:收件人电话"`
	ReceiverAddress string                 `gorm:"type:varchar(255);comment:收件人地址"`
	SenderLat       float64                `gorm:"type:decimal(10,6);comment:发件人纬度"`
	SenderLon       float64                `gorm:"type:decimal(10,6);comment:发件人经度"`
	ReceiverLat     float64                `gorm:"type:decimal(10,6);comment:收件人纬度"`
	ReceiverLon     float64                `gorm:"type:decimal(10,6);comment:收件人经度"`
	Status          domain.LogisticsStatus `gorm:"default:0;comment:状态"`
	CurrentLocation string                 `gorm:"type:varchar(255);comment:当前位置"`
	EstimatedTime   *time.Time             `gorm:"comment:预计送达时间"`
	DeliveredAt     *time.Time             `gorm:"comment:签收时间"`
	RiderID         string                 `gorm:"type:varchar(64);comment:骑手ID"`
	Traces          []*LogisticsTraceModel `gorm:"foreignKey:LogisticsID"`
	Route           *DeliveryRouteModel    `gorm:"foreignKey:LogisticsID"`
}

func (LogisticsModel) TableName() string {
	return "logistics"
}

// LogisticsTraceModel 物流轨迹写模型（持久化专用）。
type LogisticsTraceModel struct {
	gorm.Model
	LogisticsID uint64 `gorm:"not null;index;comment:物流ID"`
	TrackingNo  string `gorm:"type:varchar(64);not null;comment:物流单号"`
	Location    string `gorm:"type:varchar(255);comment:位置"`
	Description string `gorm:"type:text;comment:描述"`
	Status      string `gorm:"type:varchar(32);comment:状态描述"`
}

func (LogisticsTraceModel) TableName() string {
	return "logistics_traces"
}

// DeliveryRouteModel 配送路线写模型（持久化专用）。
type DeliveryRouteModel struct {
	gorm.Model
	LogisticsID uint64  `gorm:"not null;uniqueIndex;comment:物流ID"`
	RouteData   string  `gorm:"type:text;comment:路线数据(JSON)"`
	Distance    float64 `gorm:"type:decimal(10,2);comment:总距离(米)"`
}

func (DeliveryRouteModel) TableName() string {
	return "delivery_routes"
}

func toLogisticsModel(logistics *domain.Logistics) *LogisticsModel {
	if logistics == nil {
		return nil
	}
	return &LogisticsModel{
		Model: gorm.Model{
			ID:        logistics.ID,
			CreatedAt: logistics.CreatedAt,
			UpdatedAt: logistics.UpdatedAt,
		},
		OrderID:         logistics.OrderID,
		OrderNo:         logistics.OrderNo,
		TrackingNo:      logistics.TrackingNo,
		Carrier:         logistics.Carrier,
		CarrierCode:     logistics.CarrierCode,
		SenderName:      logistics.SenderName,
		SenderPhone:     logistics.SenderPhone,
		SenderAddress:   logistics.SenderAddress,
		ReceiverName:    logistics.ReceiverName,
		ReceiverPhone:   logistics.ReceiverPhone,
		ReceiverAddress: logistics.ReceiverAddress,
		SenderLat:       logistics.SenderLat,
		SenderLon:       logistics.SenderLon,
		ReceiverLat:     logistics.ReceiverLat,
		ReceiverLon:     logistics.ReceiverLon,
		Status:          logistics.Status,
		CurrentLocation: logistics.CurrentLocation,
		EstimatedTime:   logistics.EstimatedTime,
		DeliveredAt:     logistics.DeliveredAt,
		RiderID:         logistics.RiderID,
	}
}

func toDomainLogistics(model *LogisticsModel) *domain.Logistics {
	if model == nil {
		return nil
	}

	logistics := &domain.Logistics{
		ID:              model.ID,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		OrderID:         model.OrderID,
		OrderNo:         model.OrderNo,
		TrackingNo:      model.TrackingNo,
		Carrier:         model.Carrier,
		CarrierCode:     model.CarrierCode,
		SenderName:      model.SenderName,
		SenderPhone:     model.SenderPhone,
		SenderAddress:   model.SenderAddress,
		ReceiverName:    model.ReceiverName,
		ReceiverPhone:   model.ReceiverPhone,
		ReceiverAddress: model.ReceiverAddress,
		SenderLat:       model.SenderLat,
		SenderLon:       model.SenderLon,
		ReceiverLat:     model.ReceiverLat,
		ReceiverLon:     model.ReceiverLon,
		Status:          model.Status,
		CurrentLocation: model.CurrentLocation,
		EstimatedTime:   model.EstimatedTime,
		DeliveredAt:     model.DeliveredAt,
		RiderID:         model.RiderID,
	}

	if len(model.Traces) > 0 {
		traces := make([]*domain.LogisticsTrace, 0, len(model.Traces))
		for _, t := range model.Traces {
			traces = append(traces, toDomainLogisticsTrace(t))
		}
		logistics.Traces = traces
	}
	if model.Route != nil {
		logistics.Route = toDomainDeliveryRoute(model.Route)
	}

	return logistics
}

func toLogisticsTraceModel(trace *domain.LogisticsTrace) *LogisticsTraceModel {
	if trace == nil {
		return nil
	}
	return &LogisticsTraceModel{
		Model: gorm.Model{
			ID:        trace.ID,
			CreatedAt: trace.CreatedAt,
			UpdatedAt: trace.UpdatedAt,
		},
		LogisticsID: trace.LogisticsID,
		TrackingNo:  trace.TrackingNo,
		Location:    trace.Location,
		Description: trace.Description,
		Status:      trace.Status,
	}
}

func toDomainLogisticsTrace(model *LogisticsTraceModel) *domain.LogisticsTrace {
	if model == nil {
		return nil
	}
	return &domain.LogisticsTrace{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		LogisticsID: model.LogisticsID,
		TrackingNo:  model.TrackingNo,
		Location:    model.Location,
		Description: model.Description,
		Status:      model.Status,
	}
}

func toDeliveryRouteModel(route *domain.DeliveryRoute) *DeliveryRouteModel {
	if route == nil {
		return nil
	}
	return &DeliveryRouteModel{
		Model: gorm.Model{
			ID:        route.ID,
			CreatedAt: route.CreatedAt,
			UpdatedAt: route.UpdatedAt,
		},
		LogisticsID: route.LogisticsID,
		RouteData:   route.RouteData,
		Distance:    route.Distance,
	}
}

func toDomainDeliveryRoute(model *DeliveryRouteModel) *domain.DeliveryRoute {
	if model == nil {
		return nil
	}
	return &domain.DeliveryRoute{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		LogisticsID: model.LogisticsID,
		RouteData:   model.RouteData,
		Distance:    model.Distance,
	}
}
