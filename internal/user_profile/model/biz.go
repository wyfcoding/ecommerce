package biz

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserProfileNotFound = errors.New("user profile not found")
)

// UserProfile represents a user's profile in the business logic layer.
type UserProfile struct {
	UserID           string
	Gender           string
	AgeGroup         string
	City             string
	Interests        []string
	RecentCategories []string
	RecentBrands     []string
	TotalSpent       uint64
	OrderCount       uint32
	LastActiveTime   time.Time
	CustomTags       map[string]string
	UpdatedAt        time.Time
}

// UserBehaviorEvent represents a user behavior event in the business logic layer.
type UserBehaviorEvent struct {
	UserID       string
	BehaviorType string
	ItemID       string
	Properties   map[string]string
	EventTime    time.Time
}

// UserProfileRepo defines the interface for user profile data access.
type UserProfileRepo interface {
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *UserProfile) (*UserProfile, error)
	RecordUserBehavior(ctx context.Context, event *UserBehaviorEvent) error
}

// UserProfileUsecase is the business logic for user profile management.
type UserProfileUsecase struct {
	repo UserProfileRepo
}

// NewUserProfileUsecase creates a new UserProfileUsecase.
func NewUserProfileUsecase(repo UserProfileRepo) *UserProfileUsecase {
	return &UserProfileUsecase{repo: repo}
}

// GetUserProfile retrieves a user profile.
func (uc *UserProfileUsecase) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	profile, err := uc.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrUserProfileNotFound
	}
	return profile, nil
}

// UpdateUserProfile updates a user profile.
func (uc *UserProfileUsecase) UpdateUserProfile(ctx context.Context, profile *UserProfile) (*UserProfile, error) {
	// In a real system, you might fetch the existing profile first,
	// merge changes, and then update.
	return uc.repo.UpdateUserProfile(ctx, profile)
}

// RecordUserBehavior records a user behavior event.
func (uc *UserProfileUsecase) RecordUserBehavior(ctx context.Context, userID, behaviorType, itemID string, properties map[string]string, eventTime time.Time) error {
	event := &UserBehaviorEvent{
		UserID:       userID,
		BehaviorType: behaviorType,
		ItemID:       itemID,
		Properties:   properties,
		EventTime:    eventTime,
	}
	return uc.repo.RecordUserBehavior(ctx, event)
}
