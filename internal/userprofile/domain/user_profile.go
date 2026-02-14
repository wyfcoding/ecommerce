package domain

import (
	"context"
	"time"
)

type ProfileStatus int8

const (
	ProfileStatusActive   ProfileStatus = 1
	ProfileStatusInactive ProfileStatus = 2
	ProfileStatusArchived ProfileStatus = 3
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

type UserProfile struct {
	ID                  uint64              `json:"id"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	UserID              uint64              `json:"user_id"`
	Status              ProfileStatus       `json:"status"`
	OverallScore        int                 `json:"overall_score"`
	ActivityScore       int                 `json:"activity_score"`
	EngagementScore     int                 `json:"engagement_score"`
	ValueScore          int                 `json:"value_score"`
	LoyaltyScore        int                 `json:"loyalty_score"`
	RiskScore           int                 `json:"risk_score"`
	RiskLevel           int                 `json:"risk_level"`
	BehaviorScore       int                 `json:"behavior_score"`
	ConsumptionScore    int                 `json:"consumption_score"`
	Tags                []*UserTag          `json:"tags"`
	TagCount            int                 `json:"tag_count"`
	BehaviorFeatures    *BehaviorFeatures   `json:"behavior_features"`
	Preferences         *UserPreferences    `json:"preferences"`
	ConsumptionProfile  *ConsumptionProfile `json:"consumption_profile"`
	Segment             string              `json:"segment"`
	SegmentScore        float64             `json:"segment_score"`
	ValueSegment        string              `json:"value_segment"`
	LifecycleStage      string              `json:"lifecycle_stage"`
	LastActiveAt        *time.Time          `json:"last_active_at"`
	LastPurchaseAt      *time.Time          `json:"last_purchase_at"`
	LastCalculatedAt    *time.Time          `json:"last_calculated_at"`
	NextCalculateAt     *time.Time          `json:"next_calculate_at"`
	CalculateVersion    int                 `json:"calculate_version"`
	RecentActivities    []*RecentActivity   `json:"recent_activities"`
	ActivityCount       int64               `json:"activity_count"`
	PurchaseCount       int64               `json:"purchase_count"`
	TotalSpent          int64               `json:"total_spent"`
	ProfileCompleteness float64             `json:"profile_completeness"`
	Verified            bool                `json:"verified"`
	VerificationLevel   int                 `json:"verification_level"`
	Metadata            map[string]any      `json:"metadata"`
}

type RecentActivity struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	ProfileID    uint64    `json:"profile_id"`
	UserID       uint64    `json:"user_id"`
	ActivityType string    `json:"activity_type"`
	TargetType   string    `json:"target_type"`
	TargetID     uint64    `json:"target_id"`
	Value        string    `json:"value"`
	Score        int       `json:"score"`
}

func NewUserProfile(userID uint64) *UserProfile {
	return &UserProfile{
		UserID:           userID,
		Status:           ProfileStatusActive,
		Tags:             make([]*UserTag, 0),
		RecentActivities: make([]*RecentActivity, 0),
		Metadata:         make(map[string]any),
		CalculateVersion: 1,
	}
}

func (p *UserProfile) AddTag(tag *UserTag) error {
	for i, t := range p.Tags {
		if t.TagKey == tag.TagKey {
			p.Tags[i] = tag
			return nil
		}
	}
	p.Tags = append(p.Tags, tag)
	p.TagCount = len(p.Tags)
	p.UpdateUpdatedAt()
	return nil
}

func (p *UserProfile) RemoveTag(tagKey string) error {
	for i, t := range p.Tags {
		if t.TagKey == tagKey {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			p.TagCount = len(p.Tags)
			p.UpdateUpdatedAt()
			return nil
		}
	}
	return nil
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

func (p *UserProfile) GetActiveTags() []*UserTag {
	var tags []*UserTag
	for _, t := range p.Tags {
		if t.IsValid() {
			tags = append(tags, t)
		}
	}
	return tags
}

func (p *UserProfile) UpdateBehaviorFeatures(features *BehaviorFeatures) {
	p.BehaviorFeatures = features
	if features != nil {
		p.ActivityScore = features.ActivityScore
		p.BehaviorScore = features.ActivityScore
		p.EngagementScore = features.EngagementScore
	}
	p.UpdateUpdatedAt()
}

func (p *UserProfile) UpdatePreferences(preferences *UserPreferences) {
	p.Preferences = preferences
	p.UpdateUpdatedAt()
}

func (p *UserProfile) UpdateConsumptionProfile(consumption *ConsumptionProfile) {
	p.ConsumptionProfile = consumption
	if consumption != nil {
		p.ConsumptionScore = int(consumption.SpendingLevel)
		p.ValueScore = int(consumption.ValueSegment)
		p.ValueSegment = consumption.ValueSegment.String()
		p.TotalSpent = consumption.TotalSpent
		p.PurchaseCount = consumption.TotalOrders
		if consumption.LastPurchaseAt != nil {
			p.LastPurchaseAt = consumption.LastPurchaseAt
		}
	}
	p.UpdateUpdatedAt()
}

func (p *UserProfile) RecordActivity(activityType string) {
	activity := &RecentActivity{
		ProfileID:    p.ID,
		UserID:       p.UserID,
		ActivityType: activityType,
	}
	p.RecentActivities = append(p.RecentActivities, activity)
	if len(p.RecentActivities) > 100 {
		p.RecentActivities = p.RecentActivities[len(p.RecentActivities)-100:]
	}
	p.ActivityCount++
	now := time.Now()
	p.LastActiveAt = &now
	p.UpdateUpdatedAt()
}

func (p *UserProfile) CalculateOverallScore() {
	score := 0

	score += p.ActivityScore * 2
	score += p.EngagementScore
	score += p.ValueScore
	score += p.LoyaltyScore
	score += p.BehaviorScore

	if p.ConsumptionProfile != nil {
		score += int(p.ConsumptionProfile.SpendingLevel) * 5
	}

	p.OverallScore = score
	p.determineSegment()
	p.determineLifecycleStage()

	now := time.Now()
	p.LastCalculatedAt = &now
	p.CalculateVersion++
	p.UpdateUpdatedAt()
}

func (p *UserProfile) determineSegment() {
	switch {
	case p.OverallScore >= 80:
		p.Segment = "VIP"
	case p.OverallScore >= 60:
		p.Segment = "HIGH_VALUE"
	case p.OverallScore >= 40:
		p.Segment = "MEDIUM_VALUE"
	case p.OverallScore >= 20:
		p.Segment = "LOW_VALUE"
	default:
		p.Segment = "NEW"
	}
}

func (p *UserProfile) determineLifecycleStage() {
	if p.LastActiveAt == nil {
		p.LifecycleStage = "NEW"
		return
	}

	daysSinceActive := int(time.Since(*p.LastActiveAt).Hours() / 24)

	switch {
	case daysSinceActive > 90:
		p.LifecycleStage = "CHURNED"
	case daysSinceActive > 60:
		p.LifecycleStage = "DORMANT"
	case daysSinceActive > 30:
		p.LifecycleStage = "AT_RISK"
	case p.ActivityCount > 100 && p.PurchaseCount > 10:
		p.LifecycleStage = "LOYAL"
	case p.PurchaseCount > 3:
		p.LifecycleStage = "ACTIVE"
	default:
		p.LifecycleStage = "NEW"
	}
}

func (p *UserProfile) SetNextCalculateTime(duration time.Duration) {
	nextTime := time.Now().Add(duration)
	p.NextCalculateAt = &nextTime
}

func (p *UserProfile) NeedsRecalculation() bool {
	if p.NextCalculateAt == nil {
		return true
	}
	return time.Now().After(*p.NextCalculateAt)
}

func (p *UserProfile) Archive() {
	p.Status = ProfileStatusArchived
	p.UpdateUpdatedAt()
}

func (p *UserProfile) Activate() {
	p.Status = ProfileStatusActive
	p.UpdateUpdatedAt()
}

func (p *UserProfile) Deactivate() {
	p.Status = ProfileStatusInactive
	p.UpdateUpdatedAt()
}

func (p *UserProfile) IsArchived() bool {
	return p.Status == ProfileStatusArchived
}

func (p *UserProfile) IsActive() bool {
	return p.Status == ProfileStatusActive
}

func (p *UserProfile) UpdateUpdatedAt() {
	p.UpdatedAt = time.Now()
}

func (p *UserProfile) SetMetadata(key string, value any) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]any)
	}
	p.Metadata[key] = value
	p.UpdateUpdatedAt()
}

func (p *UserProfile) GetMetadata(key string) (any, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	val, ok := p.Metadata[key]
	return val, ok
}

func (p *UserProfile) CalculateProfileCompleteness() float64 {
	completeness := 0.0

	if len(p.Tags) > 0 {
		completeness += 20
	}
	if p.BehaviorFeatures != nil && p.BehaviorFeatures.BrowseCount > 0 {
		completeness += 20
	}
	if p.Preferences != nil && len(p.Preferences.CategoryPreferences) > 0 {
		completeness += 20
	}
	if p.ConsumptionProfile != nil && p.ConsumptionProfile.TotalOrders > 0 {
		completeness += 20
	}
	if p.PurchaseCount > 0 {
		completeness += 20
	}

	p.ProfileCompleteness = completeness
	return completeness
}

type ProfileCalculator interface {
	CalculateBehaviorFeatures(ctx context.Context, userID uint64) (*BehaviorFeatures, error)
	CalculatePreferences(ctx context.Context, userID uint64) (*UserPreferences, error)
	CalculateConsumptionProfile(ctx context.Context, userID uint64) (*ConsumptionProfile, error)
	CalculateScores(ctx context.Context, profile *UserProfile) error
}

type UserProfileRepository interface {
	Save(ctx context.Context, profile *UserProfile) error
	FindByID(ctx context.Context, id uint64) (*UserProfile, error)
	FindByUserID(ctx context.Context, userID uint64) (*UserProfile, error)
	FindBySegment(ctx context.Context, segment string, limit int) ([]*UserProfile, error)
	FindByStatus(ctx context.Context, status ProfileStatus, limit int) ([]*UserProfile, error)
	FindNeedingRecalculation(ctx context.Context, limit int) ([]*UserProfile, error)
	FindHighValue(ctx context.Context, threshold int, limit int) ([]*UserProfile, error)
	FindAtRisk(ctx context.Context, limit int) ([]*UserProfile, error)
	FindByTag(ctx context.Context, tagKey, tagValue string, limit, offset int) ([]*UserProfile, error)
	FindByScoreRange(ctx context.Context, scoreType string, minScore, maxScore int, limit int) ([]*UserProfile, error)
	FindNeedRecalculation(ctx context.Context, limit int) ([]*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
	Delete(ctx context.Context, userID uint64) error
}
