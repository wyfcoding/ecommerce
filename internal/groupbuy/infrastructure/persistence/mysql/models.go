package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
	"gorm.io/gorm"
)

// GroupbuyModel 拼团活动写模型。
type GroupbuyModel struct {
	gorm.Model
	Name          string                `gorm:"column:name;type:varchar(255);not null;comment:活动名称"`
	ProductID     uint64                `gorm:"column:product_id;not null;comment:商品ID"`
	SkuID         uint64                `gorm:"column:sku_id;not null;comment:SKU ID"`
	OriginalPrice uint64                `gorm:"column:original_price;not null;comment:原价(分)"`
	GroupPrice    uint64                `gorm:"column:group_price;not null;comment:拼团价(分)"`
	MinPeople     int32                 `gorm:"column:min_people;not null;comment:最小成团人数"`
	MaxPeople     int32                 `gorm:"column:max_people;not null;comment:最大成团人数"`
	TotalStock    int32                 `gorm:"column:total_stock;not null;comment:总库存"`
	SoldCount     int32                 `gorm:"column:sold_count;not null;default:0;comment:已售数量"`
	StartTime     time.Time             `gorm:"column:start_time;not null;comment:开始时间"`
	EndTime       time.Time             `gorm:"column:end_time;not null;comment:结束时间"`
	Status        domain.GroupbuyStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
	Description   string                `gorm:"column:description;type:text;comment:活动描述"`
}

// GroupbuyTeamModel 拼团团队写模型。
type GroupbuyTeamModel struct {
	gorm.Model
	GroupbuyID    uint64                    `gorm:"column:groupbuy_id;not null;index;comment:拼团活动ID"`
	TeamNo        string                    `gorm:"column:team_no;type:varchar(64);uniqueIndex;not null;comment:拼团编号"`
	LeaderID      uint64                    `gorm:"column:leader_id;not null;comment:团长用户ID"`
	CurrentPeople int32                     `gorm:"column:current_people;not null;default:1;comment:当前人数"`
	MaxPeople     int32                     `gorm:"column:max_people;not null;comment:最大人数"`
	Status        domain.GroupbuyTeamStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
	ExpireAt      time.Time                 `gorm:"column:expire_at;not null;comment:过期时间"`
	SuccessAt     *time.Time                `gorm:"column:success_at;comment:成团时间"`
}

// GroupbuyOrderModel 拼团订单写模型。
type GroupbuyOrderModel struct {
	gorm.Model
	GroupbuyID  uint64                     `gorm:"column:groupbuy_id;not null;index;comment:拼团活动ID"`
	TeamID      uint64                     `gorm:"column:team_id;not null;index;comment:拼团团队ID"`
	TeamNo      string                     `gorm:"column:team_no;type:varchar(64);not null;comment:拼团编号"`
	UserID      uint64                     `gorm:"column:user_id;not null;index;comment:用户ID"`
	ProductID   uint64                     `gorm:"column:product_id;not null;comment:商品ID"`
	SkuID       uint64                     `gorm:"column:sku_id;not null;comment:SKU ID"`
	Price       uint64                     `gorm:"column:price;not null;comment:单价(分)"`
	Quantity    int32                      `gorm:"column:quantity;not null;comment:数量"`
	TotalAmount uint64                     `gorm:"column:total_amount;not null;comment:总金额(分)"`
	IsLeader    bool                       `gorm:"column:is_leader;not null;default:false;comment:是否团长"`
	Status      domain.GroupbuyOrderStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:状态"`
	PaidAt      *time.Time                 `gorm:"column:paid_at;comment:支付时间"`
	RefundedAt  *time.Time                 `gorm:"column:refunded_at;comment:退款时间"`
}

func (GroupbuyModel) TableName() string      { return "groupbuys" }
func (GroupbuyTeamModel) TableName() string  { return "groupbuy_teams" }
func (GroupbuyOrderModel) TableName() string { return "groupbuy_orders" }

func toGroupbuyModel(groupbuy *domain.Groupbuy) *GroupbuyModel {
	if groupbuy == nil {
		return nil
	}
	return &GroupbuyModel{
		Model: gorm.Model{
			ID:        groupbuy.ID,
			CreatedAt: groupbuy.CreatedAt,
			UpdatedAt: groupbuy.UpdatedAt,
		},
		Name:          groupbuy.Name,
		ProductID:     groupbuy.ProductID,
		SkuID:         groupbuy.SkuID,
		OriginalPrice: groupbuy.OriginalPrice,
		GroupPrice:    groupbuy.GroupPrice,
		MinPeople:     groupbuy.MinPeople,
		MaxPeople:     groupbuy.MaxPeople,
		TotalStock:    groupbuy.TotalStock,
		SoldCount:     groupbuy.SoldCount,
		StartTime:     groupbuy.StartTime,
		EndTime:       groupbuy.EndTime,
		Status:        groupbuy.Status,
		Description:   groupbuy.Description,
	}
}

func toGroupbuy(model *GroupbuyModel) *domain.Groupbuy {
	if model == nil {
		return nil
	}
	return &domain.Groupbuy{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Name:          model.Name,
		ProductID:     model.ProductID,
		SkuID:         model.SkuID,
		OriginalPrice: model.OriginalPrice,
		GroupPrice:    model.GroupPrice,
		MinPeople:     model.MinPeople,
		MaxPeople:     model.MaxPeople,
		TotalStock:    model.TotalStock,
		SoldCount:     model.SoldCount,
		StartTime:     model.StartTime,
		EndTime:       model.EndTime,
		Status:        model.Status,
		Description:   model.Description,
	}
}

func toGroupbuyTeamModel(team *domain.GroupbuyTeam) *GroupbuyTeamModel {
	if team == nil {
		return nil
	}
	return &GroupbuyTeamModel{
		Model: gorm.Model{
			ID:        team.ID,
			CreatedAt: team.CreatedAt,
			UpdatedAt: team.UpdatedAt,
		},
		GroupbuyID:    team.GroupbuyID,
		TeamNo:        team.TeamNo,
		LeaderID:      team.LeaderID,
		CurrentPeople: team.CurrentPeople,
		MaxPeople:     team.MaxPeople,
		Status:        team.Status,
		ExpireAt:      team.ExpireAt,
		SuccessAt:     team.SuccessAt,
	}
}

func toGroupbuyTeam(model *GroupbuyTeamModel) *domain.GroupbuyTeam {
	if model == nil {
		return nil
	}
	return &domain.GroupbuyTeam{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		GroupbuyID:    model.GroupbuyID,
		TeamNo:        model.TeamNo,
		LeaderID:      model.LeaderID,
		CurrentPeople: model.CurrentPeople,
		MaxPeople:     model.MaxPeople,
		Status:        model.Status,
		ExpireAt:      model.ExpireAt,
		SuccessAt:     model.SuccessAt,
	}
}

func toGroupbuyOrderModel(order *domain.GroupbuyOrder) *GroupbuyOrderModel {
	if order == nil {
		return nil
	}
	return &GroupbuyOrderModel{
		Model: gorm.Model{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
		GroupbuyID:  order.GroupbuyID,
		TeamID:      order.TeamID,
		TeamNo:      order.TeamNo,
		UserID:      order.UserID,
		ProductID:   order.ProductID,
		SkuID:       order.SkuID,
		Price:       order.Price,
		Quantity:    order.Quantity,
		TotalAmount: order.TotalAmount,
		IsLeader:    order.IsLeader,
		Status:      order.Status,
		PaidAt:      order.PaidAt,
		RefundedAt:  order.RefundedAt,
	}
}

func toGroupbuyOrder(model *GroupbuyOrderModel) *domain.GroupbuyOrder {
	if model == nil {
		return nil
	}
	return &domain.GroupbuyOrder{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		GroupbuyID:  model.GroupbuyID,
		TeamID:      model.TeamID,
		TeamNo:      model.TeamNo,
		UserID:      model.UserID,
		ProductID:   model.ProductID,
		SkuID:       model.SkuID,
		Price:       model.Price,
		Quantity:    model.Quantity,
		TotalAmount: model.TotalAmount,
		IsLeader:    model.IsLeader,
		Status:      model.Status,
		PaidAt:      model.PaidAt,
		RefundedAt:  model.RefundedAt,
	}
}
