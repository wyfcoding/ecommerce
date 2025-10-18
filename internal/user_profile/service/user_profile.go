package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/user_profile/model"
	"ecommerce/internal/user_profile/repository"
)

var (
	ErrUserProfileNotFound = errors.New("user profile not found")
)

// UserProfileService is the business logic for user profile management.
type UserProfileService struct {
	repo repository.UserProfileRepo
}

// NewUserProfileService creates a new UserProfileService.
func NewUserProfileService(repo repository.UserProfileRepo) *UserProfileService {
	return &UserProfileService{repo: repo}
}

// GetUserProfile retrieves a user profile.
func (s *UserProfileService) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	profile, err := s.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrUserProfileNotFound
	}
	return profile, nil
}

// UpdateUserProfile updates a user profile.
func (s *UserProfileService) UpdateUserProfile(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	// In a real system, you might fetch the existing profile first,
	// merge changes, and then update.
	profile.UpdatedAt = time.Now()
	return s.repo.UpdateUserProfile(ctx, profile)
}

// RecordUserBehavior records a user behavior event.
func (s *UserProfileService) RecordUserBehavior(ctx context.Context, userID, behaviorType, itemID string, properties map[string]string, eventTime time.Time) error {
	event := &model.UserBehaviorEvent{
		UserID:       userID,
		BehaviorType: behaviorType,
		ItemID:       itemID,
		Properties:   properties,
		EventTime:    eventTime,
	}
	return s.repo.RecordUserBehavior(ctx, event)
}
