package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/domain"
	"gorm.io/gorm"
)

type RiskAssessmentModel struct {
	gorm.Model
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	TargetID     string `gorm:"column:target_id;type:varchar(64);index;not null"`
	TargetType   string `gorm:"column:target_type;type:varchar(32);not null"`
	UserID       uint64 `gorm:"column:user_id;index;not null"`
	TotalScore   int    `gorm:"column:total_score;not null"`
	Level        string `gorm:"column:level;type:varchar(32);not null"`
	MatchedRules string `gorm:"column:matched_rules;type:text"`
	Decision     string `gorm:"column:decision;type:varchar(32);not null"`
	DeviceFinger string `gorm:"column:device_finger;type:varchar(128)"`
	IPAddress    string `gorm:"column:ip_address;type:varchar(64)"`
	ReviewerID   string `gorm:"column:reviewer_id;type:varchar(64)"`
	ReviewedAt   *time.Time `gorm:"column:reviewed_at"`
	ReviewNotes  string `gorm:"column:review_notes;type:text"`
}

func (RiskAssessmentModel) TableName() string {
	return "risk_assessments"
}

type RiskRuleModel struct {
	gorm.Model
	ID         uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Name       string `gorm:"column:name;type:varchar(128);not null"`
	Type       int    `gorm:"column:type;not null"`
	Config     string `gorm:"column:config;type:text"`
	Weight     int    `gorm:"column:weight;not null"`
	IsBlocking bool   `gorm:"column:is_blocking;default:false"`
	IsActive   bool   `gorm:"column:is_active;default:true"`
}

func (RiskRuleModel) TableName() string {
	return "risk_rules"
}

type BlacklistModel struct {
	gorm.Model
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	Type      int        `gorm:"column:type;not null;index:idx_type_value,unique"`
	Value     string     `gorm:"column:value;type:varchar(128);not null;index:idx_type_value,unique"`
	Reason    string     `gorm:"column:reason;type:varchar(512)"`
	Source    string     `gorm:"column:source;type:varchar(32)"`
	ExpiresAt *time.Time `gorm:"column:expires_at;index"`
}

func (BlacklistModel) TableName() string {
	return "risk_blacklist"
}

type CreditProfileModel struct {
	gorm.Model
	ID             uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	UserID         uint64 `gorm:"column:user_id;uniqueIndex;not null"`
	Score          int    `gorm:"column:score;not null"`
	Level          string `gorm:"column:level;type:varchar(32);not null"`
	TotalLimit     int64  `gorm:"column:total_limit;not null"`
	UsedLimit      int64  `gorm:"column:used_limit;not null;default:0"`
	AvailableLimit int64  `gorm:"column:available_limit;not null"`
	Factors        string `gorm:"column:factors;type:text"`
	Status         string `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE'"`
	LastAssessedAt time.Time `gorm:"column:last_assessed_at"`
}

func (CreditProfileModel) TableName() string {
	return "credit_profiles"
}

type CreditRecordModel struct {
	gorm.Model
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	ProfileID   uint64 `gorm:"column:profile_id;index;not null"`
	Type        int    `gorm:"column:type;not null"`
	ScoreChange int    `gorm:"column:score_change;not null"`
	LimitChange int64  `gorm:"column:limit_change;not null"`
	RelatedID   string `gorm:"column:related_id;type:varchar(64)"`
	Reason      string `gorm:"column:reason;type:varchar(512)"`
	HappenedAt  time.Time `gorm:"column:happened_at;not null"`
}

func (CreditRecordModel) TableName() string {
	return "credit_records"
}

type RiskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) domain.RiskRepository {
	return &RiskRepository{db: db}
}

func (r *RiskRepository) SaveAssessment(ctx context.Context, assessment *domain.RiskAssessment) error {
	model := toRiskAssessmentModel(assessment)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *RiskRepository) FindAssessment(ctx context.Context, targetID string) (*domain.RiskAssessment, error) {
	var model RiskAssessmentModel
	if err := r.db.WithContext(ctx).Where("target_id = ?", targetID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toRiskAssessmentDomain(&model), nil
}

func (r *RiskRepository) FindActiveRules(ctx context.Context, targetType string) ([]*domain.RiskRule, error) {
	var models []RiskRuleModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&models).Error; err != nil {
		return nil, err
	}
	rules := make([]*domain.RiskRule, len(models))
	for i, m := range models {
		rules[i] = toRiskRuleDomain(&m)
	}
	return rules, nil
}

func (r *RiskRepository) IsInBlacklist(ctx context.Context, blacklistType domain.BlacklistType, value string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&BlacklistModel{}).
		Where("type = ? AND value = ?", int(blacklistType), value).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *RiskRepository) AddToBlacklist(ctx context.Context, item *domain.Blacklist) error {
	model := toBlacklistModel(item)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *RiskRepository) RemoveFromBlacklist(ctx context.Context, blacklistType domain.BlacklistType, value string) error {
	return r.db.WithContext(ctx).
		Where("type = ? AND value = ?", int(blacklistType), value).
		Delete(&BlacklistModel{}).Error
}

type CreditProfileRepository struct {
	db *gorm.DB
}

func NewCreditProfileRepository(db *gorm.DB) domain.CreditProfileRepository {
	return &CreditProfileRepository{db: db}
}

func (r *CreditProfileRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.CreditProfile, error) {
	var model CreditProfileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCreditProfileDomain(&model), nil
}

func (r *CreditProfileRepository) Save(ctx context.Context, profile *domain.CreditProfile) error {
	model := toCreditProfileModel(profile)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *CreditProfileRepository) AddRecord(ctx context.Context, record *domain.CreditRecord) error {
	model := toCreditRecordModel(record)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *CreditProfileRepository) FindRecords(ctx context.Context, profileID uint64, limit int) ([]*domain.CreditRecord, error) {
	var models []CreditRecordModel
	if err := r.db.WithContext(ctx).
		Where("profile_id = ?", profileID).
		Order("happened_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	records := make([]*domain.CreditRecord, len(models))
	for i, m := range models {
		records[i] = toCreditRecordDomain(&m)
	}
	return records, nil
}

func toRiskAssessmentModel(a *domain.RiskAssessment) *RiskAssessmentModel {
	matchedRulesJSON, _ := json.Marshal(a.MatchedRules)
	return &RiskAssessmentModel{
		ID:           a.ID,
		TargetID:     a.TargetID,
		TargetType:   a.TargetType,
		UserID:       a.UserID,
		TotalScore:   a.TotalScore,
		Level:        string(a.Level),
		MatchedRules: string(matchedRulesJSON),
		Decision:     a.Decision,
		DeviceFinger: a.DeviceFinger,
		IPAddress:    a.IPAddress,
		ReviewerID:   a.ReviewerID,
		ReviewedAt:   a.ReviewedAt,
		ReviewNotes:  a.ReviewNotes,
	}
}

func toRiskAssessmentDomain(m *RiskAssessmentModel) *domain.RiskAssessment {
	var matchedRules []*domain.MatchedRule
	_ = json.Unmarshal([]byte(m.MatchedRules), &matchedRules)
	return &domain.RiskAssessment{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt,
		TargetID:     m.TargetID,
		TargetType:   m.TargetType,
		UserID:       m.UserID,
		TotalScore:   m.TotalScore,
		Level:        domain.RiskLevel(m.Level),
		MatchedRules: matchedRules,
		Decision:     m.Decision,
		DeviceFinger: m.DeviceFinger,
		IPAddress:    m.IPAddress,
		ReviewerID:   m.ReviewerID,
		ReviewedAt:   m.ReviewedAt,
		ReviewNotes:  m.ReviewNotes,
	}
}

func toRiskRuleDomain(m *RiskRuleModel) *domain.RiskRule {
	return &domain.RiskRule{
		ID:         m.ID,
		Name:       m.Name,
		Type:       domain.RiskRuleType(m.Type),
		Config:     m.Config,
		Weight:     m.Weight,
		IsBlocking: m.IsBlocking,
		IsActive:   m.IsActive,
	}
}

func toBlacklistModel(b *domain.Blacklist) *BlacklistModel {
	return &BlacklistModel{
		Type:      int(b.Type),
		Value:     b.Value,
		Reason:    b.Reason,
		Source:    b.Source,
		ExpiresAt: b.ExpiresAt,
	}
}

func toCreditProfileModel(p *domain.CreditProfile) *CreditProfileModel {
	factorsJSON, _ := json.Marshal(p.Factors)
	return &CreditProfileModel{
		ID:             p.ID,
		UserID:         p.UserID,
		Score:          p.Score,
		Level:          string(p.Level),
		TotalLimit:     p.TotalLimit,
		UsedLimit:      p.UsedLimit,
		AvailableLimit: p.AvailableLimit,
		Factors:        string(factorsJSON),
		Status:         p.Status,
		LastAssessedAt: p.LastAssessedAt,
	}
}

func toCreditProfileDomain(m *CreditProfileModel) *domain.CreditProfile {
	var factors []*domain.FactorScore
	_ = json.Unmarshal([]byte(m.Factors), &factors)
	return &domain.CreditProfile{
		ID:             m.ID,
		UserID:         m.UserID,
		Score:          m.Score,
		Level:          domain.CreditLevel(m.Level),
		TotalLimit:     m.TotalLimit,
		UsedLimit:      m.UsedLimit,
		AvailableLimit: m.AvailableLimit,
		Factors:        factors,
		Status:         m.Status,
		LastAssessedAt: m.LastAssessedAt,
	}
}

func toCreditRecordModel(r *domain.CreditRecord) *CreditRecordModel {
	return &CreditRecordModel{
		ProfileID:   r.ProfileID,
		Type:        int(r.Type),
		ScoreChange: r.ScoreChange,
		LimitChange: r.LimitChange,
		RelatedID:   r.RelatedID,
		Reason:      r.Reason,
		HappenedAt:  r.HappenedAt,
	}
}

func toCreditRecordDomain(m *CreditRecordModel) *domain.CreditRecord {
	return &domain.CreditRecord{
		ID:          m.ID,
		ProfileID:   m.ProfileID,
		Type:        domain.CreditRecordType(m.Type),
		ScoreChange: m.ScoreChange,
		LimitChange: m.LimitChange,
		RelatedID:   m.RelatedID,
		Reason:      m.Reason,
		HappenedAt:  m.HappenedAt,
	}
}

var _ domain.RiskRepository = (*RiskRepository)(nil)
var _ domain.CreditProfileRepository = (*CreditProfileRepository)(nil)
