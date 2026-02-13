// 变更说明：新增订单拆分逻辑，支持跨店满减拆分、组合商品拆分、多仓发货拆分。
// 假设：跨店订单按商家维度拆分，多仓订单按仓库维度拆分，套餐商品按SKU维度拆分。
package domain

import (
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
)

// --- 订单拆分逻辑 ---

// OrderSplitReason 订单拆分原因
type OrderSplitReason int

const (
	SplitReasonMultiMerchant  OrderSplitReason = 1 // 多商家拆单
	SplitReasonMultiWarehouse OrderSplitReason = 2 // 多仓库拆单
	SplitReasonBundleProduct  OrderSplitReason = 3 // 套餐商品拆单
	SplitReasonPromotion      OrderSplitReason = 4 // 跨店满减拆单
	SplitReasonLogistics      OrderSplitReason = 5 // 物流限制拆单（超重/超体积）
)

// OrderSplitContext 订单拆分上下文
type OrderSplitContext struct {
	OriginalOrder  *Order           `json:"original_order"`
	SplitOrders    []*Order         `json:"split_orders"`
	SplitReason    OrderSplitReason `json:"split_reason"`
	PromotionAlloc map[string]int64 `json:"promotion_alloc"` // 优惠分摊（子订单号 -> 分摊金额）
	ShippingAlloc  map[string]int64 `json:"shipping_alloc"`  // 运费分摊
}

// OrderSplitter 订单拆分器接口
type OrderSplitter interface {
	// Split 执行订单拆分
	Split(order *Order) (*OrderSplitContext, error)
	// CanSplit 判断订单是否可拆分
	CanSplit(order *Order) bool
}

// --- 多商家拆单 ---

// MerchantOrderSplitter 多商家拆单器
type MerchantOrderSplitter struct {
	OrderNoGenerator func() string
}

// CanSplit 判断是否需要多商家拆单
func (s *MerchantOrderSplitter) CanSplit(order *Order) bool {
	merchantIDs := make(map[uint64]bool)
	for _, item := range order.Items {
		// 假设 OrderItem 有 MerchantID 字段
		if item.MerchantID != 0 {
			merchantIDs[item.MerchantID] = true
		}
	}
	return len(merchantIDs) > 1
}

// Split 执行多商家拆单
func (s *MerchantOrderSplitter) Split(order *Order) (*OrderSplitContext, error) {
	if !s.CanSplit(order) {
		return nil, fmt.Errorf("order does not need merchant split")
	}

	// 按商家分组商品
	merchantItems := make(map[uint64][]*OrderItem)
	for _, item := range order.Items {
		merchantItems[item.MerchantID] = append(merchantItems[item.MerchantID], item)
	}

	ctx := &OrderSplitContext{
		OriginalOrder:  order,
		SplitOrders:    make([]*Order, 0, len(merchantItems)),
		SplitReason:    SplitReasonMultiMerchant,
		PromotionAlloc: make(map[string]int64),
		ShippingAlloc:  make(map[string]int64),
	}

	totalAmount := order.TotalAmount
	promotionAmount := order.DiscountAmount
	shippingFee := order.ShippingFee

	for merchantID, items := range merchantItems {
		subOrderNo := s.OrderNoGenerator()
		subOrder := &Order{
			OrderNo:         subOrderNo,
			UserID:          order.UserID,
			Status:          order.Status,
			PaymentStatus:   order.PaymentStatus,
			ShippingStatus:  order.ShippingStatus,
			ShippingAddress: order.ShippingAddress,
			Items:           items,
			Remark:          fmt.Sprintf("拆单自 %s，商家ID: %d", order.OrderNo, merchantID),
		}

		// 计算子订单金额
		var subTotal int64
		for _, item := range items {
			item.TotalPrice = item.Price * int64(item.Quantity)
			subTotal += item.TotalPrice
		}
		subOrder.TotalAmount = subTotal

		// 按金额比例分摊优惠和运费
		ratio := float64(subTotal) / float64(totalAmount)
		allocatedPromotion := int64(float64(promotionAmount) * ratio)
		allocatedShipping := int64(float64(shippingFee) * ratio)

		subOrder.DiscountAmount = allocatedPromotion
		subOrder.ShippingFee = allocatedShipping
		subOrder.ActualAmount = subTotal - allocatedPromotion + allocatedShipping

		ctx.SplitOrders = append(ctx.SplitOrders, subOrder)
		ctx.PromotionAlloc[subOrderNo] = allocatedPromotion
		ctx.ShippingAlloc[subOrderNo] = allocatedShipping
	}

	return ctx, nil
}

// --- 多仓库拆单 ---

// WarehouseOrderSplitter 多仓库拆单器
type WarehouseOrderSplitter struct {
	OrderNoGenerator func() string
	WarehouseFinder  func(skuID uint64) uint64 // 根据SKU找最优仓库
}

// CanSplit 判断是否需要多仓库拆单
func (s *WarehouseOrderSplitter) CanSplit(order *Order) bool {
	warehouseIDs := make(map[uint64]bool)
	for _, item := range order.Items {
		warehouseID := s.WarehouseFinder(item.SkuID)
		warehouseIDs[warehouseID] = true
	}
	return len(warehouseIDs) > 1
}

// Split 执行多仓库拆单
func (s *WarehouseOrderSplitter) Split(order *Order) (*OrderSplitContext, error) {
	if !s.CanSplit(order) {
		return nil, fmt.Errorf("order does not need warehouse split")
	}

	// 按仓库分组商品
	warehouseItems := make(map[uint64][]*OrderItem)
	for _, item := range order.Items {
		warehouseID := s.WarehouseFinder(item.SkuID)
		warehouseItems[warehouseID] = append(warehouseItems[warehouseID], item)
	}

	ctx := &OrderSplitContext{
		OriginalOrder:  order,
		SplitOrders:    make([]*Order, 0, len(warehouseItems)),
		SplitReason:    SplitReasonMultiWarehouse,
		PromotionAlloc: make(map[string]int64),
		ShippingAlloc:  make(map[string]int64),
	}

	for warehouseID, items := range warehouseItems {
		subOrderNo := s.OrderNoGenerator()
		subOrder := &Order{
			OrderNo:         subOrderNo,
			UserID:          order.UserID,
			Status:          order.Status,
			PaymentStatus:   order.PaymentStatus,
			ShippingStatus:  order.ShippingStatus,
			ShippingAddress: order.ShippingAddress,
			Items:           items,
			Remark:          fmt.Sprintf("拆单自 %s，仓库ID: %d", order.OrderNo, warehouseID),
		}

		// 计算子订单金额
		var subTotal int64
		for _, item := range items {
			subTotal += item.TotalPrice
		}
		subOrder.TotalAmount = subTotal
		subOrder.ActualAmount = subTotal // 简化处理，实际需要按比例分摊优惠

		ctx.SplitOrders = append(ctx.SplitOrders, subOrder)
	}

	return ctx, nil
}

// --- 套餐商品拆单 ---

// BundleProduct 套餐商品
type BundleProduct struct {
	ID          uint64             `json:"id"`
	Name        string             `json:"name"`
	BundlePrice int64              `json:"bundle_price"` // 套餐价
	Components  []*BundleComponent `json:"components"`   // 组成商品
}

// BundleComponent 套餐组成商品
type BundleComponent struct {
	SkuID         uint64 `json:"sku_id"`
	ProductID     uint64 `json:"product_id"`
	ProductName   string `json:"product_name"`
	Quantity      int32  `json:"quantity"`
	OriginalPrice int64  `json:"original_price"` // 原价（用于分摊计算）
}

// BundleOrderSplitter 套餐拆单器
type BundleOrderSplitter struct {
	BundleFinder func(productID uint64) *BundleProduct // 查找套餐定义
}

// ExpandBundleItems 展开套餐商品为明细商品
func (s *BundleOrderSplitter) ExpandBundleItems(items []*OrderItem) ([]*OrderItem, error) {
	expandedItems := make([]*OrderItem, 0)

	for _, item := range items {
		bundle := s.BundleFinder(item.ProductID)
		if bundle == nil {
			// 非套餐商品，直接保留
			expandedItems = append(expandedItems, item)
			continue
		}

		// 计算套餐总原价（用于分摊）
		var totalOriginalPrice int64
		for _, comp := range bundle.Components {
			totalOriginalPrice += comp.OriginalPrice * int64(comp.Quantity)
		}

		// 展开套餐为明细商品
		for _, comp := range bundle.Components {
			// 按原价比例分摊套餐价
			allocatedPrice := bundle.BundlePrice * (comp.OriginalPrice * int64(comp.Quantity)) / totalOriginalPrice

			expandedItem := &OrderItem{
				ProductID:   comp.ProductID,
				SkuID:       comp.SkuID,
				ProductName: comp.ProductName,
				Quantity:    comp.Quantity * item.Quantity, // 乘以套餐购买数量
				Price:       allocatedPrice / int64(comp.Quantity),
				TotalPrice:  allocatedPrice * int64(item.Quantity),
			}
			expandedItems = append(expandedItems, expandedItem)
		}
	}

	return expandedItems, nil
}

// --- 跨店满减拆单 ---

// CrossStorePromotion 跨店满减活动
type CrossStorePromotion struct {
	ID                     uint64   `json:"id"`
	Name                   string   `json:"name"`
	Threshold              int64    `json:"threshold"`               // 满减门槛
	DiscountAmount         int64    `json:"discount_amount"`         // 优惠金额
	ParticipatingMerchants []uint64 `json:"participating_merchants"` // 参与商家
}

// CrossStorePromotionSplitter 跨店满减拆单器
type CrossStorePromotionSplitter struct {
	OrderNoGenerator func() string
	Promotion        *CrossStorePromotion
}

// AllocateDiscount 按商家分摊优惠金额
func (s *CrossStorePromotionSplitter) AllocateDiscount(order *Order) map[uint64]int64 {
	// 按商家统计金额
	merchantAmounts := make(map[uint64]int64)
	var totalAmount int64
	for _, item := range order.Items {
		merchantAmounts[item.MerchantID] += item.TotalPrice
		totalAmount += item.TotalPrice
	}

	// 检查是否达到满减门槛
	if totalAmount < s.Promotion.Threshold {
		return nil
	}

	// 按金额比例分摊优惠
	discountAlloc := make(map[uint64]int64)
	var allocatedTotal int64
	merchants := make([]uint64, 0, len(merchantAmounts))
	for mid := range merchantAmounts {
		merchants = append(merchants, mid)
	}

	for i, merchantID := range merchants {
		if i == len(merchants)-1 {
			// 最后一个商家承担余数
			discountAlloc[merchantID] = s.Promotion.DiscountAmount - allocatedTotal
		} else {
			ratio := float64(merchantAmounts[merchantID]) / float64(totalAmount)
			allocated := int64(float64(s.Promotion.DiscountAmount) * ratio)
			discountAlloc[merchantID] = allocated
			allocatedTotal += allocated
		}
	}

	return discountAlloc
}

// --- 物流限制拆单 ---

// LogisticsLimitSplitter 物流限制拆单器（超重/超体积）
type LogisticsLimitSplitter struct {
	MaxWeight        float64 // 最大重量（kg）
	MaxVolume        float64 // 最大体积（m³）
	OrderNoGenerator func() string
}

// ItemDimension 商品尺寸信息
type ItemDimension struct {
	SkuID  uint64  `json:"sku_id"`
	Weight float64 `json:"weight"` // 重量（kg）
	Length float64 `json:"length"` // 长（cm）
	Width  float64 `json:"width"`  // 宽（cm）
	Height float64 `json:"height"` // 高（cm）
}

// CalculateVolume 计算体积（m³）
func (d *ItemDimension) CalculateVolume() float64 {
	return (d.Length * d.Width * d.Height) / 1000000 // cm³ -> m³
}

// Split 按物流限制拆单
func (s *LogisticsLimitSplitter) Split(order *Order, dimensions map[uint64]*ItemDimension) (*OrderSplitContext, error) {
	ctx := &OrderSplitContext{
		OriginalOrder: order,
		SplitOrders:   make([]*Order, 0),
		SplitReason:   SplitReasonLogistics,
	}

	var currentItems []*OrderItem
	var currentWeight, currentVolume float64

	for _, item := range order.Items {
		dim := dimensions[item.SkuID]
		if dim == nil {
			// 无尺寸信息，直接加入当前包裹
			currentItems = append(currentItems, item)
			continue
		}

		itemWeight := dim.Weight * float64(item.Quantity)
		itemVolume := dim.CalculateVolume() * float64(item.Quantity)

		// 检查是否超限
		if currentWeight+itemWeight > s.MaxWeight || currentVolume+itemVolume > s.MaxVolume {
			// 创建新包裹
			if len(currentItems) > 0 {
				subOrderNo := s.OrderNoGenerator()
				subOrder := s.createSubOrder(order, subOrderNo, currentItems)
				ctx.SplitOrders = append(ctx.SplitOrders, subOrder)
			}
			currentItems = []*OrderItem{item}
			currentWeight = itemWeight
			currentVolume = itemVolume
		} else {
			currentItems = append(currentItems, item)
			currentWeight += itemWeight
			currentVolume += itemVolume
		}
	}

	// 处理剩余商品
	if len(currentItems) > 0 {
		subOrderNo := s.OrderNoGenerator()
		subOrder := s.createSubOrder(order, subOrderNo, currentItems)
		ctx.SplitOrders = append(ctx.SplitOrders, subOrder)
	}

	return ctx, nil
}

// createSubOrder 创建子订单
func (s *LogisticsLimitSplitter) createSubOrder(parent *Order, orderNo string, items []*OrderItem) *Order {
	var subTotal int64
	for _, item := range items {
		subTotal += item.TotalPrice
	}

	return &Order{
		OrderNo:         orderNo,
		UserID:          parent.UserID,
		Status:          parent.Status,
		ShippingAddress: parent.ShippingAddress,
		Items:           items,
		TotalAmount:     subTotal,
		ActualAmount:    subTotal,
		Remark:          fmt.Sprintf("物流拆单自 %s", parent.OrderNo),
	}
}

// --- OrderItem 扩展字段 ---

// OrderItemExt 订单商品扩展（为了支持拆单）
type OrderItemExt struct {
	*OrderItem
	MerchantID   uint64  `json:"merchant_id"`    // 商家ID
	WarehouseID  uint64  `json:"warehouse_id"`   // 仓库ID
	Weight       float64 `json:"weight"`         // 重量
	Volume       float64 `json:"volume"`         // 体积
	IsBundleItem bool    `json:"is_bundle_item"` // 是否套餐商品
	BundleID     uint64  `json:"bundle_id"`      // 套餐ID
}

// --- 订单合并 ---

// OrderMerger 订单合并器（用于合并同商家的多个订单）
type OrderMerger struct {
	MergeWindow time.Duration // 合并时间窗口
}

// CanMerge 判断两个订单是否可以合并
func (m *OrderMerger) CanMerge(order1, order2 *Order) bool {
	// 同用户
	if order1.UserID != order2.UserID {
		return false
	}
	// 同收货地址
	if order1.ShippingAddress == nil || order2.ShippingAddress == nil {
		return false
	}
	if order1.ShippingAddress.DetailedAddress != order2.ShippingAddress.DetailedAddress {
		return false
	}
	// 都未支付
	if order1.PaymentStatus != pb.PaymentStatus_UNPAID || order2.PaymentStatus != pb.PaymentStatus_UNPAID {
		return false
	}
	// 在时间窗口内
	if order2.CreatedAt.Sub(order1.CreatedAt) > m.MergeWindow {
		return false
	}
	return true
}

// Merge 合并订单
func (m *OrderMerger) Merge(orders []*Order) *Order {
	if len(orders) == 0 {
		return nil
	}
	if len(orders) == 1 {
		return orders[0]
	}

	merged := &Order{
		OrderNo:         orders[0].OrderNo + "_MERGED",
		UserID:          orders[0].UserID,
		Status:          pb.OrderStatus_PENDING_PAYMENT,
		PaymentStatus:   pb.PaymentStatus_UNPAID,
		ShippingAddress: orders[0].ShippingAddress,
		Items:           make([]*OrderItem, 0),
		Remark:          "合并订单",
	}

	var totalAmount, discountAmount, shippingFee int64
	for _, order := range orders {
		merged.Items = append(merged.Items, order.Items...)
		totalAmount += order.TotalAmount
		discountAmount += order.DiscountAmount
		shippingFee += order.ShippingFee
	}

	merged.TotalAmount = totalAmount
	merged.DiscountAmount = discountAmount
	merged.ShippingFee = shippingFee
	merged.ActualAmount = totalAmount - discountAmount + shippingFee

	return merged
}
