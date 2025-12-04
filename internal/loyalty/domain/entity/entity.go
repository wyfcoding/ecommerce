package entity

import (
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Loyalty模块的业务错误。
var (
	ErrInsufficientPoints = errors.New("积分不足")  // 积分余额不足以完成操作。
	ErrPointsExpired      = errors.New("积分已过期") // 积分已超过有效期。
)

// MemberLevel 定义了会员账户的等级。
type MemberLevel string

const (
	MemberLevelBronze   MemberLevel = "BRONZE"   // 青铜会员。
	MemberLevelSilver   MemberLevel = "SILVER"   // 白银会员。
	MemberLevelGold     MemberLevel = "GOLD"     // 黄金会员。
	MemberLevelPlatinum MemberLevel = "PLATINUM" // 铂金会员。
	MemberLevelDiamond  MemberLevel = "DIAMOND"  // 钻石会员。
)

// MemberAccount 实体是忠诚度模块的聚合根。
// 它代表一个用户的会员账户信息，包含了会员等级、积分余额、消费总额等。
type MemberAccount struct {
	gorm.Model                  // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	UserID          uint64      `gorm:"not null;uniqueIndex;comment:用户ID" json:"user_id"`            // 用户ID，唯一索引，不允许为空。
	Level           MemberLevel `gorm:"type:varchar(32);default:'BRONZE';comment:会员等级" json:"level"` // 会员等级，默认为青铜。
	TotalPoints     int64       `gorm:"not null;default:0;comment:总积分" json:"total_points"`          // 历史累计总积分。
	AvailablePoints int64       `gorm:"not null;default:0;comment:可用积分" json:"available_points"`     // 当前可用的积分余额。
	FrozenPoints    int64       `gorm:"not null;default:0;comment:冻结积分" json:"frozen_points"`        // 因特定操作（如订单待支付）而冻结的积分。
	TotalSpent      uint64      `gorm:"not null;default:0;comment:总消费金额" json:"total_spent"`         // 历史累计总消费金额。
}

// NewMemberAccount 创建并返回一个新的 MemberAccount 实体实例。
// userID: 关联的用户ID。
func NewMemberAccount(userID uint64) *MemberAccount {
	return &MemberAccount{
		UserID:          userID,
		Level:           MemberLevelBronze, // 新账户默认为青铜会员。
		TotalPoints:     0,
		AvailablePoints: 0,
		FrozenPoints:    0,
		TotalSpent:      0,
	}
}

// AddPoints 增加会员账户的积分。
// points: 待增加的积分数量。
func (a *MemberAccount) AddPoints(points int64) {
	a.TotalPoints += points     // 增加总积分。
	a.AvailablePoints += points // 增加可用积分。
	a.checkLevelUpgrade()       // 检查是否满足等级升级条件。
}

// DeductPoints 扣减会员账户的积分。
// points: 待扣减的积分数量。
// 如果可用积分不足，则返回 ErrInsufficientPoints 错误。
func (a *MemberAccount) DeductPoints(points int64) error {
	if a.AvailablePoints < points {
		return ErrInsufficientPoints // 积分不足。
	}
	a.AvailablePoints -= points // 扣减可用积分。
	return nil
}

// FreezePoints 冻结会员账户的积分。
// points: 待冻结的积分数量。
// 如果可用积分不足，则返回 ErrInsufficientPoints 错误。
func (a *MemberAccount) FreezePoints(points int64) error {
	if a.AvailablePoints < points {
		return ErrInsufficientPoints // 可用积分不足以冻结。
	}
	a.AvailablePoints -= points // 从可用积分中扣除。
	a.FrozenPoints += points    // 增加冻结积分。
	return nil
}

// UnfreezePoints 解冻会员账户的积分。
// points: 待解冻的积分数量。
func (a *MemberAccount) UnfreezePoints(points int64) {
	if a.FrozenPoints >= points { // 确保有足够的冻结积分可以解冻。
		a.FrozenPoints -= points    // 减少冻结积分。
		a.AvailablePoints += points // 增加可用积分。
	}
}

// AddSpent 增加会员账户的总消费金额。
// amount: 待增加的消费金额。
func (a *MemberAccount) AddSpent(amount uint64) {
	a.TotalSpent += amount // 增加总消费金额。
	a.checkLevelUpgrade()  // 检查是否满足等级升级条件。
}

// checkLevelUpgrade 根据用户的总消费金额检查并升级会员等级。
// 这是一个简化的等级升级逻辑。
func (a *MemberAccount) checkLevelUpgrade() {
	if a.TotalSpent >= 100000 { // 例如，消费1000元升级为钻石会员。
		a.Level = MemberLevelDiamond
	} else if a.TotalSpent >= 50000 {
		a.Level = MemberLevelPlatinum
	} else if a.TotalSpent >= 20000 {
		a.Level = MemberLevelGold
	} else if a.TotalSpent >= 5000 {
		a.Level = MemberLevelSilver
	} else {
		a.Level = MemberLevelBronze
	}
}

// PointsTransaction 实体代表一次积分交易记录。
type PointsTransaction struct {
	gorm.Model                 // 嵌入gorm.Model。
	UserID          uint64     `gorm:"not null;index;comment:用户ID" json:"user_id"`                     // 关联的用户ID，索引字段。
	TransactionType string     `gorm:"type:varchar(32);not null;comment:交易类型" json:"transaction_type"` // 交易类型，例如“收入”、“支出”、“兑换”。
	Points          int64      `gorm:"not null;comment:积分变动" json:"points"`                            // 积分变动数量（正数表示增加，负数表示减少）。
	Balance         int64      `gorm:"not null;comment:变动后余额" json:"balance"`                          // 变动后的积分余额。
	OrderID         uint64     `gorm:"index;comment:关联订单ID" json:"order_id"`                           // 关联的订单ID（如果交易与订单相关），索引字段。
	Description     string     `gorm:"type:varchar(255);comment:描述" json:"description"`                // 交易描述。
	ExpireAt        *time.Time `gorm:"comment:过期时间" json:"expire_at"`                                  // 积分的过期时间。
}

// NewPointsTransaction 创建并返回一个新的 PointsTransaction 实体实例。
// userID: 用户ID。
// transactionType: 交易类型。
// points: 积分变动。
// balance: 变动后余额。
// orderID: 关联订单ID。
// description: 描述。
// expireAt: 过期时间。
func NewPointsTransaction(userID uint64, transactionType string, points, balance int64, orderID uint64, description string, expireAt *time.Time) *PointsTransaction {
	return &PointsTransaction{
		UserID:          userID,
		TransactionType: transactionType,
		Points:          points,
		Balance:         balance,
		OrderID:         orderID,
		Description:     description,
		ExpireAt:        expireAt,
	}
}

// IsExpired 检查积分交易记录是否已过期。
func (t *PointsTransaction) IsExpired() bool {
	if t.ExpireAt == nil {
		return false // 如果没有设置过期时间，则永不过期。
	}
	return time.Now().After(*t.ExpireAt) // 检查当前时间是否晚于过期时间。
}

// MemberBenefit 实体代表会员等级所享有的权益。
type MemberBenefit struct {
	gorm.Model               // 嵌入gorm.Model。
	Level        MemberLevel `gorm:"type:varchar(32);not null;index;comment:会员等级" json:"level"`       // 关联的会员等级，索引字段。
	Name         string      `gorm:"type:varchar(64);not null;comment:权益名称" json:"name"`              // 权益名称。
	Description  string      `gorm:"type:text;comment:权益描述" json:"description"`                       // 权益描述。
	DiscountRate float64     `gorm:"type:decimal(5,2);default:1.00;comment:折扣率" json:"discount_rate"` // 折扣率，例如0.9表示9折。
	PointsRate   float64     `gorm:"type:decimal(5,2);default:1.00;comment:积分倍率" json:"points_rate"`  // 积分倍率，例如1.5表示1.5倍积分。
	Enabled      bool        `gorm:"default:true;comment:是否启用" json:"enabled"`                        // 权益是否启用，默认为启用。
}

// NewMemberBenefit 创建并返回一个新的 MemberBenefit 实体实例。
// level: 会员等级。
// name: 权益名称。
// description: 权益描述。
// discountRate: 折扣率。
// pointsRate: 积分倍率。
func NewMemberBenefit(level MemberLevel, name, description string, discountRate, pointsRate float64) *MemberBenefit {
	return &MemberBenefit{
		Level:        level,
		Name:         name,
		Description:  description,
		DiscountRate: discountRate,
		PointsRate:   pointsRate,
		Enabled:      true, // 默认启用。
	}
}

// Enable 启用会员权益。
func (b *MemberBenefit) Enable() {
	b.Enabled = true
}

// Disable 禁用会员权益。
func (b *MemberBenefit) Disable() {
	b.Enabled = false
}

// UpdateRates 更新会员权益的折扣率和积分倍率。
func (b *MemberBenefit) UpdateRates(discountRate, pointsRate float64) {
	b.DiscountRate = discountRate
	b.PointsRate = pointsRate
}
