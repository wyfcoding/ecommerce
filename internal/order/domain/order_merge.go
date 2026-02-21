package domain

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// OrderMergeService 订单合并服务
type OrderMergeService interface {
	// FindMergeableOrders 查找可合并的订单
	FindMergeableOrders(ctx context.Context, userID uint64, merchantID uint64, status pb.OrderStatus) ([]*Order, error)

	// CreateMergeBatch 创建合并批次
	CreateMergeBatch(ctx context.Context, orderIDs []uint64, operator string) (*MergeBatch, error)

	// ExecuteMerge 执行订单合并
	ExecuteMerge(ctx context.Context, batchID uint64, operator string) error

	// SplitMergeBatch 拆分合并批次
	SplitMergeBatch(ctx context.Context, batchID uint64, reason string, operator string) error
}

// MergeBatch 合并批次
type MergeBatch struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	BatchNo     string     `json:"batch_no"`
	MerchantID  uint64     `json:"merchant_id"`
	UserID      uint64     `json:"user_id"`
	Status      string     `json:"status"` // PENDING, MERGED, SPLIT
	OrderCount  int        `json:"order_count"`
	TotalAmount int64      `json:"total_amount"`
	ShippingFee int64      `json:"shipping_fee"`
	Remark      string     `json:"remark"`
	MergedAt    *time.Time `json:"merged_at"`
	SplitAt     *time.Time `json:"split_at"`
}

// OrderMergeRule 订单合并规则
type OrderMergeRule struct {
	ID               uint          `json:"id"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	RuleName         string        `json:"rule_name"`
	MerchantID       uint64        `json:"merchant_id"`
	Enabled          bool          `json:"enabled"`
	MaxMergeInterval time.Duration `json:"max_merge_interval"` // 最大合并间隔
	MaxMergeAmount   int64         `json:"max_merge_amount"`   // 最大合并金额
	MinMergeAmount   int64         `json:"min_merge_amount"`   // 最小合并金额
	MaxMergeCount    int           `json:"max_merge_count"`    // 最大合并订单数
	AutoMergeEnabled bool          `json:"auto_merge_enabled"` // 是否自动合并
	Priority         int           `json:"priority"`           // 优先级
	Conditions       string        `json:"conditions"`         // 合并条件 (JSON)
}

// OrderMergeCriteria 订单合并条件
type OrderMergeCriteria struct {
	// 时间条件
	MaxTimeGap time.Duration `json:"max_time_gap"`

	// 金额条件
	MinTotalAmount int64 `json:"min_total_amount"`
	MaxTotalAmount int64 `json:"max_total_amount"`

	// 商品条件
	AllowDifferentProducts bool `json:"allow_different_products"`
	MaxProductTypes        int  `json:"max_product_types"`

	// 物流条件
	SameShippingAddress bool `json:"same_shipping_address"`
	SameLogistics       bool `json:"same_logistics"`

	// 支付条件
	SamePaymentMethod bool `json:"same_payment_method"`

	// 状态条件
	AllowedStatuses []pb.OrderStatus `json:"allowed_statuses"`
}

// MergeBatchItem 合并批次项
type MergeBatchItem struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	BatchID   uint64     `json:"batch_id"`
	OrderID   uint64     `json:"order_id"`
	OrderNo   string     `json:"order_no"`
	Status    string     `json:"status"` // PENDING, MERGED, SPLIT
	MergedAt  *time.Time `json:"merged_at"`
}

// OrderMergeManager 订单合并管理器
type OrderMergeManager struct {
	orderRepo OrderRepository
}

// NewOrderMergeManager 创建订单合并管理器
func NewOrderMergeManager(orderRepo OrderRepository) *OrderMergeManager {
	return &OrderMergeManager{
		orderRepo: orderRepo,
	}
}

// CheckMergeEligibility 检查订单合并资格
func (m *OrderMergeManager) CheckMergeEligibility(order1, order2 *Order, criteria *OrderMergeCriteria) (bool, string) {
	// 检查商家是否相同
	if order1.Items[0].MerchantID != order2.Items[0].MerchantID {
		return false, "different merchants"
	}

	// 检查用户是否相同
	if order1.UserID != order2.UserID {
		return false, "different users"
	}

	// 检查时间间隔
	timeGap := order2.CreatedAt.Sub(order1.CreatedAt)
	if timeGap < 0 {
		timeGap = -timeGap
	}
	if timeGap > criteria.MaxTimeGap {
		return false, fmt.Sprintf("time gap too large: %v", timeGap)
	}

	// 检查总金额
	totalAmount := order1.ActualAmount + order2.ActualAmount
	if totalAmount < criteria.MinTotalAmount {
		return false, fmt.Sprintf("total amount too small: %d", totalAmount)
	}
	if totalAmount > criteria.MaxTotalAmount {
		return false, fmt.Sprintf("total amount too large: %d", totalAmount)
	}

	// 检查收货地址
	if criteria.SameShippingAddress {
		if !compareShippingAddress(order1.ShippingAddress, order2.ShippingAddress) {
			return false, "different shipping addresses"
		}
	}

	// 检查支付方式
	if criteria.SamePaymentMethod && order1.PaymentMethod != order2.PaymentMethod {
		return false, "different payment methods"
	}

	// 检查订单状态
	statusAllowed := false
	for _, status := range criteria.AllowedStatuses {
		if order1.Status == status && order2.Status == status {
			statusAllowed = true
			break
		}
	}
	if !statusAllowed {
		return false, "order status not allowed for merge"
	}

	return true, ""
}

// FindMergeableOrders 查找可合并的订单
func (m *OrderMergeManager) FindMergeableOrders(ctx context.Context, userID uint64, merchantID uint64, criteria *OrderMergeCriteria) ([]*Order, error) {
	// 获取用户的所有待支付或已支付订单
	orders, err := m.orderRepo.FindByUserAndMerchant(ctx, userID, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}

	// 按创建时间排序
	sortedOrders := sortOrdersByTime(orders)

	// 查找可合并的订单组
	var mergeableGroups [][]*Order
	currentGroup := []*Order{sortedOrders[0]}

	for i := 1; i < len(sortedOrders); i++ {
		order := sortedOrders[i]

		// 检查是否可以加入当前组
		canMerge := true
		for _, groupOrder := range currentGroup {
			eligible, _ := m.CheckMergeEligibility(groupOrder, order, criteria)
			if !eligible {
				canMerge = false
				break
			}
		}

		if canMerge {
			currentGroup = append(currentGroup, order)
		} else {
			if len(currentGroup) > 1 {
				mergeableGroups = append(mergeableGroups, currentGroup)
			}
			currentGroup = []*Order{order}
		}
	}

	// 处理最后一组
	if len(currentGroup) > 1 {
		mergeableGroups = append(mergeableGroups, currentGroup)
	}

	// 展平结果
	var result []*Order
	for _, group := range mergeableGroups {
		result = append(result, group...)
	}

	return result, nil
}

// CreateMergeBatch 创建合并批次
func (m *OrderMergeManager) CreateMergeBatch(ctx context.Context, orderIDs []uint64, operator string) (*MergeBatch, error) {
	if len(orderIDs) < 2 {
		return nil, fmt.Errorf("at least 2 orders required for merge")
	}

	// 获取订单信息
	var orders []*Order
	var totalAmount int64
	var userID uint64
	var merchantID uint64

	for _, orderID := range orderIDs {
		order, err := m.orderRepo.FindByID(ctx, 0, orderID) // 暂时使用0作为userID
		if err != nil {
			return nil, fmt.Errorf("failed to find order %d: %w", orderID, err)
		}

		if userID == 0 {
			userID = order.UserID
			merchantID = order.Items[0].MerchantID
		} else if order.UserID != userID || order.Items[0].MerchantID != merchantID {
			return nil, fmt.Errorf("orders must belong to same user and merchant")
		}

		orders = append(orders, order)
		totalAmount += order.ActualAmount
	}

	// 创建合并批次
	batch := &MergeBatch{
		BatchNo:     generateBatchNo(),
		MerchantID:  merchantID,
		UserID:      userID,
		Status:      "PENDING",
		OrderCount:  len(orderIDs),
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存合并批次
	// TODO: 实现合并批次的保存逻辑

	// 更新订单状态为待合并
	for _, order := range orders {
		order.AddLog(operator, "MERGE_PENDING", order.Status.String(), order.Status.String(), "Order marked for merge")
		err := m.orderRepo.Update(ctx, order)
		if err != nil {
			return nil, fmt.Errorf("failed to update order %s: %w", order.OrderNo, err)
		}
	}

	return batch, nil
}

// ExecuteMerge 执行订单合并
func (m *OrderMergeManager) ExecuteMerge(ctx context.Context, batchID uint64, operator string) error {
	// TODO: 实现订单合并逻辑
	// 1. 验证批次状态
	// 2. 更新订单状态为已合并
	// 3. 生成合并发货单
	// 4. 更新库存
	// 5. 记录合并日志

	return nil
}

// SplitMergeBatch 拆分合并批次
func (m *OrderMergeManager) SplitMergeBatch(ctx context.Context, batchID uint64, reason string, operator string) error {
	// TODO: 实现合并批次拆分逻辑
	// 1. 验证批次状态
	// 2. 恢复订单原始状态
	// 3. 更新批次状态为已拆分
	// 4. 记录拆分日志

	return nil
}

// Helper functions

func compareShippingAddress(addr1, addr2 *ShippingAddress) bool {
	if addr1 == nil || addr2 == nil {
		return addr1 == addr2
	}

	return addr1.RecipientName == addr2.RecipientName &&
		addr1.PhoneNumber == addr2.PhoneNumber &&
		addr1.Province == addr2.Province &&
		addr1.City == addr2.City &&
		addr1.District == addr2.District &&
		addr1.DetailedAddress == addr2.DetailedAddress &&
		addr1.PostalCode == addr2.PostalCode
}

func sortOrdersByTime(orders []*Order) []*Order {
	// 使用插入排序按创建时间升序排列
	sorted := make([]*Order, len(orders))
	copy(sorted, orders)

	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1

		for j >= 0 && sorted[j].CreatedAt.After(key.CreatedAt) {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	return sorted
}

func generateBatchNo() string {
	return fmt.Sprintf("MB%d", time.Now().UnixNano())
}
