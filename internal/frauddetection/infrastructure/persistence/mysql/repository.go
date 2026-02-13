package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/frauddetection/domain"
	"gorm.io/gorm"
)

type RiskAssessmentModel struct {
	ID               uint      `gorm:"primaryKey"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	TransactionID    string    `gorm:"index;size:64"`
	UserID           uint64    `gorm:"index"`
	RiskLevel        int       `gorm:"column:risk_level"`
	RiskScore        int       `gorm:"column:risk_score"`
	RecommendedAction string   `gorm:"size:32"`
	RiskFactors      string    `gorm:"type:json"`
	TriggeredRules   string    `gorm:"type:json"`
	Explanation      string    `gorm:"type:text"`
}

func (RiskAssessmentModel) TableName() string {
	return "fraud_risk_assessments"
}

type BlacklistModel struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Type      string    `gorm:"index;size:32"`
	Value     string    `gorm:"index;size:256"`
	Reason    string    `gorm:"size:512"`
	CreatedBy uint64
	ExpiresAt *time.Time
	IsActive  bool `gorm:"default:true"`
}

func (BlacklistModel) TableName() string {
	return "fraud_blacklist"
}

type RiskRuleModel struct {
	ID          uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Name        string    `gorm:"size:128"`
	Description string    `gorm:"type:text"`
	Type        string    `gorm:"size:32"`
	Condition   string    `gorm:"type:text"`
	RiskWeight  int
	Action      string `gorm:"size:32"`
	Enabled     bool   `gorm:"default:true"`
	Priority    int    `gorm:"default:100"`
}

func (RiskRuleModel) TableName() string {
	return "fraud_risk_rules"
}

type DeviceFingerprintModel struct {
	ID                uint      `gorm:"primaryKey"`
	Fingerprint       string    `gorm:"uniqueIndex;size:128"`
	FirstSeenUserID   uint64
	AssociatedAccounts string  `gorm:"type:json"`
	AssociatedIPs     string   `gorm:"type:json"`
	IsSuspicious      bool     `gorm:"default:false"`
	DeviceType        string   `gorm:"size:64"`
	OS                string   `gorm:"size:64"`
	Browser           string   `gorm:"size:64"`
	FirstSeenAt       time.Time
	LastSeenAt        time.Time
}

func (DeviceFingerprintModel) TableName() string {
	return "fraud_device_fingerprints"
}

type UserRiskProfileModel struct {
	ID                  uint       `gorm:"primaryKey"`
	UserID              uint64     `gorm:"uniqueIndex"`
	RiskScore           int
	RiskLevel           int
	TotalTransactions   int
	FlaggedTransactions int
	BlockedTransactions int
	RiskTags            string     `gorm:"type:json"`
	FirstTransactionAt  *time.Time
	LastTransactionAt   *time.Time
	LastAssessmentAt    *time.Time
}

func (UserRiskProfileModel) TableName() string {
	return "fraud_user_risk_profiles"
}

type SuspiciousActivityModel struct {
	ID                uint      `gorm:"primaryKey"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UserID            uint64    `gorm:"index"`
	ActivityType      string    `gorm:"size:64"`
	Description       string    `gorm:"type:text"`
	IPAddress         string    `gorm:"size:64"`
	DeviceFingerprint string    `gorm:"size:128"`
	Metadata          string    `gorm:"type:json"`
}

func (SuspiciousActivityModel) TableName() string {
	return "fraud_suspicious_activities"
}

type riskAssessmentRepository struct {
	db *gorm.DB
}

func NewRiskAssessmentRepository(db *gorm.DB) domain.RiskAssessmentRepository {
	return &riskAssessmentRepository{db: db}
}

func (r *riskAssessmentRepository) Save(ctx context.Context, assessment *domain.RiskAssessment) error {
	model := toRiskAssessmentModel(assessment)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *riskAssessmentRepository) FindByID(ctx context.Context, id string) (*domain.RiskAssessment, error) {
	var model RiskAssessmentModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRiskAssessmentNotFound
		}
		return nil, err
	}
	return toDomainRiskAssessment(&model), nil
}

func (r *riskAssessmentRepository) FindByTransactionID(ctx context.Context, transactionID string) (*domain.RiskAssessment, error) {
	var model RiskAssessmentModel
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRiskAssessmentNotFound
		}
		return nil, err
	}
	return toDomainRiskAssessment(&model), nil
}

func (r *riskAssessmentRepository) FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*domain.RiskAssessment, error) {
	var models []RiskAssessmentModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.RiskAssessment, len(models))
	for i, m := range models {
		result[i] = toDomainRiskAssessment(&m)
	}
	return result, nil
}

type blacklistRepository struct {
	db *gorm.DB
}

func NewBlacklistRepository(db *gorm.DB) domain.BlacklistRepository {
	return &blacklistRepository{db: db}
}

func (r *blacklistRepository) Save(ctx context.Context, entry *domain.BlacklistEntry) error {
	model := toBlacklistModel(entry)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *blacklistRepository) FindByID(ctx context.Context, id string) (*domain.BlacklistEntry, error) {
	var model BlacklistModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBlacklistEntryNotFound
		}
		return nil, err
	}
	return toDomainBlacklistEntry(&model), nil
}

func (r *blacklistRepository) FindByTypeAndValue(ctx context.Context, blacklistType domain.BlacklistType, value string) (*domain.BlacklistEntry, error) {
	var model BlacklistModel
	if err := r.db.WithContext(ctx).Where("type = ? AND value = ?", blacklistType, value).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBlacklistEntryNotFound
		}
		return nil, err
	}
	return toDomainBlacklistEntry(&model), nil
}

func (r *blacklistRepository) FindByType(ctx context.Context, blacklistType domain.BlacklistType, limit, offset int) ([]*domain.BlacklistEntry, int64, error) {
	var models []BlacklistModel
	var total int64

	query := r.db.WithContext(ctx).Model(&BlacklistModel{}).Where("is_active = ?", true)
	if blacklistType != "" {
		query = query.Where("type = ?", blacklistType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.BlacklistEntry, len(models))
	for i, m := range models {
		result[i] = toDomainBlacklistEntry(&m)
	}
	return result, total, nil
}

func (r *blacklistRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&BlacklistModel{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *blacklistRepository) Exists(ctx context.Context, blacklistType domain.BlacklistType, value string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&BlacklistModel{}).
		Where("type = ? AND value = ? AND is_active = ?", blacklistType, value, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count > 0, err
}

type riskRuleRepository struct {
	db *gorm.DB
}

func NewRiskRuleRepository(db *gorm.DB) domain.RiskRuleRepository {
	return &riskRuleRepository{db: db}
}

func (r *riskRuleRepository) Save(ctx context.Context, rule *domain.RiskRule) error {
	model := toRiskRuleModel(rule)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *riskRuleRepository) FindByID(ctx context.Context, id string) (*domain.RiskRule, error) {
	var model RiskRuleModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRiskRuleNotFound
		}
		return nil, err
	}
	return toDomainRiskRule(&model), nil
}

func (r *riskRuleRepository) FindAll(ctx context.Context) ([]*domain.RiskRule, error) {
	var models []RiskRuleModel
	if err := r.db.WithContext(ctx).Order("priority asc").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.RiskRule, len(models))
	for i, m := range models {
		result[i] = toDomainRiskRule(&m)
	}
	return result, nil
}

func (r *riskRuleRepository) FindByType(ctx context.Context, ruleType domain.RuleType) ([]*domain.RiskRule, error) {
	var models []RiskRuleModel
	if err := r.db.WithContext(ctx).Where("type = ?", ruleType).Order("priority asc").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.RiskRule, len(models))
	for i, m := range models {
		result[i] = toDomainRiskRule(&m)
	}
	return result, nil
}

func (r *riskRuleRepository) FindEnabled(ctx context.Context) ([]*domain.RiskRule, error) {
	var models []RiskRuleModel
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("priority asc").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.RiskRule, len(models))
	for i, m := range models {
		result[i] = toDomainRiskRule(&m)
	}
	return result, nil
}

func (r *riskRuleRepository) Update(ctx context.Context, rule *domain.RiskRule) error {
	model := toRiskRuleModel(rule)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *riskRuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&RiskRuleModel{}, "id = ?", id).Error
}

type deviceFingerprintRepository struct {
	db *gorm.DB
}

func NewDeviceFingerprintRepository(db *gorm.DB) domain.DeviceFingerprintRepository {
	return &deviceFingerprintRepository{db: db}
}

func (r *deviceFingerprintRepository) Save(ctx context.Context, fingerprint *domain.DeviceFingerprint) error {
	model := toDeviceFingerprintModel(fingerprint)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *deviceFingerprintRepository) FindByFingerprint(ctx context.Context, fingerprint string) (*domain.DeviceFingerprint, error) {
	var model DeviceFingerprintModel
	if err := r.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainDeviceFingerprint(&model), nil
}

type userRiskProfileRepository struct {
	db *gorm.DB
}

func NewUserRiskProfileRepository(db *gorm.DB) domain.UserRiskProfileRepository {
	return &userRiskProfileRepository{db: db}
}

func (r *userRiskProfileRepository) Save(ctx context.Context, profile *domain.UserRiskProfile) error {
	model := toUserRiskProfileModel(profile)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *userRiskProfileRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.UserRiskProfile, error) {
	var model UserRiskProfileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainUserRiskProfile(&model), nil
}

type suspiciousActivityRepository struct {
	db *gorm.DB
}

func NewSuspiciousActivityRepository(db *gorm.DB) domain.SuspiciousActivityRepository {
	return &suspiciousActivityRepository{db: db}
}

func (r *suspiciousActivityRepository) Save(ctx context.Context, activity *domain.SuspiciousActivity) error {
	model := toSuspiciousActivityModel(activity)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *suspiciousActivityRepository) FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*domain.SuspiciousActivity, error) {
	var models []SuspiciousActivityModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.SuspiciousActivity, len(models))
	for i, m := range models {
		result[i] = toDomainSuspiciousActivity(&m)
	}
	return result, nil
}

func toRiskAssessmentModel(a *domain.RiskAssessment) *RiskAssessmentModel {
	factors, _ := json.Marshal(a.RiskFactors)
	rules, _ := json.Marshal(a.TriggeredRules)
	return &RiskAssessmentModel{
		TransactionID:    a.TransactionID,
		UserID:           a.UserID,
		RiskLevel:        int(a.RiskLevel),
		RiskScore:        a.RiskScore,
		RecommendedAction: string(a.RecommendedAction),
		RiskFactors:      string(factors),
		TriggeredRules:   string(rules),
		Explanation:      a.Explanation,
	}
}

func toDomainRiskAssessment(m *RiskAssessmentModel) *domain.RiskAssessment {
	var factors []*domain.RiskFactor
	var rules []string
	json.Unmarshal([]byte(m.RiskFactors), &factors)
	json.Unmarshal([]byte(m.TriggeredRules), &rules)
	return &domain.RiskAssessment{
		ID:               string(rune(m.ID)),
		CreatedAt:        m.CreatedAt,
		TransactionID:    m.TransactionID,
		UserID:           m.UserID,
		RiskLevel:        domain.RiskLevel(m.RiskLevel),
		RiskScore:        m.RiskScore,
		RecommendedAction: domain.RiskAction(m.RecommendedAction),
		RiskFactors:      factors,
		TriggeredRules:   rules,
		Explanation:      m.Explanation,
	}
}

func toBlacklistModel(e *domain.BlacklistEntry) *BlacklistModel {
	return &BlacklistModel{
		Type:      string(e.Type),
		Value:     e.Value,
		Reason:    e.Reason,
		CreatedBy: e.CreatedBy,
		ExpiresAt: e.ExpiresAt,
		IsActive:  e.IsActive,
	}
}

func toDomainBlacklistEntry(m *BlacklistModel) *domain.BlacklistEntry {
	return &domain.BlacklistEntry{
		ID:        string(rune(m.ID)),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Type:      domain.BlacklistType(m.Type),
		Value:     m.Value,
		Reason:    m.Reason,
		CreatedBy: m.CreatedBy,
		ExpiresAt: m.ExpiresAt,
		IsActive:  m.IsActive,
	}
}

func toRiskRuleModel(r *domain.RiskRule) *RiskRuleModel {
	return &RiskRuleModel{
		Name:        r.Name,
		Description: r.Description,
		Type:        string(r.Type),
		Condition:   r.Condition,
		RiskWeight:  r.RiskWeight,
		Action:      string(r.Action),
		Enabled:     r.Enabled,
		Priority:    r.Priority,
	}
}

func toDomainRiskRule(m *RiskRuleModel) *domain.RiskRule {
	return &domain.RiskRule{
		ID:          string(rune(m.ID)),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Name:        m.Name,
		Description: m.Description,
		Type:        domain.RuleType(m.Type),
		Condition:   m.Condition,
		RiskWeight:  m.RiskWeight,
		Action:      domain.RiskAction(m.Action),
		Enabled:     m.Enabled,
		Priority:    m.Priority,
	}
}

func toDeviceFingerprintModel(d *domain.DeviceFingerprint) *DeviceFingerprintModel {
	accounts, _ := json.Marshal(d.AssociatedAccounts)
	ips, _ := json.Marshal(d.AssociatedIPs)
	return &DeviceFingerprintModel{
		Fingerprint:        d.Fingerprint,
		FirstSeenUserID:    d.FirstSeenUserID,
		AssociatedAccounts: string(accounts),
		AssociatedIPs:      string(ips),
		IsSuspicious:       d.IsSuspicious,
		DeviceType:         d.DeviceType,
		OS:                 d.OS,
		Browser:            d.Browser,
		FirstSeenAt:        d.FirstSeenAt,
		LastSeenAt:         d.LastSeenAt,
	}
}

func toDomainDeviceFingerprint(m *DeviceFingerprintModel) *domain.DeviceFingerprint {
	var accounts []uint64
	var ips []string
	json.Unmarshal([]byte(m.AssociatedAccounts), &accounts)
	json.Unmarshal([]byte(m.AssociatedIPs), &ips)
	return &domain.DeviceFingerprint{
		Fingerprint:            m.Fingerprint,
		FirstSeenUserID:        m.FirstSeenUserID,
		AssociatedAccounts:     accounts,
		AssociatedIPs:          ips,
		IsSuspicious:           m.IsSuspicious,
		DeviceType:             m.DeviceType,
		OS:                     m.OS,
		Browser:                m.Browser,
		FirstSeenAt:            m.FirstSeenAt,
		LastSeenAt:             m.LastSeenAt,
		AssociatedAccountsCount: len(accounts),
	}
}

func toUserRiskProfileModel(p *domain.UserRiskProfile) *UserRiskProfileModel {
	tags, _ := json.Marshal(p.RiskTags)
	return &UserRiskProfileModel{
		UserID:              p.UserID,
		RiskScore:           p.RiskScore,
		RiskLevel:           int(p.RiskLevel),
		TotalTransactions:   p.TotalTransactions,
		FlaggedTransactions: p.FlaggedTransactions,
		BlockedTransactions: p.BlockedTransactions,
		RiskTags:            string(tags),
		FirstTransactionAt:  p.FirstTransactionAt,
		LastTransactionAt:   p.LastTransactionAt,
		LastAssessmentAt:    p.LastAssessmentAt,
	}
}

func toDomainUserRiskProfile(m *UserRiskProfileModel) *domain.UserRiskProfile {
	var tags []string
	json.Unmarshal([]byte(m.RiskTags), &tags)
	return &domain.UserRiskProfile{
		UserID:              m.UserID,
		RiskScore:           m.RiskScore,
		RiskLevel:           domain.RiskLevel(m.RiskLevel),
		TotalTransactions:   m.TotalTransactions,
		FlaggedTransactions: m.FlaggedTransactions,
		BlockedTransactions: m.BlockedTransactions,
		RiskTags:            tags,
		FirstTransactionAt:  m.FirstTransactionAt,
		LastTransactionAt:   m.LastTransactionAt,
		LastAssessmentAt:    m.LastAssessmentAt,
	}
}

func toSuspiciousActivityModel(a *domain.SuspiciousActivity) *SuspiciousActivityModel {
	metadata, _ := json.Marshal(a.Metadata)
	return &SuspiciousActivityModel{
		UserID:            a.UserID,
		ActivityType:      a.ActivityType,
		Description:       a.Description,
		IPAddress:         a.IPAddress,
		DeviceFingerprint: a.DeviceFingerprint,
		Metadata:          string(metadata),
	}
}

func toDomainSuspiciousActivity(m *SuspiciousActivityModel) *domain.SuspiciousActivity {
	var metadata map[string]string
	json.Unmarshal([]byte(m.Metadata), &metadata)
	return &domain.SuspiciousActivity{
		ID:                string(rune(m.ID)),
		CreatedAt:         m.CreatedAt,
		UserID:            m.UserID,
		ActivityType:      m.ActivityType,
		Description:       m.Description,
		IPAddress:         m.IPAddress,
		DeviceFingerprint: m.DeviceFingerprint,
		Metadata:          metadata,
	}
}
