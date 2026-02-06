package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"gorm.io/gorm"
)

// CategoryMultipliers JSON 序列化的类目倍率。
type CategoryMultipliers map[string]float64

func (m CategoryMultipliers) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *CategoryMultipliers) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// MemberAccountModel 会员账户写模型。
type MemberAccountModel struct {
	gorm.Model
	UserID          uint64             `gorm:"not null;uniqueIndex;comment:用户ID"`
	Level           domain.MemberLevel `gorm:"type:varchar(32);default:'BRONZE';comment:会员等级"`
	TotalPoints     int64              `gorm:"not null;default:0;comment:总积分"`
	AvailablePoints int64              `gorm:"not null;default:0;comment:可用积分"`
	FrozenPoints    int64              `gorm:"not null;default:0;comment:冻结积分"`
	TotalSpent      uint64             `gorm:"not null;default:0;comment:总消费金额"`
}

// PointsTransactionModel 积分交易写模型。
type PointsTransactionModel struct {
	gorm.Model
	UserID          uint64     `gorm:"not null;index;comment:用户ID"`
	TransactionType string     `gorm:"type:varchar(32);not null;comment:交易类型"`
	Points          int64      `gorm:"not null;comment:积分变动"`
	Balance         int64      `gorm:"not null;comment:变动后余额"`
	OrderID         uint64     `gorm:"index;comment:关联订单ID"`
	Description     string     `gorm:"type:varchar(255);comment:描述"`
	ExpireAt        *time.Time `gorm:"comment:过期时间"`
}

// MemberBenefitModel 会员权益写模型。
type MemberBenefitModel struct {
	gorm.Model
	Level        domain.MemberLevel  `gorm:"type:varchar(32);not null;index;comment:会员等级"`
	Name         string              `gorm:"type:varchar(64);not null;comment:权益名称"`
	Description  string              `gorm:"type:text;comment:权益描述"`
	DiscountRate float64             `gorm:"type:decimal(5,2);default:1.00;comment:折扣率"`
	PointsRate   float64             `gorm:"type:decimal(5,2);default:1.00;comment:积分倍率"`
	Multipliers  CategoryMultipliers `gorm:"type:json;comment:类目特定倍率"`
	Enabled      bool                `gorm:"default:true;comment:是否启用"`
}

func (MemberAccountModel) TableName() string     { return "member_accounts" }
func (PointsTransactionModel) TableName() string { return "points_transactions" }
func (MemberBenefitModel) TableName() string     { return "member_benefits" }

func toAccountModel(account *domain.MemberAccount) *MemberAccountModel {
	if account == nil {
		return nil
	}
	return &MemberAccountModel{
		Model: gorm.Model{
			ID:        uint(account.ID),
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		UserID:          account.UserID,
		Level:           account.Level,
		TotalPoints:     account.TotalPoints,
		AvailablePoints: account.AvailablePoints,
		FrozenPoints:    account.FrozenPoints,
		TotalSpent:      account.TotalSpent,
	}
}

func toAccount(model *MemberAccountModel) *domain.MemberAccount {
	if model == nil {
		return nil
	}
	return &domain.MemberAccount{
		ID:              uint64(model.ID),
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		UserID:          model.UserID,
		Level:           model.Level,
		TotalPoints:     model.TotalPoints,
		AvailablePoints: model.AvailablePoints,
		FrozenPoints:    model.FrozenPoints,
		TotalSpent:      model.TotalSpent,
	}
}

func toTransactionModel(tx *domain.PointsTransaction) *PointsTransactionModel {
	if tx == nil {
		return nil
	}
	return &PointsTransactionModel{
		Model: gorm.Model{
			ID:        uint(tx.ID),
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
		},
		UserID:          tx.UserID,
		TransactionType: tx.TransactionType,
		Points:          tx.Points,
		Balance:         tx.Balance,
		OrderID:         tx.OrderID,
		Description:     tx.Description,
		ExpireAt:        tx.ExpireAt,
	}
}

func toTransaction(model *PointsTransactionModel) *domain.PointsTransaction {
	if model == nil {
		return nil
	}
	return &domain.PointsTransaction{
		ID:              uint64(model.ID),
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		UserID:          model.UserID,
		TransactionType: model.TransactionType,
		Points:          model.Points,
		Balance:         model.Balance,
		OrderID:         model.OrderID,
		Description:     model.Description,
		ExpireAt:        model.ExpireAt,
	}
}

func toBenefitModel(benefit *domain.MemberBenefit) *MemberBenefitModel {
	if benefit == nil {
		return nil
	}
	return &MemberBenefitModel{
		Model: gorm.Model{
			ID:        uint(benefit.ID),
			CreatedAt: benefit.CreatedAt,
			UpdatedAt: benefit.UpdatedAt,
		},
		Level:        benefit.Level,
		Name:         benefit.Name,
		Description:  benefit.Description,
		DiscountRate: benefit.DiscountRate,
		PointsRate:   benefit.PointsRate,
		Multipliers:  CategoryMultipliers(benefit.Multipliers),
		Enabled:      benefit.Enabled,
	}
}

func toBenefit(model *MemberBenefitModel) *domain.MemberBenefit {
	if model == nil {
		return nil
	}
	return &domain.MemberBenefit{
		ID:           uint64(model.ID),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		Level:        model.Level,
		Name:         model.Name,
		Description:  model.Description,
		DiscountRate: model.DiscountRate,
		PointsRate:   model.PointsRate,
		Multipliers:  domain.CategoryMultipliers(model.Multipliers),
		Enabled:      model.Enabled,
	}
}
