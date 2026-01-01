package domain

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInsufficientPoints = errors.New("积分不足")
	ErrPointsExpired      = errors.New("积分已过期")
)

// MemberLevel 结构体定义。
type MemberLevel string

const (
	MemberLevelBronze   MemberLevel = "BRONZE"
	MemberLevelSilver   MemberLevel = "SILVER"
	MemberLevelGold     MemberLevel = "GOLD"
	MemberLevelPlatinum MemberLevel = "PLATINUM"
	MemberLevelDiamond  MemberLevel = "DIAMOND"
)

// MemberAccount 结构体定义。
type MemberAccount struct {
	gorm.Model
	UserID          uint64      `gorm:"not null;uniqueIndex;comment:用户ID" json:"user_id"`
	Level           MemberLevel `gorm:"type:varchar(32);default:'BRONZE';comment:会员等级" json:"level"`
	TotalPoints     int64       `gorm:"not null;default:0;comment:总积分" json:"total_points"`
	AvailablePoints int64       `gorm:"not null;default:0;comment:可用积分" json:"available_points"`
	FrozenPoints    int64       `gorm:"not null;default:0;comment:冻结积分" json:"frozen_points"`
	TotalSpent      uint64      `gorm:"not null;default:0;comment:总消费金额" json:"total_spent"`
}

// NewMemberAccount 函数。
func NewMemberAccount(userID uint64) *MemberAccount {
	return &MemberAccount{
		UserID:          userID,
		Level:           MemberLevelBronze,
		TotalPoints:     0,
		AvailablePoints: 0,
		FrozenPoints:    0,
		TotalSpent:      0,
	}
}

func (a *MemberAccount) AddPoints(points int64) {
	a.TotalPoints += points
	a.AvailablePoints += points
	a.checkLevelUpgrade()
}

func (a *MemberAccount) DeductPoints(points int64) error {
	if a.AvailablePoints < points {
		return ErrInsufficientPoints
	}
	a.AvailablePoints -= points
	return nil
}

func (a *MemberAccount) FreezePoints(points int64) error {
	if a.AvailablePoints < points {
		return ErrInsufficientPoints
	}
	a.AvailablePoints -= points
	a.FrozenPoints += points
	return nil
}

func (a *MemberAccount) UnfreezePoints(points int64) {
	if a.FrozenPoints >= points {
		a.FrozenPoints -= points
		a.AvailablePoints += points
	}
}

func (a *MemberAccount) AddSpent(amount uint64) {
	a.TotalSpent += amount
	a.checkLevelUpgrade()
}

func (a *MemberAccount) checkLevelUpgrade() {
	if a.TotalSpent >= 100000 {
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

// PointsTransaction 结构体定义。
type PointsTransaction struct {
	gorm.Model
	UserID          uint64     `gorm:"not null;index;comment:用户ID" json:"user_id"`
	TransactionType string     `gorm:"type:varchar(32);not null;comment:交易类型" json:"transaction_type"`
	Points          int64      `gorm:"not null;comment:积分变动" json:"points"`
	Balance         int64      `gorm:"not null;comment:变动后余额" json:"balance"`
	OrderID         uint64     `gorm:"index;comment:关联订单ID" json:"order_id"`
	Description     string     `gorm:"type:varchar(255);comment:描述" json:"description"`
	ExpireAt        *time.Time `gorm:"comment:过期时间" json:"expire_at"`
}

// NewPointsTransaction 函数。
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

func (t *PointsTransaction) IsExpired() bool {
	if t.ExpireAt == nil {
		return false
	}
	return time.Now().After(*t.ExpireAt)
}

// CategoryMultipliers 定义类目积分倍率。
type CategoryMultipliers map[string]float64

// MemberBenefit 结构体定义。
type MemberBenefit struct {
	gorm.Model
	Level        MemberLevel         `gorm:"type:varchar(32);not null;index;comment:会员等级" json:"level"`
	Name         string              `gorm:"type:varchar(64);not null;comment:权益名称" json:"name"`
	Description  string              `gorm:"type:text;comment:权益描述" json:"description"`
	DiscountRate float64             `gorm:"type:decimal(5,2);default:1.00;comment:折扣率" json:"discount_rate"`
	PointsRate   float64             `gorm:"type:decimal(5,2);default:1.00;comment:积分倍率" json:"points_rate"`
	Multipliers  CategoryMultipliers `gorm:"type:json;serializer:json;comment:类目特定倍率" json:"multipliers"`
	Enabled      bool                `gorm:"default:true;comment:是否启用" json:"enabled"`
}

// NewMemberBenefit 函数。
func NewMemberBenefit(level MemberLevel, name, description string, discountRate, pointsRate float64) *MemberBenefit {
	return &MemberBenefit{
		Level:        level,
		Name:         name,
		Description:  description,
		DiscountRate: discountRate,
		PointsRate:   pointsRate,
		Multipliers:  make(CategoryMultipliers),
		Enabled:      true,
	}
}

func (b *MemberBenefit) Enable() {
	b.Enabled = true
}

func (b *MemberBenefit) Disable() {
	b.Enabled = false
}

func (b *MemberBenefit) UpdateRates(discountRate, pointsRate float64) {
	b.DiscountRate = discountRate
	b.PointsRate = pointsRate
}
