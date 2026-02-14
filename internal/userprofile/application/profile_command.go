package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/userprofile/domain"
)

var (
	ErrProfileNotFound   = errors.New("user profile not found")
	ErrTagNotFound       = errors.New("tag not found")
	ErrInvalidTagValue   = errors.New("invalid tag value")
	ErrCalculationFailed = errors.New("profile calculation failed")
)

type ProfileCommandService struct {
	profileRepo     domain.UserProfileRepository
	tagRepo         domain.UserTagRepository
	tagDefRepo      domain.TagDefinitionRepository
	behaviorRepo    domain.BehaviorFeaturesRepository
	preferenceRepo  domain.UserPreferencesRepository
	consumptionRepo domain.ConsumptionProfileRepository
	ruleRepo        domain.ProfileRuleRepository
	segmentRepo     domain.ProfileSegmentRepository
	calculator      domain.ProfileCalculator
	eventPublisher  EventPublisher
}

type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

func NewProfileCommandService(
	profileRepo domain.UserProfileRepository,
	tagRepo domain.UserTagRepository,
	tagDefRepo domain.TagDefinitionRepository,
	behaviorRepo domain.BehaviorFeaturesRepository,
	preferenceRepo domain.UserPreferencesRepository,
	consumptionRepo domain.ConsumptionProfileRepository,
	ruleRepo domain.ProfileRuleRepository,
	segmentRepo domain.ProfileSegmentRepository,
	calculator domain.ProfileCalculator,
	eventPublisher EventPublisher,
) *ProfileCommandService {
	return &ProfileCommandService{
		profileRepo:     profileRepo,
		tagRepo:         tagRepo,
		tagDefRepo:      tagDefRepo,
		behaviorRepo:    behaviorRepo,
		preferenceRepo:  preferenceRepo,
		consumptionRepo: consumptionRepo,
		ruleRepo:        ruleRepo,
		segmentRepo:     segmentRepo,
		calculator:      calculator,
		eventPublisher:  eventPublisher,
	}
}

func (s *ProfileCommandService) CreateProfile(ctx context.Context, userID uint64) (*domain.UserProfile, error) {
	existing, _ := s.profileRepo.FindByUserID(ctx, userID)
	if existing != nil {
		return existing, nil
	}

	profile := domain.NewUserProfile(userID)

	if err := s.profileRepo.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	if s.eventPublisher != nil {
		event := domain.NewProfileCreatedEvent(profile.ID, userID)
		s.eventPublisher.Publish(ctx, event)
	}

	return profile, nil
}

func (s *ProfileCommandService) AddTag(ctx context.Context, userID uint64, tagKey, tagValue string, category domain.TagCategory, source domain.TagSource, confidence float64) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		profile, err = s.CreateProfile(ctx, userID)
		if err != nil {
			return err
		}
	}

	tagDef, err := s.tagDefRepo.FindByTagKey(ctx, tagKey)
	if err == nil && tagDef != nil {
		if !tagDef.IsValueAllowed(tagValue) {
			return ErrInvalidTagValue
		}
	}

	tag := domain.NewUserTag(profile.ID, userID, tagKey, tagValue, category, source)
	tag.SetConfidence(confidence)

	if err := profile.AddTag(tag); err != nil {
		return err
	}

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if err := s.tagRepo.Save(ctx, tag); err != nil {
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if s.eventPublisher != nil {
		event := domain.NewTagAddedEvent(profile.ID, userID, tagKey, tagValue, category)
		s.eventPublisher.Publish(ctx, event)
	}

	return nil
}

func (s *ProfileCommandService) RemoveTag(ctx context.Context, userID uint64, tagKey string) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrProfileNotFound
	}

	tag := profile.GetTag(tagKey)
	if tag == nil {
		return ErrTagNotFound
	}

	if err := profile.RemoveTag(tagKey); err != nil {
		return err
	}

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if s.eventPublisher != nil {
		event := domain.NewTagRemovedEvent(profile.ID, userID, tagKey, tag.TagValue)
		s.eventPublisher.Publish(ctx, event)
	}

	return nil
}

func (s *ProfileCommandService) BatchAddTags(ctx context.Context, userID uint64, tags []*domain.UserTag) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		profile, err = s.CreateProfile(ctx, userID)
		if err != nil {
			return err
		}
	}

	for _, tag := range tags {
		tag.ProfileID = profile.ID
		tag.UserID = userID
		if err := profile.AddTag(tag); err != nil {
			continue
		}
		if err := s.tagRepo.Save(ctx, tag); err != nil {
			continue
		}
	}

	return s.profileRepo.Update(ctx, profile)
}

func (s *ProfileCommandService) RecordBehavior(ctx context.Context, userID uint64, behaviorType domain.BehaviorType, targetType string, targetID uint64, value string, duration int64) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		profile, err = s.CreateProfile(ctx, userID)
		if err != nil {
			return err
		}
	}

	features, err := s.behaviorRepo.FindByUserID(ctx, userID)
	if err != nil {
		features = domain.NewBehaviorFeatures(userID, profile.ID)
	}

	switch behaviorType {
	case domain.BehaviorTypeBrowse:
		features.RecordBrowse(targetID, 0, duration)
	case domain.BehaviorTypeSearch:
		features.RecordSearch(value)
	case domain.BehaviorTypePurchase:
		features.RecordPurchase(0, 0)
	case domain.BehaviorTypeAddToCart:
		features.RecordCart()
	case domain.BehaviorTypeShare:
		features.RecordShare()
	case domain.BehaviorTypeComment:
		features.RecordComment()
	case domain.BehaviorTypeReview:
		features.RecordReview()
	case domain.BehaviorTypeReturn:
		features.RecordReturn()
	case domain.BehaviorTypeCancel:
		features.RecordCancel()
	}

	recentBehavior := domain.NewRecentBehavior(userID, behaviorType, targetType, targetID)
	recentBehavior.Value = value
	recentBehavior.Duration = duration
	features.AddRecentBehavior(recentBehavior)

	if err := s.behaviorRepo.Save(ctx, features); err != nil {
		return fmt.Errorf("failed to save behavior features: %w", err)
	}

	profile.UpdateBehaviorFeatures(features)
	profile.RecordActivity(string(behaviorType))

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if s.eventPublisher != nil {
		event := domain.NewBehaviorRecordedEvent(profile.ID, userID, string(behaviorType), targetType, targetID)
		s.eventPublisher.Publish(ctx, event)
	}

	return nil
}

func (s *ProfileCommandService) RecalculateProfile(ctx context.Context, userID uint64) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		profile, err = s.CreateProfile(ctx, userID)
		if err != nil {
			return err
		}
	}

	if s.calculator != nil {
		behaviorFeatures, err := s.calculator.CalculateBehaviorFeatures(ctx, userID)
		if err == nil && behaviorFeatures != nil {
			profile.UpdateBehaviorFeatures(behaviorFeatures)
			s.behaviorRepo.Save(ctx, behaviorFeatures)
		}

		preferences, err := s.calculator.CalculatePreferences(ctx, userID)
		if err == nil && preferences != nil {
			profile.UpdatePreferences(preferences)
			s.preferenceRepo.Save(ctx, preferences)
		}

		consumption, err := s.calculator.CalculateConsumptionProfile(ctx, userID)
		if err == nil && consumption != nil {
			profile.UpdateConsumptionProfile(consumption)
			s.consumptionRepo.Save(ctx, consumption)
		}

		if err := s.calculator.CalculateScores(ctx, profile); err != nil {
			return fmt.Errorf("failed to calculate scores: %w", err)
		}
	}

	profile.CalculateOverallScore()
	profile.SetNextCalculateTime(time.Hour * 24)

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

func (s *ProfileCommandService) UpdateConsumption(ctx context.Context, userID uint64, amount int64, categoryID, brandID uint64, paymentMethod string, purchasedAt time.Time) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		profile, err = s.CreateProfile(ctx, userID)
		if err != nil {
			return err
		}
	}

	consumption, err := s.consumptionRepo.FindByUserID(ctx, userID)
	if err != nil {
		consumption = domain.NewConsumptionProfile(userID, profile.ID)
	}

	consumption.RecordPurchase(amount, categoryID, brandID, paymentMethod, purchasedAt)
	consumption.CalculateTrend()
	consumption.CalculateChurnRisk()
	consumption.DetermineValueSegment()

	if consumption.RFM == nil {
		consumption.RFM = domain.NewRFMScore(profile.ID, userID)
	}
	consumption.RFM.Calculate(
		int(time.Since(purchasedAt).Hours()/24),
		int(consumption.TotalOrders),
		consumption.TotalSpent,
	)

	if err := s.consumptionRepo.Save(ctx, consumption); err != nil {
		return fmt.Errorf("failed to save consumption profile: %w", err)
	}

	profile.UpdateConsumptionProfile(consumption)

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

func (s *ProfileCommandService) ApplyRules(ctx context.Context, userID uint64) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrProfileNotFound
	}

	rules, err := s.ruleRepo.FindEffective(ctx)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if rule.RuleType == domain.RuleTypeTagging {
			if s.evaluateRuleConditions(rule, profile) {
				for _, action := range rule.Actions {
					if action.ActionType == "ADD_TAG" {
						tag := domain.NewUserTag(profile.ID, userID, action.TargetField, action.Value, rule.Category, domain.TagSourceAlgorithm)
						profile.AddTag(tag)
						s.tagRepo.Save(ctx, tag)
					}
				}

				if s.eventPublisher != nil {
					event := domain.NewRuleTriggeredEvent(profile.ID, userID, rule.ID, rule.Name, rule.RuleType, "APPLIED", "SUCCESS")
					s.eventPublisher.Publish(ctx, event)
				}
			}
		}
	}

	return s.profileRepo.Update(ctx, profile)
}

func (s *ProfileCommandService) evaluateRuleConditions(rule *domain.ProfileRule, profile *domain.UserProfile) bool {
	for _, condition := range rule.Conditions {
		matched := s.evaluateCondition(condition, profile)
		if condition.LogicType == "AND" && !matched {
			return false
		}
		if condition.LogicType == "OR" && matched {
			return true
		}
	}
	return true
}

func (s *ProfileCommandService) evaluateCondition(condition *domain.RuleCondition, profile *domain.UserProfile) bool {
	switch condition.Field {
	case "overall_score":
		return s.compareValue(profile.OverallScore, condition.Operator, condition.Value)
	case "activity_score":
		return s.compareValue(profile.ActivityScore, condition.Operator, condition.Value)
	case "engagement_score":
		return s.compareValue(profile.EngagementScore, condition.Operator, condition.Value)
	default:
		tag := profile.GetTag(condition.Field)
		if tag != nil {
			return tag.TagValue == condition.Value
		}
	}
	return false
}

func (s *ProfileCommandService) compareValue(actual any, operator, expected string) bool {
	return true
}

func (s *ProfileCommandService) ArchiveProfile(ctx context.Context, userID uint64) error {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrProfileNotFound
	}

	profile.Archive()

	return s.profileRepo.Update(ctx, profile)
}

func (s *ProfileCommandService) DeleteProfile(ctx context.Context, userID uint64) error {
	return s.profileRepo.Delete(ctx, userID)
}
