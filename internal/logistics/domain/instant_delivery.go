// 变更说明：新增即时配送功能，支持同城即时配送、预约配送、自提柜、配送员实时定位。
// 假设：即时配送默认30分钟内响应，自提柜保留48小时。
package domain

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// --- 即时配送类型 ---

// InstantDeliveryType 即时配送类型
type InstantDeliveryType int

const (
	InstantTypeExpress    InstantDeliveryType = 1 // 极速达（2小时内）
	InstantTypeSameDay    InstantDeliveryType = 2 // 当日达
	InstantTypeNextDay    InstantDeliveryType = 3 // 次日达
	InstantTypeScheduled  InstantDeliveryType = 4 // 预约配送
	InstantTypeSelfPickup InstantDeliveryType = 5 // 自提
)

// --- 即时配送状态 ---

// InstantDeliveryStatus 即时配送状态
type InstantDeliveryStatus int

const (
	InstantStatusPending      InstantDeliveryStatus = 1  // 待接单
	InstantStatusAccepted     InstantDeliveryStatus = 2  // 已接单
	InstantStatusArrivedShop  InstantDeliveryStatus = 3  // 已到店
	InstantStatusPickedUp     InstantDeliveryStatus = 4  // 已取货
	InstantStatusDelivering   InstantDeliveryStatus = 5  // 配送中
	InstantStatusArrived      InstantDeliveryStatus = 6  // 已到达
	InstantStatusDelivered    InstantDeliveryStatus = 7  // 已送达
	InstantStatusStoredLocker InstantDeliveryStatus = 8  // 已存柜
	InstantStatusSelfPickedUp InstantDeliveryStatus = 9  // 已自提
	InstantStatusAbnormal     InstantDeliveryStatus = 10 // 异常
	InstantStatusCancelled    InstantDeliveryStatus = 11 // 已取消
)

// --- 即时配送聚合根 ---

// InstantDelivery 即时配送聚合根
type InstantDelivery struct {
	ID                    uint64                `json:"id"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	DeliveryNo            string                `json:"delivery_no"` // 配送单号
	OrderID               uint64                `json:"order_id"`
	OrderNo               string                `json:"order_no"`
	Type                  InstantDeliveryType   `json:"type"`
	Status                InstantDeliveryStatus `json:"status"`
	UserID                uint64                `json:"user_id"`
	MerchantID            uint64                `json:"merchant_id"`
	CourierID             string                `json:"courier_id"`              // 配送员ID
	CourierName           string                `json:"courier_name"`            // 配送员姓名
	CourierPhone          string                `json:"courier_phone"`           // 配送员电话
	CourierLocation       *GeoLocation          `json:"courier_location"`        // 配送员实时位置
	ShopAddress           *GeoAddress           `json:"shop_address"`            // 取货地址
	DeliveryAddress       *GeoAddress           `json:"delivery_address"`        // 配送地址
	EstimatedPickupTime   *time.Time            `json:"estimated_pickup_time"`   // 预计取货时间
	EstimatedDeliveryTime *time.Time            `json:"estimated_delivery_time"` // 预计送达时间
	ActualPickupTime      *time.Time            `json:"actual_pickup_time"`      // 实际取货时间
	ActualDeliveryTime    *time.Time            `json:"actual_delivery_time"`    // 实际送达时间
	ScheduledDeliveryTime *time.Time            `json:"scheduled_delivery_time"` // 预约配送时间（预约配送用）
	Distance              float64               `json:"distance"`                // 配送距离（公里）
	DeliveryFee           int64                 `json:"delivery_fee"`            // 配送费（分）
	Tips                  int64                 `json:"tips"`                    // 小费（分）
	LockerInfo            *LockerInfo           `json:"locker_info"`             // 自提柜信息
	Traces                []*DeliveryTrace      `json:"traces"`                  // 配送轨迹
	Remark                string                `json:"remark"`
}

// GeoLocation 地理位置
type GeoLocation struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
	Accuracy  float64   `json:"accuracy"` // 精度（米）
	Heading   float64   `json:"heading"`  // 方向（度）
	Speed     float64   `json:"speed"`    // 速度（km/h）
}

// GeoAddress 地理地址
type GeoAddress struct {
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Street       string  `json:"street"`
	DetailAddr   string  `json:"detail_addr"`
	ContactName  string  `json:"contact_name"`
	ContactPhone string  `json:"contact_phone"`
	POI          string  `json:"poi"` // 门店/小区名
}

// LockerInfo 自提柜信息
type LockerInfo struct {
	LockerID      string     `json:"locker_id"`      // 柜机ID
	LockerName    string     `json:"locker_name"`    // 柜机名称
	LockerAddress string     `json:"locker_address"` // 柜机地址
	BoxNo         string     `json:"box_no"`         // 格口号
	PickupCode    string     `json:"pickup_code"`    // 取件码
	StoredAt      *time.Time `json:"stored_at"`      // 存入时间
	ExpireAt      *time.Time `json:"expire_at"`      // 过期时间
	PickedUpAt    *time.Time `json:"picked_up_at"`   // 取件时间
}

// DeliveryTrace 配送轨迹
type DeliveryTrace struct {
	ID          uint64       `json:"id"`
	Timestamp   time.Time    `json:"timestamp"`
	Status      string       `json:"status"`
	Location    *GeoLocation `json:"location"`
	Description string       `json:"description"`
}

// NewInstantDelivery 创建即时配送
func NewInstantDelivery(deliveryNo, orderNo string, orderID, userID, merchantID uint64, deliveryType InstantDeliveryType, shopAddr, deliveryAddr *GeoAddress) *InstantDelivery {
	d := &InstantDelivery{
		DeliveryNo:      deliveryNo,
		OrderID:         orderID,
		OrderNo:         orderNo,
		Type:            deliveryType,
		Status:          InstantStatusPending,
		UserID:          userID,
		MerchantID:      merchantID,
		ShopAddress:     shopAddr,
		DeliveryAddress: deliveryAddr,
		Traces:          make([]*DeliveryTrace, 0),
	}
	d.calculateDistance()
	d.calculateDeliveryFee()
	d.estimateDeliveryTime()
	return d
}

// calculateDistance 计算配送距离（Haversine公式）
func (d *InstantDelivery) calculateDistance() {
	if d.ShopAddress == nil || d.DeliveryAddress == nil {
		return
	}

	// Haversine公式计算球面距离
	lat1 := d.ShopAddress.Latitude * math.Pi / 180
	lat2 := d.DeliveryAddress.Latitude * math.Pi / 180
	deltaLat := (d.DeliveryAddress.Latitude - d.ShopAddress.Latitude) * math.Pi / 180
	deltaLon := (d.DeliveryAddress.Longitude - d.ShopAddress.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	earthRadius := 6371.0 // 地球半径（公里）
	d.Distance = earthRadius * c
}

// calculateDeliveryFee 计算配送费
func (d *InstantDelivery) calculateDeliveryFee() {
	// 基础费用 + 距离费用
	baseFee := int64(500) // 5元起步
	distanceFee := int64(0)

	if d.Distance > 3 {
		// 超过3公里，每公里加1元
		distanceFee = int64((d.Distance - 3) * 100)
	}

	// 极速达加价
	if d.Type == InstantTypeExpress {
		baseFee += 300 // 加3元
	}

	d.DeliveryFee = baseFee + distanceFee
}

// estimateDeliveryTime 预估送达时间
func (d *InstantDelivery) estimateDeliveryTime() {
	now := time.Now()

	switch d.Type {
	case InstantTypeExpress:
		// 极速达：30分钟取货 + 距离/30公里每小时
		pickupTime := now.Add(30 * time.Minute)
		deliveryMinutes := int(d.Distance / 30 * 60)
		if deliveryMinutes < 30 {
			deliveryMinutes = 30
		}
		deliveryTime := pickupTime.Add(time.Duration(deliveryMinutes) * time.Minute)
		d.EstimatedPickupTime = &pickupTime
		d.EstimatedDeliveryTime = &deliveryTime
	case InstantTypeSameDay:
		// 当日达：当天18:00前
		deliveryTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
		if now.After(deliveryTime) {
			// 如果已过18点，则次日
			deliveryTime = deliveryTime.Add(24 * time.Hour)
		}
		d.EstimatedDeliveryTime = &deliveryTime
	case InstantTypeNextDay:
		// 次日达：次日18:00前
		deliveryTime := time.Date(now.Year(), now.Month(), now.Day()+1, 18, 0, 0, 0, now.Location())
		d.EstimatedDeliveryTime = &deliveryTime
	}
}

// AssignCourier 分配配送员
func (d *InstantDelivery) AssignCourier(ctx context.Context, courierID, courierName, courierPhone string) error {
	if d.Status != InstantStatusPending {
		return errors.New("can only assign courier for pending status")
	}

	d.CourierID = courierID
	d.CourierName = courierName
	d.CourierPhone = courierPhone
	d.Status = InstantStatusAccepted

	d.addTrace(InstantStatusAccepted.String(), nil, fmt.Sprintf("Courier %s accepted", courierName))
	return nil
}

// UpdateCourierLocation 更新配送员位置
func (d *InstantDelivery) UpdateCourierLocation(location *GeoLocation) {
	d.CourierLocation = location
	d.addTrace("LOCATION_UPDATE", location, "Courier location updated")
}

// ArriveShop 到达取货点
func (d *InstantDelivery) ArriveShop(ctx context.Context) error {
	if d.Status != InstantStatusAccepted {
		return errors.New("can only arrive shop for accepted status")
	}

	d.Status = InstantStatusArrivedShop
	d.addTrace(InstantStatusArrivedShop.String(), d.CourierLocation, "Courier arrived at shop")
	return nil
}

// PickupGoods 取货
func (d *InstantDelivery) PickupGoods(ctx context.Context) error {
	if d.Status != InstantStatusArrivedShop {
		return errors.New("can only pickup for arrived shop status")
	}

	now := time.Now()
	d.Status = InstantStatusPickedUp
	d.ActualPickupTime = &now

	d.addTrace(InstantStatusPickedUp.String(), d.CourierLocation, "Goods picked up")
	return nil
}

// StartDelivering 开始配送
func (d *InstantDelivery) StartDelivering(ctx context.Context) error {
	if d.Status != InstantStatusPickedUp {
		return errors.New("can only start delivering for picked up status")
	}

	d.Status = InstantStatusDelivering
	d.addTrace(InstantStatusDelivering.String(), d.CourierLocation, "Delivery started")
	return nil
}

// ArriveDestination 到达目的地
func (d *InstantDelivery) ArriveDestination(ctx context.Context) error {
	if d.Status != InstantStatusDelivering {
		return errors.New("can only arrive for delivering status")
	}

	d.Status = InstantStatusArrived
	d.addTrace(InstantStatusArrived.String(), d.CourierLocation, "Courier arrived at destination")
	return nil
}

// ConfirmDelivery 确认送达
func (d *InstantDelivery) ConfirmDelivery(ctx context.Context, signature string) error {
	if d.Status != InstantStatusArrived {
		return errors.New("can only confirm delivery for arrived status")
	}

	now := time.Now()
	d.Status = InstantStatusDelivered
	d.ActualDeliveryTime = &now

	d.addTrace(InstantStatusDelivered.String(), d.CourierLocation, fmt.Sprintf("Delivered, signature: %s", signature))
	return nil
}

// StoreToLocker 存入自提柜
func (d *InstantDelivery) StoreToLocker(ctx context.Context, lockerInfo *LockerInfo) error {
	if d.Status != InstantStatusArrived && d.Status != InstantStatusDelivering {
		return errors.New("invalid status for store to locker")
	}

	now := time.Now()
	expireAt := now.Add(48 * time.Hour) // 48小时过期

	lockerInfo.StoredAt = &now
	lockerInfo.ExpireAt = &expireAt
	d.LockerInfo = lockerInfo
	d.Status = InstantStatusStoredLocker

	d.addTrace(InstantStatusStoredLocker.String(), nil, fmt.Sprintf("Stored to locker %s, box %s, code %s", lockerInfo.LockerName, lockerInfo.BoxNo, lockerInfo.PickupCode))
	return nil
}

// ConfirmSelfPickup 确认自提
func (d *InstantDelivery) ConfirmSelfPickup(ctx context.Context, pickupCode string) error {
	if d.Status != InstantStatusStoredLocker {
		return errors.New("can only self pickup for stored locker status")
	}
	if d.LockerInfo == nil || d.LockerInfo.PickupCode != pickupCode {
		return errors.New("invalid pickup code")
	}

	now := time.Now()
	d.LockerInfo.PickedUpAt = &now
	d.Status = InstantStatusSelfPickedUp
	d.ActualDeliveryTime = &now

	d.addTrace(InstantStatusSelfPickedUp.String(), nil, "Self pickup completed")
	return nil
}

// MarkAbnormal 标记异常
func (d *InstantDelivery) MarkAbnormal(ctx context.Context, reason string) error {
	d.Status = InstantStatusAbnormal
	d.addTrace(InstantStatusAbnormal.String(), d.CourierLocation, reason)
	return nil
}

// Cancel 取消配送
func (d *InstantDelivery) Cancel(ctx context.Context, reason string) error {
	if d.Status >= InstantStatusDelivered {
		return errors.New("cannot cancel completed delivery")
	}

	d.Status = InstantStatusCancelled
	d.addTrace(InstantStatusCancelled.String(), nil, reason)
	return nil
}

// GetRemainingDistance 获取剩余配送距离
func (d *InstantDelivery) GetRemainingDistance() float64 {
	if d.CourierLocation == nil || d.DeliveryAddress == nil {
		return d.Distance
	}

	// 计算配送员当前位置到目的地的距离
	lat1 := d.CourierLocation.Latitude * math.Pi / 180
	lat2 := d.DeliveryAddress.Latitude * math.Pi / 180
	deltaLat := (d.DeliveryAddress.Latitude - d.CourierLocation.Latitude) * math.Pi / 180
	deltaLon := (d.DeliveryAddress.Longitude - d.CourierLocation.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6371.0 * c
}

// AIPredictionConfig AI 预测配置
type AIPredictionConfig struct {
	WeatherFactor float64 `json:"weather_factor"` // 天气系数 (1.0 = 晴, 1.5 = 雨)
	TrafficFactor float64 `json:"traffic_factor"` // 交通系数 (1.0 = 通畅, 2.0 = 拥堵)
	HistoryWeight float64 `json:"history_weight"` // 历史权重
}

// PredictDeliveryTimeAI 基于 AI 模型预估配送时效（模拟实现）
func (d *InstantDelivery) PredictDeliveryTimeAI(ctx context.Context, config AIPredictionConfig) *time.Time {
	if d.Distance <= 0 {
		return d.EstimatedDeliveryTime
	}

	// 基础时间：按 20km/h 速度计算
	baseMinutes := (d.Distance / 20.0) * 60.0

	// 叠加环境因子
	adjustedMinutes := baseMinutes * config.WeatherFactor * config.TrafficFactor

	// 加上人工取货环节预估（平均 15 分钟）
	totalMinutes := adjustedMinutes + 15

	eta := time.Now().Add(time.Duration(totalMinutes) * time.Minute)

	// 更新领域模型中的预估时效
	d.EstimatedDeliveryTime = &eta

	d.addTrace("AI_PREDICTION", nil, fmt.Sprintf("AI Predicted ETA: %s (Factors: W=%.1f, T=%.1f)", eta.Format("15:04"), config.WeatherFactor, config.TrafficFactor))

	return &eta
}

// addTrace 添加配送轨迹
func (d *InstantDelivery) addTrace(status string, location *GeoLocation, description string) {
	d.Traces = append(d.Traces, &DeliveryTrace{
		Timestamp:   time.Now(),
		Status:      status,
		Location:    location,
		Description: description,
	})
}

// String 返回状态字符串
func (s InstantDeliveryStatus) String() string {
	names := []string{"", "PENDING", "ACCEPTED", "ARRIVED_SHOP", "PICKED_UP", "DELIVERING", "ARRIVED", "DELIVERED", "STORED_LOCKER", "SELF_PICKED_UP", "ABNORMAL", "CANCELLED"}
	if int(s) < len(names) {
		return names[s]
	}
	return "UNKNOWN"
}

// --- 即时配送仓储接口 ---

// InstantDeliveryRepository 即时配送仓储接口
type InstantDeliveryRepository interface {
	Save(ctx context.Context, delivery *InstantDelivery) error
	Update(ctx context.Context, delivery *InstantDelivery) error
	FindByID(ctx context.Context, id uint64) (*InstantDelivery, error)
	FindByDeliveryNo(ctx context.Context, deliveryNo string) (*InstantDelivery, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*InstantDelivery, error)
	FindByCourierID(ctx context.Context, courierID string, status []InstantDeliveryStatus) ([]*InstantDelivery, error)
	FindPendingInArea(ctx context.Context, lat, lon, radiusKm float64) ([]*InstantDelivery, error)
}

// --- 配送员调度 ---

// CourierDispatcher 配送员调度器
type CourierDispatcher struct {
	MaxDistanceKm float64 // 最大接单距离
}

// Courier 配送员
type Courier struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Phone         string       `json:"phone"`
	Location      *GeoLocation `json:"location"`
	Status        string       `json:"status"`         // ONLINE/OFFLINE/BUSY
	Rating        float64      `json:"rating"`         // 评分
	TodayOrders   int          `json:"today_orders"`   // 今日订单数
	MaxOrders     int          `json:"max_orders"`     // 最大同时接单数
	CurrentOrders int          `json:"current_orders"` // 当前进行中订单数
}

// FindBestCourier 寻找最优配送员
func (d *CourierDispatcher) FindBestCourier(ctx context.Context, shopLocation *GeoLocation, availableCouriers []*Courier) *Courier {
	if len(availableCouriers) == 0 {
		return nil
	}

	var bestCourier *Courier
	bestScore := float64(-1)

	for _, courier := range availableCouriers {
		if courier.Status != "ONLINE" || courier.CurrentOrders >= courier.MaxOrders {
			continue
		}

		// 计算距离
		distance := d.calculateDistance(shopLocation, courier.Location)
		if distance > d.MaxDistanceKm {
			continue
		}

		// 计算得分：距离越近、评分越高、订单数越少，得分越高
		score := 100 - distance*10 + courier.Rating*5 - float64(courier.CurrentOrders)*10
		if score > bestScore {
			bestScore = score
			bestCourier = courier
		}
	}

	return bestCourier
}

// calculateDistance 计算两点之间的距离
func (d *CourierDispatcher) calculateDistance(loc1, loc2 *GeoLocation) float64 {
	if loc1 == nil || loc2 == nil {
		return math.MaxFloat64
	}

	lat1 := loc1.Latitude * math.Pi / 180
	lat2 := loc2.Latitude * math.Pi / 180
	deltaLat := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
	deltaLon := (loc2.Longitude - loc1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6371.0 * c
}

// --- 自提柜管理 ---

// SmartLocker 智能快递柜
type SmartLocker struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Address  string       `json:"address"`
	Location *GeoLocation `json:"location"`
	Boxes    []*LockerBox `json:"boxes"`
	Status   string       `json:"status"` // ONLINE/OFFLINE/MAINTENANCE
}

// LockerBox 柜机格口
type LockerBox struct {
	BoxNo   string `json:"box_no"`
	Size    string `json:"size"`   // SMALL/MEDIUM/LARGE
	Status  string `json:"status"` // EMPTY/OCCUPIED/RESERVED/FAULT
	OrderID uint64 `json:"order_id"`
}

// FindAvailableBox 查找可用格口
func (l *SmartLocker) FindAvailableBox(size string) *LockerBox {
	for _, box := range l.Boxes {
		if box.Status == "EMPTY" && box.Size == size {
			return box
		}
	}
	// 如果没有指定尺寸，尝试找更大的
	sizeOrder := []string{"SMALL", "MEDIUM", "LARGE"}
	startIdx := 0
	for i, s := range sizeOrder {
		if s == size {
			startIdx = i + 1
			break
		}
	}
	for i := startIdx; i < len(sizeOrder); i++ {
		for _, box := range l.Boxes {
			if box.Status == "EMPTY" && box.Size == sizeOrder[i] {
				return box
			}
		}
	}
	return nil
}

// ReserveBox 预留格口
func (l *SmartLocker) ReserveBox(boxNo string, orderID uint64) error {
	for _, box := range l.Boxes {
		if box.BoxNo == boxNo {
			if box.Status != "EMPTY" {
				return errors.New("box is not available")
			}
			box.Status = "RESERVED"
			box.OrderID = orderID
			return nil
		}
	}
	return errors.New("box not found")
}

// OccupyBox 占用格口
func (l *SmartLocker) OccupyBox(boxNo string) error {
	for _, box := range l.Boxes {
		if box.BoxNo == boxNo {
			box.Status = "OCCUPIED"
			return nil
		}
	}
	return errors.New("box not found")
}

// ReleaseBox 释放格口
func (l *SmartLocker) ReleaseBox(boxNo string) error {
	for _, box := range l.Boxes {
		if box.BoxNo == boxNo {
			box.Status = "EMPTY"
			box.OrderID = 0
			return nil
		}
	}
	return errors.New("box not found")
}
