package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"gorm.io/gorm"
)

// CampaignModel 营销活动写模型。
type CampaignModel struct {
	gorm.Model
	Name         string                `gorm:"type:varchar(128);not null;comment:活动名称"`
	CampaignType domain.CampaignType   `gorm:"type:varchar(32);not null;comment:活动类型"`
	Description  string                `gorm:"type:text;comment:活动描述"`
	StartTime    time.Time             `gorm:"not null;comment:开始时间"`
	EndTime      time.Time             `gorm:"not null;comment:结束时间"`
	Budget       uint64                `gorm:"not null;default:0;comment:预算"`
	Spent        uint64                `gorm:"not null;default:0;comment:已花费"`
	TargetUsers  int64                 `gorm:"not null;default:0;comment:目标用户数"`
	ReachedUsers int64                 `gorm:"not null;default:0;comment:触达用户数"`
	Status       domain.CampaignStatus `gorm:"default:0;comment:状态"`
	Rules        domain.JSONMap        `gorm:"type:json;comment:规则配置"`
}

// CampaignParticipationModel 参与记录写模型。
type CampaignParticipationModel struct {
	gorm.Model
	CampaignID uint64 `gorm:"not null;index;comment:活动ID"`
	UserID     uint64 `gorm:"not null;index;comment:用户ID"`
	OrderID    uint64 `gorm:"index;comment:订单ID"`
	Discount   uint64 `gorm:"not null;default:0;comment:优惠金额"`
}

// BannerModel 广告位写模型。
type BannerModel struct {
	gorm.Model
	Title      string    `gorm:"type:varchar(128);not null;comment:标题"`
	ImageURL   string    `gorm:"type:varchar(255);not null;comment:图片URL"`
	LinkURL    string    `gorm:"type:varchar(255);comment:跳转URL"`
	Position   string    `gorm:"type:varchar(32);not null;comment:位置"`
	Priority   int32     `gorm:"default:0;comment:优先级"`
	StartTime  time.Time `gorm:"not null;comment:开始时间"`
	EndTime    time.Time `gorm:"not null;comment:结束时间"`
	ClickCount int64     `gorm:"default:0;comment:点击数"`
	Enabled    bool      `gorm:"default:true;comment:是否启用"`
}

func (CampaignModel) TableName() string {
	return "campaigns"
}

func (CampaignParticipationModel) TableName() string {
	return "campaign_participations"
}

func (BannerModel) TableName() string {
	return "banners"
}

func toCampaignModel(c *domain.Campaign) *CampaignModel {
	if c == nil {
		return nil
	}
	return &CampaignModel{
		Model: gorm.Model{
			ID:        uint(c.ID),
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		Name:         c.Name,
		CampaignType: c.CampaignType,
		Description:  c.Description,
		StartTime:    c.StartTime,
		EndTime:      c.EndTime,
		Budget:       c.Budget,
		Spent:        c.Spent,
		TargetUsers:  c.TargetUsers,
		ReachedUsers: c.ReachedUsers,
		Status:       c.Status,
		Rules:        c.Rules,
	}
}

func toCampaign(model *CampaignModel) *domain.Campaign {
	if model == nil {
		return nil
	}
	return &domain.Campaign{
		ID:           uint64(model.ID),
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		Name:         model.Name,
		CampaignType: model.CampaignType,
		Description:  model.Description,
		StartTime:    model.StartTime,
		EndTime:      model.EndTime,
		Budget:       model.Budget,
		Spent:        model.Spent,
		TargetUsers:  model.TargetUsers,
		ReachedUsers: model.ReachedUsers,
		Status:       model.Status,
		Rules:        model.Rules,
	}
}

func toParticipationModel(p *domain.CampaignParticipation) *CampaignParticipationModel {
	if p == nil {
		return nil
	}
	return &CampaignParticipationModel{
		Model: gorm.Model{
			ID:        uint(p.ID),
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		CampaignID: p.CampaignID,
		UserID:     p.UserID,
		OrderID:    p.OrderID,
		Discount:   p.Discount,
	}
}

func toParticipation(model *CampaignParticipationModel) *domain.CampaignParticipation {
	if model == nil {
		return nil
	}
	return &domain.CampaignParticipation{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		CampaignID: model.CampaignID,
		UserID:     model.UserID,
		OrderID:    model.OrderID,
		Discount:   model.Discount,
	}
}

func toBannerModel(b *domain.Banner) *BannerModel {
	if b == nil {
		return nil
	}
	return &BannerModel{
		Model: gorm.Model{
			ID:        uint(b.ID),
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		},
		Title:      b.Title,
		ImageURL:   b.ImageURL,
		LinkURL:    b.LinkURL,
		Position:   b.Position,
		Priority:   b.Priority,
		StartTime:  b.StartTime,
		EndTime:    b.EndTime,
		ClickCount: b.ClickCount,
		Enabled:    b.Enabled,
	}
}

func toBanner(model *BannerModel) *domain.Banner {
	if model == nil {
		return nil
	}
	return &domain.Banner{
		ID:         uint64(model.ID),
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		Title:      model.Title,
		ImageURL:   model.ImageURL,
		LinkURL:    model.LinkURL,
		Position:   model.Position,
		Priority:   model.Priority,
		StartTime:  model.StartTime,
		EndTime:    model.EndTime,
		ClickCount: model.ClickCount,
		Enabled:    model.Enabled,
	}
}
