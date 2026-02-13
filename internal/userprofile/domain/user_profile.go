package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserProfileNotFound   = errors.New("user profile not found")
	ErrTagNotFound           = errors.New("tag not found")
	ErrInvalidTagValue       = errors.New("invalid tag value")
	ErrProfileAlreadyExpired = errors.New("profile already expired")
)

type ProfileStatus int8

const (
	ProfileStatusActive    ProfileStatus = 1
	ProfileStatusInactive  ProfileStatus = 2
	ProfileStatusArchived  ProfileStatus = 3
)

func (s ProfileStatus) String() string {
	switch s {
	case ProfileStatusActive:
		return "ACTIVE"
	case ProfileStatusInactive:
		return "INACTIVE"
	case ProfileStatusArchived:
		return "ARCHIVED"
	default:
		return "UNKNOWN"
	}
}

type ProfileScore int8

const (
	ProfileScoreLow      ProfileScore = 1
	ProfileScoreMedium   ProfileScore = 2
	ProfileScoreHigh     ProfileScore = 3
	ProfileScoreVeryHigh ProfileScore = 4
)

type UserProfile struct {
	ID                  uint64              `json:"id"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	UserID              uint64              `json:"user_id"`
	Status              ProfileStatus       `json:"status"`
	ProfileVersion      int                 `json:"profile_version"`
	LastCalculatedAt    *time.Time          `json:"last_calculated_at"`
	NextCalculateAt     *time.Time          `json:"next_calculate_at"`
	Tags                []*UserTag          `json:"tags"`
	BehaviorFeatures    *BehaviorFeatures   `json:"behavior_features"`
	Preferences         *UserPreferences    `json:"preferences"`
	ConsumptionProfile  *ConsumptionProfile `json:"consumption_profile"`
	RiskProfile         *RiskProfile        `json:"risk_profile"`
	SocialProfile       *SocialProfile      `json:"social_profile"`
	ActivityScore       int                 `json:"activity_score"`
	EngagementScore     int                 `json:"engagement_score"`
	ValueScore          int                 `json:"value_score"`
	LoyaltyScore        int                 `json:"loyalty_score"`
	OverallScore        int                 `json:"overall_score"`
	ProfileCompleteness int                 `json:"profile_completeness"`
	DataSources         []string            `json:"data_sources"`
	LastActiveAt        *time.Time          `json:"last_active_at"`
	ExpiresAt           *time.Time          `json:"expires_at"`
}

type RiskProfile struct {
	RiskLevel       int     `json:"risk_level"`
	FraudScore      float64 `json:"fraud_score"`
	CreditScore     int     `json:"credit_score"`
	TrustScore      int     `json:"trust_score"`
	BlacklistFlags  []string `json:"blacklist_flags"`
	LastRiskCheckAt *time.Time `json:"last_risk_check_at"`
}

type SocialProfile struct {
	ShareCount      int     `json:"share_count"`
	CommentCount    int     `json:"comment_count"`
	LikeCount       int     `json:"like_count"`
	FollowCount     int     `json:"follow_count"`
	FanCount        int     `json:"fan_count"`
	SocialInfluence float64 `json:"social_influence"`
	ReferralCount   int     `json:"referral_count"`
}

func NewUserProfile(userID uint64) *UserProfile {
	return &UserProfile{
		UserID:         userID,
		Status:         ProfileStatusActive,
		ProfileVersion: 1,
		Tags:           make([]*UserTag, 0),
		DataSources:    make([]string, 0),
	}
}

func (p *UserProfile) AddTag(tag *UserTag) error {
	for _, t := range p.Tags {
		if t.TagKey == tag.TagKey {
			t.TagValue = tag.TagValue
			t.Confidence = tag.Confidence
			t.Source = tag.Source
			t.UpdatedAt = time.Now()
			return nil
		}
	}
	p.Tags = append(p.Tags, tag)
	return nil
}

func (p *UserProfile) RemoveTag(tagKey string) error {
	for i, t := range p.Tags {
		if t.TagKey == tagKey {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			return nil
		}
	}
	return ErrTagNotFound
}

func (p *UserProfile) GetTag(tagKey string) *UserTag {
	for _, t := range p.Tags {
		if t.TagKey == tagKey {
			return t
		}
	}
	return nil
}

func (p *UserProfile) GetTagsByCategory(category TagCategory) []*UserTag {
	var tags []*UserTag
	for _, t := range p.Tags {
		if t.Category == category {
			tags = append(tags, t)
		}
	}
	return tags
}

func (p *UserProfile) UpdateBehaviorFeatures(features *BehaviorFeatures) {
	p.BehaviorFeatures = features
	p.markUpdated()
}

func (p *UserProfile) UpdatePreferences(preferences *UserPreferences) {
	p.Preferences = preferences
	p.markUpdated()
}

func (p *UserProfile) UpdateConsumptionProfile(profile *ConsumptionProfile) {
	p.ConsumptionProfile = profile
	p.markUpdated()
}

func (p *UserProfile) CalculateOverallScore() {
	var totalScore int
	var count int

	if p.ActivityScore > 0 {
		totalScore += p.ActivityScore
		count++
	}
	if p.EngagementScore > 0 {
		totalScore += p.EngagementScore
		count++
	}
	if p.ValueScore > 0 {
		totalScore += p.ValueScore
		count++
	}
	if p.LoyaltyScore > 0 {
		totalScore += p.LoyaltyScore
		count++
	}

	if count > 0 {
		p.OverallScore = totalScore / count
	}

	p.calculateCompleteness()
}

func (p *UserProfile) calculateCompleteness() {
	var completeness int

	if len(p.Tags) > 0 {
		completeness += 20
	}
	if p.BehaviorFeatures != nil {
		completeness += 25
	}
	if p.Preferences != nil {
		completeness += 25
	}
	if p.ConsumptionProfile != nil {
		completeness += 20
	}
	if p.RiskProfile != nil {
		completeness += 10
	}

	p.ProfileCompleteness = completeness
}

func (p *UserProfile) markUpdated() {
	now := time.Now()
	p.UpdatedAt = now
	p.ProfileVersion++
	p.LastCalculatedAt = &now
}

func (p *UserProfile) SetNextCalculateTime(duration time.Duration) {
	nextTime := time.Now().Add(duration)
	p.NextCalculateAt = &nextTime
}

func (p *UserProfile) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}

func (p *UserProfile) NeedsRecalculation() bool {
	if p.NextCalculateAt == nil {
		return true
	}
	return time.Now().After(*p.NextCalculateAt)
}

func (p *UserProfile) Activate() {
	p.Status = ProfileStatusActive
}

func (p *UserProfile) Deactivate() {
	p.Status = ProfileStatusInactive
}

func (p *UserProfile) Archive() {
	p.Status = ProfileStatusArchived
}

func (p *UserProfile) IsActive() bool {
	return p.Status == ProfileStatusActive
}

func (p *UserProfile) AddDataSource(source string) {
	for _, s := range p.DataSources {
		if s == source {
			return
		}
	}
	p.DataSources = append(p.DataSources, source)
}

func (p *UserProfile) RecordActivity(activityType string) {
	now := time.Now()
	p.LastActiveAt = &now

	switch activityType {
	case "browse":
		if p.BehaviorFeatures != nil {
			p.BehaviorFeatures.BrowseCount++
		}
	case "search":
		if p.BehaviorFeatures != nil {
			p.BehaviorFeatures.SearchCount++
		}
	case "purchase":
		if p.BehaviorFeatures != nil {
			p.BehaviorFeatures.PurchaseCount++
		}
	case "share":
		if p.SocialProfile != nil {
			p.SocialProfile.ShareCount++
		}
	}
}

type UserProfileRepository interface {
	Save(ctx context.Context, profile *UserProfile) error
	FindByID(ctx context.Context, id uint64) (*UserProfile, error)
	FindByUserID(ctx context.Context, userID uint64) (*UserProfile, error)
	FindByTag(ctx context.Context, tagKey, tagValue string, limit, offset int) ([]*UserProfile, error)
	FindByScoreRange(ctx context.Context, scoreType string, minScore, maxScore int, limit int) ([]*UserProfile, error)
	FindActive(ctx context.Context, limit, offset int) ([]*UserProfile, error)
	FindNeedRecalculation(ctx context.Context, limit int) ([]*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
	Delete(ctx context.Context, userID uint64) error
}

type ProfileCalculator interface {
	CalculateProfile(ctx context.Context, userID uint64) (*UserProfile, error)
	CalculateBehaviorFeatures(ctx context.Context, userID uint64) (*BehaviorFeatures, error)
	CalculatePreferences(ctx context.Context, userID uint64) (*UserPreferences, error)
	CalculateConsumptionProfile(ctx context.Context, userID uint64) (*ConsumptionProfile, error)
	CalculateScores(ctx context.Context, profile *UserProfile) error
}

type ProfileService interface {
	GetOrCreateProfile(ctx context.Context, userID uint64) (*UserProfile, error)
	UpdateTag(ctx context.Context, userID uint64, tag *UserTag) error
	BatchUpdateTags(ctx context.Context, userID uint64, tags []*UserTag) error
	RecalculateProfile(ctx context.Context, userID uint64) error
	GetProfilesBySegment(ctx context.Context, segment string, limit int) ([]*UserProfile, error)
}
