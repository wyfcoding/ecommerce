// 生成摘要：
// - 从 delivery 服务合并到 logistics 域。
// - 最后一公里配送（即时配送、骑手调度）属于物流域子聚合。
// - 关键实体：LastMileDeliveryOrder（配送订单聚合根）、DeliveryDriver（骑手实体）。
// - 并发控制策略：乐观锁 (Version) + 状态机校验。
package domain

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/shopspring/decimal"
)

// 配送模块业务错误定义。
var (
	ErrDeliveryNotFound        = errors.New("delivery order not found")
	ErrDeliveryAlreadyAssigned = errors.New("delivery already assigned to a driver")
	ErrDriverNotAvailable      = errors.New("no available driver")
	ErrInvalidDeliveryTransit  = errors.New("invalid delivery status transition")
	ErrDeliveryTimeout         = errors.New("delivery has timed out")
	ErrDeliveryAreaNotCovered  = errors.New("delivery area not covered")
)

// LastMileDeliveryStatus 配送状态。
type LastMileDeliveryStatus string

const (
	LMDeliveryPending   LastMileDeliveryStatus = "PENDING"
	LMDeliveryAssigned  LastMileDeliveryStatus = "ASSIGNED"
	LMDeliveryPickedUp  LastMileDeliveryStatus = "PICKED_UP"
	LMDeliveryInTransit LastMileDeliveryStatus = "IN_TRANSIT"
	LMDeliveryArrived   LastMileDeliveryStatus = "ARRIVED"
	LMDeliveryDelivered LastMileDeliveryStatus = "DELIVERED"
	LMDeliveryFailed    LastMileDeliveryStatus = "FAILED"
	LMDeliveryCancelled LastMileDeliveryStatus = "CANCELLED"
	LMDeliveryReturning LastMileDeliveryStatus = "RETURNING"
)

// LastMileDeliveryType 配送类型。
type LastMileDeliveryType string

const (
	LMTypeInstant    LastMileDeliveryType = "INSTANT"
	LMTypeSameDay    LastMileDeliveryType = "SAME_DAY"
	LMTypeNextDay    LastMileDeliveryType = "NEXT_DAY"
	LMTypeScheduled  LastMileDeliveryType = "SCHEDULED"
	LMTypeSelfPickup LastMileDeliveryType = "SELF_PICKUP"
)

// DeliveryAddress 配送地址值对象。
type DeliveryAddress struct {
	Province  string  `json:"province"`
	City      string  `json:"city"`
	District  string  `json:"district"`
	Street    string  `json:"street"`
	Detail    string  `json:"detail"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// DeliveryLocationLog 位置轨迹记录。
type DeliveryLocationLog struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

// LastMileDeliveryOrder 最后一公里配送订单聚合根。
type LastMileDeliveryOrder struct {
	ID            uint64                 `json:"id"`
	DeliveryNo    string                 `json:"delivery_no"`
	OrderID       uint64                 `json:"order_id"`
	OrderNo       string                 `json:"order_no"`
	MerchantID    uint64                 `json:"merchant_id"`
	UserID        uint64                 `json:"user_id"`
	DriverID      uint64                 `json:"driver_id"`
	Type          LastMileDeliveryType   `json:"type"`
	Status        LastMileDeliveryStatus `json:"status"`
	PickupAddress *DeliveryAddress       `json:"pickup_address"`
	DeliveryAddr  *DeliveryAddress       `json:"delivery_address"`
	ReceiverName  string                 `json:"receiver_name"`
	ReceiverPhone string                 `json:"receiver_phone"`
	DeliveryFee   decimal.Decimal        `json:"delivery_fee"`
	Tips          decimal.Decimal        `json:"tips"`
	Distance      float64                `json:"distance"`
	EstimatedTime int32                  `json:"estimated_time"`
	ActualTime    int32                  `json:"actual_time"`
	ScheduledAt   *time.Time             `json:"scheduled_at"`
	AssignedAt    *time.Time             `json:"assigned_at"`
	PickedUpAt    *time.Time             `json:"picked_up_at"`
	DeliveredAt   *time.Time             `json:"delivered_at"`
	Remark        string                 `json:"remark"`
	ProofImages   []string               `json:"proof_images"`
	FailReason    string                 `json:"fail_reason"`
	Locations     []*DeliveryLocationLog `json:"locations"`
	Version       int64                  `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// DeliveryDriverStatus 骑手状态。
type DeliveryDriverStatus string

const (
	DriverOffline    DeliveryDriverStatus = "OFFLINE"
	DriverOnline     DeliveryDriverStatus = "ONLINE"
	DriverBusy       DeliveryDriverStatus = "BUSY"
	DriverDelivering DeliveryDriverStatus = "DELIVERING"
)

// DeliveryDriver 骑手实体。
type DeliveryDriver struct {
	ID              uint64               `json:"id"`
	Name            string               `json:"name"`
	Phone           string               `json:"phone"`
	Status          DeliveryDriverStatus `json:"status"`
	Latitude        float64              `json:"latitude"`
	Longitude       float64              `json:"longitude"`
	CurrentOrders   int32                `json:"current_orders"`
	MaxOrders       int32                `json:"max_orders"`
	Rating          decimal.Decimal      `json:"rating"`
	TotalDeliveries int64                `json:"total_deliveries"`
	OnlineAt        *time.Time           `json:"online_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// IsAvailable 判断骑手是否可接单。
func (d *DeliveryDriver) IsAvailable() bool {
	return d.Status == DriverOnline && d.CurrentOrders < d.MaxOrders
}

// DistanceTo 计算骑手到指定坐标的距离（Haversine 公式，单位：公里）。
func (d *DeliveryDriver) DistanceTo(lat, lon float64) float64 {
	const earthRadius = 6371.0
	dLat := (lat - d.Latitude) * math.Pi / 180
	dLon := (lon - d.Longitude) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(d.Latitude*math.Pi/180)*math.Cos(lat*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// NewLastMileDeliveryOrder 创建配送订单。
func NewLastMileDeliveryOrder(
	deliveryNo string, orderID uint64, orderNo string,
	merchantID, userID uint64, deliveryType LastMileDeliveryType,
	pickup, delivery *DeliveryAddress, receiverName, receiverPhone string,
) *LastMileDeliveryOrder {
	return &LastMileDeliveryOrder{
		DeliveryNo:    deliveryNo,
		OrderID:       orderID,
		OrderNo:       orderNo,
		MerchantID:    merchantID,
		UserID:        userID,
		Type:          deliveryType,
		Status:        LMDeliveryPending,
		PickupAddress: pickup,
		DeliveryAddr:  delivery,
		ReceiverName:  receiverName,
		ReceiverPhone: receiverPhone,
		Locations:     make([]*DeliveryLocationLog, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// AssignDriver 分配骑手。
func (o *LastMileDeliveryOrder) AssignDriver(driverID uint64) error {
	if o.Status != LMDeliveryPending {
		return ErrInvalidDeliveryTransit
	}
	o.DriverID = driverID
	o.Status = LMDeliveryAssigned
	now := time.Now()
	o.AssignedAt = &now
	o.UpdatedAt = now
	return nil
}

// PickUp 骑手取货。
func (o *LastMileDeliveryOrder) PickUp() error {
	if o.Status != LMDeliveryAssigned {
		return ErrInvalidDeliveryTransit
	}
	o.Status = LMDeliveryPickedUp
	now := time.Now()
	o.PickedUpAt = &now
	o.UpdatedAt = now
	return nil
}

// StartDelivery 开始配送。
func (o *LastMileDeliveryOrder) StartDelivery() error {
	if o.Status != LMDeliveryPickedUp {
		return ErrInvalidDeliveryTransit
	}
	o.Status = LMDeliveryInTransit
	o.UpdatedAt = time.Now()
	return nil
}

// ConfirmDelivered 确认送达。
func (o *LastMileDeliveryOrder) ConfirmDelivered(proofImages []string) error {
	allowed := map[LastMileDeliveryStatus]bool{
		LMDeliveryInTransit: true,
		LMDeliveryArrived:   true,
	}
	if !allowed[o.Status] {
		return ErrInvalidDeliveryTransit
	}
	o.Status = LMDeliveryDelivered
	o.ProofImages = proofImages
	now := time.Now()
	o.DeliveredAt = &now
	o.UpdatedAt = now
	if o.PickedUpAt != nil {
		o.ActualTime = int32(now.Sub(*o.PickedUpAt).Minutes())
	}
	return nil
}

// Fail 标记配送失败。
func (o *LastMileDeliveryOrder) Fail(reason string) error {
	allowed := map[LastMileDeliveryStatus]bool{
		LMDeliveryAssigned:  true,
		LMDeliveryPickedUp:  true,
		LMDeliveryInTransit: true,
		LMDeliveryArrived:   true,
	}
	if !allowed[o.Status] {
		return ErrInvalidDeliveryTransit
	}
	o.Status = LMDeliveryFailed
	o.FailReason = reason
	o.UpdatedAt = time.Now()
	return nil
}

// CancelDelivery 取消配送。
func (o *LastMileDeliveryOrder) CancelDelivery() error {
	allowed := map[LastMileDeliveryStatus]bool{
		LMDeliveryPending:  true,
		LMDeliveryAssigned: true,
	}
	if !allowed[o.Status] {
		return ErrInvalidDeliveryTransit
	}
	o.Status = LMDeliveryCancelled
	o.UpdatedAt = time.Now()
	return nil
}

// AddLocation 添加位置轨迹。
func (o *LastMileDeliveryOrder) AddLocation(lat, lon, speed, heading float64) {
	o.Locations = append(o.Locations, &DeliveryLocationLog{
		Latitude:  lat,
		Longitude: lon,
		Speed:     speed,
		Heading:   heading,
		Timestamp: time.Now(),
	})
}

// IsTimeout 判断配送是否超时。
func (o *LastMileDeliveryOrder) IsTimeout() bool {
	if o.EstimatedTime <= 0 || o.PickedUpAt == nil {
		return false
	}
	deadline := o.PickedUpAt.Add(time.Duration(o.EstimatedTime) * time.Minute)
	return time.Now().After(deadline) && o.Status != LMDeliveryDelivered
}

// CalculateDeliveryDistance 计算取货点到配送点的直线距离。
func (o *LastMileDeliveryOrder) CalculateDeliveryDistance() float64 {
	if o.PickupAddress == nil || o.DeliveryAddr == nil {
		return 0
	}
	const earthRadius = 6371.0
	dLat := (o.DeliveryAddr.Latitude - o.PickupAddress.Latitude) * math.Pi / 180
	dLon := (o.DeliveryAddr.Longitude - o.PickupAddress.Longitude) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(o.PickupAddress.Latitude*math.Pi/180)*math.Cos(o.DeliveryAddr.Latitude*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	o.Distance = earthRadius * c
	return o.Distance
}

// LastMileDeliveryRepository 最后一公里配送仓储接口。
type LastMileDeliveryRepository interface {
	Save(ctx context.Context, order *LastMileDeliveryOrder) error
	GetByID(ctx context.Context, id uint64) (*LastMileDeliveryOrder, error)
	GetByDeliveryNo(ctx context.Context, deliveryNo string) (*LastMileDeliveryOrder, error)
	GetByOrderID(ctx context.Context, orderID uint64) (*LastMileDeliveryOrder, error)
	ListPending(ctx context.Context, limit int) ([]*LastMileDeliveryOrder, error)
	ListByDriver(ctx context.Context, driverID uint64, status LastMileDeliveryStatus) ([]*LastMileDeliveryOrder, error)
	UpdateStatus(ctx context.Context, id uint64, status LastMileDeliveryStatus, version int64) error
}

// DeliveryDriverRepository 骑手仓储接口。
type DeliveryDriverRepository interface {
	Save(ctx context.Context, driver *DeliveryDriver) error
	GetByID(ctx context.Context, id uint64) (*DeliveryDriver, error)
	ListAvailable(ctx context.Context, lat, lon, radiusKm float64) ([]*DeliveryDriver, error)
	UpdateLocation(ctx context.Context, driverID uint64, lat, lon float64) error
	UpdateStatus(ctx context.Context, driverID uint64, status DeliveryDriverStatus) error
}
