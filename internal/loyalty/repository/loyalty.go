package repository

import (
	"context"

	"ecommerce/internal/loyalty/model"
)

// LoyaltyRepo defines the data storage interface for loyalty data.
// The business layer depends on this interface, not on a concrete data implementation.
type LoyaltyRepo interface {
	GetUserLoyaltyProfile(ctx context.Context, userID string) (*model.UserLoyaltyProfile, error)
	CreateUserLoyaltyProfile(ctx context.Context, profile *model.UserLoyaltyProfile) (*model.UserLoyaltyProfile, error)
	UpdateUserLoyaltyProfile(ctx context.Context, profile *model.UserLoyaltyProfile) (*model.UserLoyaltyProfile, error)
	AddPointsTransaction(ctx context.Context, transaction *model.PointsTransaction) (*model.PointsTransaction, error)
	ListPointsTransactions(ctx context.Context, userID string, pageSize, pageToken int32) ([]*model.PointsTransaction, int32, error)
}