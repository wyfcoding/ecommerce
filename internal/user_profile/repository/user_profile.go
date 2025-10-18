package repository

import (
	"context"

	"ecommerce/internal/user_profile/model"
)

// UserProfileRepo defines the interface for user profile data access.
type UserProfileRepo interface {
	GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
	RecordUserBehavior(ctx context.Context, event *model.UserBehaviorEvent) error
}
