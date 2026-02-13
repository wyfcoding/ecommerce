package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
)

type ProfileQueryService struct {
	profileRepo    domain.UserProfileRepository
	tagRepo        domain.UserTagRepository
	behaviorRepo   domain.BehaviorFeaturesRepository
	preferenceRepo domain.UserPreferencesRepository
	consumptionRepo domain.ConsumptionProfileRepository
	segmentRepo    domain.ProfileSegmentRepository
}

func NewProfileQueryService(
	profileRepo domain.UserProfileRepository,
	tagRepo domain.UserTagRepository,
	behaviorRepo domain.BehaviorFeaturesRepository,
	preferenceRepo domain.UserPreferencesRepository,
	consumptionRepo domain.ConsumptionProfileRepository,
	segmentRepo domain.ProfileSegmentRepository,
) *ProfileQueryService {
	return &ProfileQueryService{
		profileRepo:    profileRepo,
		tagRepo:        tagRepo,
		behaviorRepo:   behaviorRepo,
		preferenceRepo: preferenceRepo,
		consumptionRepo: consumptionRepo,
		segmentRepo:    segmentRepo,
	}
}

func (s *ProfileQueryService) GetProfile(ctx context.Context, userID uint64) (*domain.UserProfile, error) {
	return s.profileRepo.FindByUserID(ctx, userID)
}

func (s *ProfileQueryService) GetTags(ctx context.Context, userID uint64) ([]*domain.UserTag, error) {
	return s.tagRepo.FindByUserID(ctx, userID)
}

func (s *ProfileQueryService) GetTagsByCategory(ctx context.Context, userID uint64, category domain.TagCategory) ([]*domain.UserTag, error) {
	return s.tagRepo.FindByCategory(ctx, userID, category)
}

func (s *ProfileQueryService) GetBehaviorFeatures(ctx context.Context, userID uint64) (*domain.BehaviorFeatures, error) {
	return s.behaviorRepo.FindByUserID(ctx, userID)
}

func (s *ProfileQueryService) GetPreferences(ctx context.Context, userID uint64) (*domain.UserPreferences, error) {
	return s.preferenceRepo.FindByUserID(ctx, userID)
}

func (s *ProfileQueryService) GetConsumptionProfile(ctx context.Context, userID uint64) (*domain.ConsumptionProfile, error) {
	return s.consumptionRepo.FindByUserID(ctx, userID)
}

func (s *ProfileQueryService) GetTopCategories(ctx context.Context, userID uint64, limit int) ([]uint64, error) {
	features, err := s.behaviorRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return features.GetTopCategories(limit), nil
}

func (s *ProfileQueryService) GetTopKeywords(ctx context.Context, userID uint64, limit int) ([]string, error) {
	features, err := s.behaviorRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return features.GetTopKeywords(limit), nil
}

func (s *ProfileQueryService) GetPreferredCategories(ctx context.Context, userID uint64, limit int) ([]*domain.CategoryPreference, error) {
	preferences, err := s.preferenceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return preferences.GetTopCategories(limit), nil
}

func (s *ProfileQueryService) GetPreferredBrands(ctx context.Context, userID uint64, limit int) ([]*domain.BrandPreference, error) {
	preferences, err := s.preferenceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return preferences.GetTopBrands(limit), nil
}

func (s *ProfileQueryService) GetUsersByTag(ctx context.Context, tagKey, tagValue string, limit, offset int) ([]*domain.UserProfile, error) {
	return s.profileRepo.FindByTag(ctx, tagKey, tagValue, limit, offset)
}

func (s *ProfileQueryService) GetUsersByScoreRange(ctx context.Context, scoreType string, minScore, maxScore int, limit int) ([]*domain.UserProfile, error) {
	return s.profileRepo.FindByScoreRange(ctx, scoreType, minScore, maxScore, limit)
}

func (s *ProfileQueryService) GetUsersBySpendingLevel(ctx context.Context, level domain.SpendingLevel, limit int) ([]*domain.ConsumptionProfile, error) {
	return s.consumptionRepo.FindBySpendingLevel(ctx, level, limit)
}

func (s *ProfileQueryService) GetUsersByValueSegment(ctx context.Context, segment domain.ValueSegment, limit int) ([]*domain.ConsumptionProfile, error) {
	return s.consumptionRepo.FindByValueSegment(ctx, segment, limit)
}

func (s *ProfileQueryService) GetHighChurnRiskUsers(ctx context.Context, threshold float64, limit int) ([]*domain.ConsumptionProfile, error) {
	return s.consumptionRepo.FindHighChurnRisk(ctx, threshold, limit)
}

func (s *ProfileQueryService) GetProfilesNeedingRecalculation(ctx context.Context, limit int) ([]*domain.UserProfile, error) {
	return s.profileRepo.FindNeedRecalculation(ctx, limit)
}

func (s *ProfileQueryService) GetSegmentUsers(ctx context.Context, segmentNo string, limit, offset int) ([]*domain.UserProfile, error) {
	segment, err := s.segmentRepo.FindBySegmentNo(ctx, segmentNo)
	if err != nil {
		return nil, err
	}

	var profiles []*domain.UserProfile
	for _, criteria := range segment.Criteria {
		results, err := s.profileRepo.FindByTag(ctx, criteria.Field, criteria.Value, limit, offset)
		if err != nil {
			continue
		}
		profiles = append(profiles, results...)
	}

	if len(profiles) > limit {
		profiles = profiles[:limit]
	}

	return profiles, nil
}

type ProfileSummary struct {
	UserID              uint64   `json:"user_id"`
	Status              string   `json:"status"`
	OverallScore        int      `json:"overall_score"`
	ActivityScore       int      `json:"activity_score"`
	EngagementScore     int      `json:"engagement_score"`
	ValueScore          int      `json:"value_score"`
	LoyaltyScore        int      `json:"loyalty_score"`
	ProfileCompleteness int      `json:"profile_completeness"`
	SpendingLevel       string   `json:"spending_level"`
	ValueSegment        string   `json:"value_segment"`
	TopCategories       []uint64 `json:"top_categories"`
	TopBrands           []uint64 `json:"top_brands"`
	ChurnRisk           float64  `json:"churn_risk"`
	LastActiveAt        string   `json:"last_active_at"`
}

func (s *ProfileQueryService) GetProfileSummary(ctx context.Context, userID uint64) (*ProfileSummary, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &ProfileSummary{
		UserID:              profile.UserID,
		Status:              profile.Status.String(),
		OverallScore:        profile.OverallScore,
		ActivityScore:       profile.ActivityScore,
		EngagementScore:     profile.EngagementScore,
		ValueScore:          profile.ValueScore,
		LoyaltyScore:        profile.LoyaltyScore,
		ProfileCompleteness: profile.ProfileCompleteness,
	}

	if profile.ConsumptionProfile != nil {
		summary.SpendingLevel = profile.ConsumptionProfile.SpendingLevel.String()
		summary.ValueSegment = profile.ConsumptionProfile.ValueSegment.String()
		summary.ChurnRisk = profile.ConsumptionProfile.ChurnRisk
	}

	if profile.BehaviorFeatures != nil {
		summary.TopCategories = profile.BehaviorFeatures.GetTopCategories(5)
	}

	if profile.Preferences != nil {
		topBrands := profile.Preferences.GetTopBrands(5)
		brandIDs := make([]uint64, len(topBrands))
		for i, b := range topBrands {
			brandIDs[i] = b.BrandID
		}
		summary.TopBrands = brandIDs
	}

	if profile.LastActiveAt != nil {
		summary.LastActiveAt = profile.LastActiveAt.Format("2006-01-02 15:04:05")
	}

	return summary, nil
}

type TagStatistics struct {
	TagKey    string `json:"tag_key"`
	TagName   string `json:"tag_name"`
	Category  string `json:"category"`
	UserCount int64  `json:"user_count"`
}

func (s *ProfileQueryService) GetTagStatistics(ctx context.Context, tagKey string) (*TagStatistics, error) {
	profiles, err := s.profileRepo.FindByTag(ctx, tagKey, "", 1000, 0)
	if err != nil {
		return nil, err
	}

	stats := &TagStatistics{
		TagKey:    tagKey,
		UserCount: int64(len(profiles)),
	}

	if len(profiles) > 0 {
		tag := profiles[0].GetTag(tagKey)
		if tag != nil {
			stats.TagName = tag.TagKey
			stats.Category = tag.Category.String()
		}
	}

	return stats, nil
}
