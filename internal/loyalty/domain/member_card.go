// 生成摘要：
// - 从 membercard 服务合并到 loyalty 域。
// - 会员卡/储值卡属于忠诚度域子聚合，与积分体系互补。
// - 关键实体：MemberCard（储值卡聚合根）。
// - 并发控制策略：乐观锁 (Version) + 余额扣减原子操作。
package domain

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CardStatus 储值卡/礼品卡状态。
type CardStatus string

const (
	// CardActive 卡已激活。
	CardActive CardStatus = "ACTIVE"
	// CardInactive 卡未激活。
	CardInactive CardStatus = "INACTIVE"
	// CardFrozen 卡已冻结。
	CardFrozen CardStatus = "FROZEN"
	// CardExpired 卡已过期。
	CardExpired CardStatus = "EXPIRED"
)

// MemberCard 礼品卡/储值卡聚合根，防超发。
type MemberCard struct {
	gorm.Model
	// CardNo 卡号，全局唯一。
	CardNo string `gorm:"type:varchar(64);uniqueIndex" json:"card_no"`
	// Password 卡密哈希。
	Password string `gorm:"type:varchar(128);comment:卡密哈希" json:"-"`
	// UserID 绑定者 ID。
	UserID uint64 `gorm:"index;comment:绑定者ID" json:"user_id"`
	// Balance 当前余额。
	Balance decimal.Decimal `gorm:"type:decimal(20,4)" json:"balance"`
	// FaceValue 面值。
	FaceValue decimal.Decimal `gorm:"type:decimal(20,4);comment:面值" json:"face_value"`
	// Status 卡状态。
	Status CardStatus `gorm:"type:varchar(16)" json:"status"`
	// Version 乐观锁版本号，用于余额扣减并发控制。
	Version int64 `gorm:"default:0" json:"version"`
}

// Deduct 从储值卡扣款。
// 校验卡状态和余额是否充足。
func (c *MemberCard) Deduct(amount decimal.Decimal) error {
	if c.Status != CardActive {
		return errors.New("card is not active")
	}
	if c.Balance.LessThan(amount) {
		return errors.New("insufficient card balance")
	}
	c.Balance = c.Balance.Sub(amount)
	return nil
}
