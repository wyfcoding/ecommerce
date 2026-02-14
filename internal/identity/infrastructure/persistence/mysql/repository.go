package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/identity/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserPersonaModel struct {
	gorm.Model
	ID          uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      uint64  `gorm:"column:user_id;uniqueIndex;not null"`
	Type        string  `gorm:"column:type;type:varchar(32);not null"`
	Tags        string  `gorm:"column:tags;type:text"`
	Interests   string  `gorm:"column:interests;type:text"`
	Industry    string  `gorm:"column:industry;type:varchar(64)"`
	Skills      string  `gorm:"column:skills;type:text"`
	Certs       string  `gorm:"column:certs;type:text"`
	Experience  int     `gorm:"column:experience;default:0"`
	AvgOrderVal int64   `gorm:"column:avg_order_val;default:0"`
	FreqCat     string  `gorm:"column:freq_cat;type:text"`
	PrefTime    string  `gorm:"column:pref_time;type:varchar(32)"`
	Followers   int32   `gorm:"column:followers;default:0"`
	Following   int32   `gorm:"column:following;default:0"`
	Engagement  float64 `gorm:"column:engagement;default:0"`
}

func (UserPersonaModel) TableName() string {
	return "identity_user_personas"
}

type LinkedAccountModel struct {
	gorm.Model
	ID            uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID        uint64     `gorm:"column:user_id;index;not null"`
	Provider      string     `gorm:"column:provider;type:varchar(32);not null"`
	ExternalID    string     `gorm:"column:external_id;type:varchar(128);not null"`
	ExternalName  string     `gorm:"column:external_name;type:varchar(128)"`
	ExternalEmail string     `gorm:"column:external_email;type:varchar(128)"`
	AccessToken   string     `gorm:"column:access_token;type:text"`
	RefreshToken  string     `gorm:"column:refresh_token;type:text"`
	Expiry        *time.Time `gorm:"column:expiry"`
	AvatarURL     string     `gorm:"column:avatar_url;type:varchar(512)"`
	BoundAt       time.Time  `gorm:"column:bound_at;not null"`
}

func (LinkedAccountModel) TableName() string {
	return "identity_linked_accounts"
}

type AuthSessionModel struct {
	gorm.Model
	ID           string    `gorm:"column:id;type:varchar(64);primaryKey"`
	UserID       uint64    `gorm:"column:user_id;index;not null"`
	ClientID     string    `gorm:"column:client_id;type:varchar(64)"`
	DeviceID     string    `gorm:"column:device_id;type:varchar(128)"`
	DeviceName   string    `gorm:"column:device_name;type:varchar(128)"`
	IPAddress    string    `gorm:"column:ip_address;type:varchar(64)"`
	Location     string    `gorm:"column:location;type:varchar(128)"`
	UserAgent    string    `gorm:"column:user_agent;type:varchar(256)"`
	Status       int       `gorm:"column:status;not null;default:1"`
	LastActiveAt time.Time `gorm:"column:last_active_at"`
	ExpiresAt    time.Time `gorm:"column:expires_at;index"`
}

func (AuthSessionModel) TableName() string {
	return "identity_auth_sessions"
}

type UserMappingModel struct {
	gorm.Model
	EcommerceUserID string    `gorm:"column:ecommerce_user_id;type:varchar(64);uniqueIndex;not null"`
	TradingUserID   string    `gorm:"column:trading_user_id;type:varchar(64);uniqueIndex;not null"`
	BoundAt         time.Time `gorm:"column:bound_at;not null"`
	Status          string    `gorm:"column:status;type:varchar(32);not null;default:'ACTIVE'"`
}

func (UserMappingModel) TableName() string {
	return "identity_user_mappings"
}

type KYCRecordModel struct {
	gorm.Model
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID         uint64     `gorm:"column:user_id;uniqueIndex;not null"`
	Level          int        `gorm:"column:level;not null;default:0"`
	Status         int        `gorm:"column:status;not null;default:0"`
	IDType         string     `gorm:"column:id_type;type:varchar(32)"`
	IDNumber       string     `gorm:"column:id_number;type:varchar(64)"`
	IDName         string     `gorm:"column:id_name;type:varchar(128)"`
	FrontImage     string     `gorm:"column:front_image;type:varchar(512)"`
	BackImage      string     `gorm:"column:back_image;type:varchar(512)"`
	SelfieImage    string     `gorm:"column:selfie_image;type:varchar(512)"`
	FaceMatchScore float64    `gorm:"column:face_match_score;default:0"`
	RiskScore      int        `gorm:"column:risk_score;default:0"`
	Notes          string     `gorm:"column:notes;type:text"`
	SubmittedAt    *time.Time `gorm:"column:submitted_at"`
	ReviewedAt     *time.Time `gorm:"column:reviewed_at"`
	ReviewedBy     string     `gorm:"column:reviewed_by;type:varchar(64)"`
	ExpiredAt      *time.Time `gorm:"column:expired_at"`
}

func (KYCRecordModel) TableName() string {
	return "identity_kyc_records"
}

type KYCSessionModel struct {
	gorm.Model
	ID          string    `gorm:"column:id;type:varchar(64);primaryKey"`
	UserID      uint64    `gorm:"column:user_id;index;not null"`
	TargetLevel int       `gorm:"column:target_level;not null"`
	Step        string    `gorm:"column:step;type:varchar(32)"`
	IsCompleted bool      `gorm:"column:is_completed;default:false"`
	Result      string    `gorm:"column:result;type:varchar(32)"`
	StartTime   time.Time `gorm:"column:start_time;not null"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null"`
}

func (KYCSessionModel) TableName() string {
	return "identity_kyc_sessions"
}

type RiskProfileModel struct {
	gorm.Model
	UserID        uint64    `gorm:"column:user_id;uniqueIndex;not null"`
	RiskLevel     string    `gorm:"column:risk_level;type:varchar(32);not null"`
	Score         int       `gorm:"column:score;default:0"`
	Labels        string    `gorm:"column:labels;type:text"`
	LastCheckedAt time.Time `gorm:"column:last_checked_at"`
}

func (RiskProfileModel) TableName() string {
	return "identity_risk_profiles"
}

type IdentityRepository struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) domain.IdentityRepository {
	return &IdentityRepository{db: db}
}

func (r *IdentityRepository) FindPersonaByUserID(ctx context.Context, userID uint64) (*domain.UserPersona, error) {
	var model UserPersonaModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPersonaDomain(&model), nil
}

func (r *IdentityRepository) SavePersona(ctx context.Context, persona *domain.UserPersona) error {
	model := toPersonaModel(persona)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "tags", "interests", "industry", "skills", "certs", "experience", "avg_order_val", "freq_cat", "pref_time", "followers", "following", "engagement", "updated_at"}),
	}).Create(model).Error
}

func (r *IdentityRepository) FindLinkedAccounts(ctx context.Context, userID uint64) ([]*domain.LinkedAccount, error) {
	var models []LinkedAccountModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	accounts := make([]*domain.LinkedAccount, len(models))
	for i, m := range models {
		accounts[i] = toLinkedAccountDomain(&m)
	}
	return accounts, nil
}

func (r *IdentityRepository) BindAccount(ctx context.Context, account *domain.LinkedAccount) error {
	model := toLinkedAccountModel(account)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *IdentityRepository) UnbindAccount(ctx context.Context, userID uint64, provider domain.AccountProvider) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, string(provider)).
		Delete(&LinkedAccountModel{}).Error
}

func (r *IdentityRepository) FindSession(ctx context.Context, sessionID string) (*domain.AuthSession, error) {
	var model AuthSessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAuthSessionDomain(&model), nil
}

func (r *IdentityRepository) FindSessionsByUserID(ctx context.Context, userID uint64) ([]*domain.AuthSession, error) {
	var models []AuthSessionModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	sessions := make([]*domain.AuthSession, len(models))
	for i, m := range models {
		sessions[i] = toAuthSessionDomain(&m)
	}
	return sessions, nil
}

func (r *IdentityRepository) SaveSession(ctx context.Context, session *domain.AuthSession) error {
	model := toAuthSessionModel(session)
	return r.db.WithContext(ctx).Save(model).Error
}

type KYCRepository struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) domain.KYCRepository {
	return &KYCRepository{db: db}
}

func (r *KYCRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.KYCRecord, error) {
	var model KYCRecordModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toKYCRecordDomain(&model), nil
}

func (r *KYCRepository) Save(ctx context.Context, record *domain.KYCRecord) error {
	model := toKYCRecordModel(record)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *KYCRepository) CreateSession(ctx context.Context, session *domain.KYCSession) error {
	model := toKYCSessionModel(session)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *KYCRepository) GetSession(ctx context.Context, sessionID string) (*domain.KYCSession, error) {
	var model KYCSessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toKYCSessionDomain(&model), nil
}

func (r *KYCRepository) GetRiskProfile(ctx context.Context, userID uint64) (*domain.RiskProfile, error) {
	var model RiskProfileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toRiskProfileDomain(&model), nil
}

func (r *KYCRepository) UpdateRiskProfile(ctx context.Context, profile *domain.RiskProfile) error {
	model := toRiskProfileModel(profile)
	return r.db.WithContext(ctx).Save(model).Error
}

type UserMappingRepository struct {
	db *gorm.DB
}

func NewUserMappingRepository(db *gorm.DB) *UserMappingRepository {
	return &UserMappingRepository{db: db}
}

func (r *UserMappingRepository) FindByEcommerceUserID(ctx context.Context, ecommerceUserID string) (*domain.UserMapping, error) {
	var model UserMappingModel
	if err := r.db.WithContext(ctx).Where("ecommerce_user_id = ?", ecommerceUserID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserMappingDomain(&model), nil
}

func (r *UserMappingRepository) FindByTradingUserID(ctx context.Context, tradingUserID string) (*domain.UserMapping, error) {
	var model UserMappingModel
	if err := r.db.WithContext(ctx).Where("trading_user_id = ?", tradingUserID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserMappingDomain(&model), nil
}

func (r *UserMappingRepository) Save(ctx context.Context, mapping *domain.UserMapping) error {
	model := toUserMappingModel(mapping)
	return r.db.WithContext(ctx).Create(model).Error
}

func toPersonaDomain(m *UserPersonaModel) *domain.UserPersona {
	return &domain.UserPersona{
		ID:        m.ID,
		UserID:    m.UserID,
		Type:      domain.PersonaType(m.Type),
		Tags:      splitString(m.Tags),
		Interests: splitString(m.Interests),
		Professional: &domain.ProfInfo{
			Industry:     m.Industry,
			Skills:       splitString(m.Skills),
			Certificates: splitString(m.Certs),
			Experience:   m.Experience,
		},
		ShoppingHabit: &domain.HabitInfo{
			AvgOrderValue: m.AvgOrderVal,
			FreqCategory:  splitString(m.FreqCat),
			PreferredTime: m.PrefTime,
		},
		SocialMetrics: &domain.SocialInfo{
			Followers:  m.Followers,
			Following:  m.Following,
			Engagement: m.Engagement,
		},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toPersonaModel(p *domain.UserPersona) *UserPersonaModel {
	m := &UserPersonaModel{
		ID:          p.ID,
		UserID:      p.UserID,
		Type:        string(p.Type),
		Tags:        joinString(p.Tags),
		Interests:   joinString(p.Interests),
		AvgOrderVal: p.ShoppingHabit.AvgOrderValue,
		FreqCat:     joinString(p.ShoppingHabit.FreqCategory),
		PrefTime:    p.ShoppingHabit.PreferredTime,
		Followers:   p.SocialMetrics.Followers,
		Following:   p.SocialMetrics.Following,
		Engagement:  p.SocialMetrics.Engagement,
	}
	if p.Professional != nil {
		m.Industry = p.Professional.Industry
		m.Skills = joinString(p.Professional.Skills)
		m.Certs = joinString(p.Professional.Certificates)
		m.Experience = p.Professional.Experience
	}
	return m
}

func toLinkedAccountDomain(m *LinkedAccountModel) *domain.LinkedAccount {
	return &domain.LinkedAccount{
		ID:            m.ID,
		UserID:        m.UserID,
		Provider:      domain.AccountProvider(m.Provider),
		ExternalID:    m.ExternalID,
		ExternalName:  m.ExternalName,
		ExternalEmail: m.ExternalEmail,
		AccessToken:   m.AccessToken,
		RefreshToken:  m.RefreshToken,
		Expiry:        m.Expiry,
		AvatarURL:     m.AvatarURL,
		BoundAt:       m.BoundAt,
	}
}

func toLinkedAccountModel(a *domain.LinkedAccount) *LinkedAccountModel {
	return &LinkedAccountModel{
		UserID:        a.UserID,
		Provider:      string(a.Provider),
		ExternalID:    a.ExternalID,
		ExternalName:  a.ExternalName,
		ExternalEmail: a.ExternalEmail,
		AccessToken:   a.AccessToken,
		RefreshToken:  a.RefreshToken,
		Expiry:        a.Expiry,
		AvatarURL:     a.AvatarURL,
		BoundAt:       a.BoundAt,
	}
}

func toAuthSessionDomain(m *AuthSessionModel) *domain.AuthSession {
	return &domain.AuthSession{
		ID:           m.ID,
		UserID:       m.UserID,
		ClientID:     m.ClientID,
		DeviceID:     m.DeviceID,
		DeviceName:   m.DeviceName,
		IPAddress:    m.IPAddress,
		Location:     m.Location,
		UserAgent:    m.UserAgent,
		Status:       domain.SessionStatus(m.Status),
		LastActiveAt: m.LastActiveAt,
		CreatedAt:    m.CreatedAt,
		ExpiresAt:    m.ExpiresAt,
	}
}

func toAuthSessionModel(s *domain.AuthSession) *AuthSessionModel {
	return &AuthSessionModel{
		ID:           s.ID,
		UserID:       s.UserID,
		ClientID:     s.ClientID,
		DeviceID:     s.DeviceID,
		DeviceName:   s.DeviceName,
		IPAddress:    s.IPAddress,
		Location:     s.Location,
		UserAgent:    s.UserAgent,
		Status:       int(s.Status),
		LastActiveAt: s.LastActiveAt,
		ExpiresAt:    s.ExpiresAt,
	}
}

func toKYCRecordDomain(m *KYCRecordModel) *domain.KYCRecord {
	return &domain.KYCRecord{
		ID:     m.ID,
		UserID: m.UserID,
		Level:  domain.KYCLevel(m.Level),
		Status: domain.KYCStatus(m.Status),
		IDInfo: &domain.IDDocument{
			Type:        domain.IDType(m.IDType),
			Number:      m.IDNumber,
			Name:        m.IDName,
			FrontImage:  m.FrontImage,
			BackImage:   m.BackImage,
			SelfieImage: m.SelfieImage,
		},
		FaceMatchScore: m.FaceMatchScore,
		RiskScore:      m.RiskScore,
		Notes:          m.Notes,
		SubmittedAt:    m.SubmittedAt,
		ReviewedAt:     m.ReviewedAt,
		ReviewedBy:     m.ReviewedBy,
		ExpiredAt:      m.ExpiredAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toKYCRecordModel(r *domain.KYCRecord) *KYCRecordModel {
	m := &KYCRecordModel{
		UserID:         r.UserID,
		Level:          int(r.Level),
		Status:         int(r.Status),
		FaceMatchScore: r.FaceMatchScore,
		RiskScore:      r.RiskScore,
		Notes:          r.Notes,
		SubmittedAt:    r.SubmittedAt,
		ReviewedAt:     r.ReviewedAt,
		ReviewedBy:     r.ReviewedBy,
		ExpiredAt:      r.ExpiredAt,
	}
	if r.IDInfo != nil {
		m.IDType = string(r.IDInfo.Type)
		m.IDNumber = r.IDInfo.Number
		m.IDName = r.IDInfo.Name
		m.FrontImage = r.IDInfo.FrontImage
		m.BackImage = r.IDInfo.BackImage
		m.SelfieImage = r.IDInfo.SelfieImage
	}
	return m
}

func toKYCSessionDomain(m *KYCSessionModel) *domain.KYCSession {
	return &domain.KYCSession{
		ID:          m.ID,
		UserID:      m.UserID,
		TargetLevel: domain.KYCLevel(m.TargetLevel),
		Step:        m.Step,
		IsCompleted: m.IsCompleted,
		Result:      m.Result,
		StartTime:   m.StartTime,
		ExpiresAt:   m.ExpiresAt,
	}
}

func toKYCSessionModel(s *domain.KYCSession) *KYCSessionModel {
	return &KYCSessionModel{
		ID:          s.ID,
		UserID:      s.UserID,
		TargetLevel: int(s.TargetLevel),
		Step:        s.Step,
		IsCompleted: s.IsCompleted,
		Result:      s.Result,
		StartTime:   s.StartTime,
		ExpiresAt:   s.ExpiresAt,
	}
}

func toRiskProfileDomain(m *RiskProfileModel) *domain.RiskProfile {
	return &domain.RiskProfile{
		UserID:        m.UserID,
		RiskLevel:     m.RiskLevel,
		Score:         m.Score,
		Labels:        splitString(m.Labels),
		LastCheckedAt: m.LastCheckedAt,
	}
}

func toRiskProfileModel(p *domain.RiskProfile) *RiskProfileModel {
	return &RiskProfileModel{
		UserID:        p.UserID,
		RiskLevel:     p.RiskLevel,
		Score:         p.Score,
		Labels:        joinString(p.Labels),
		LastCheckedAt: p.LastCheckedAt,
	}
}

func toUserMappingDomain(m *UserMappingModel) *domain.UserMapping {
	return &domain.UserMapping{
		EcommerceUserID: m.EcommerceUserID,
		TradingUserID:   m.TradingUserID,
		BoundAt:         m.BoundAt,
		Status:          m.Status,
	}
}

func toUserMappingModel(m *domain.UserMapping) *UserMappingModel {
	return &UserMappingModel{
		EcommerceUserID: m.EcommerceUserID,
		TradingUserID:   m.TradingUserID,
		BoundAt:         m.BoundAt,
		Status:          m.Status,
	}
}

func splitString(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, v := range []byte(s) {
		if v != 0 {
			result = append(result, string(v))
		}
	}
	return result
}

func joinString(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	var result []byte
	for _, v := range arr {
		if len(v) > 0 {
			result = append(result, v[0])
		}
	}
	return string(result)
}

var _ domain.KYCRepository = (*KYCRepository)(nil)
var _ domain.IdentityRepository = (*IdentityRepository)(nil)

func NullStringToPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func NullTimeToPtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
