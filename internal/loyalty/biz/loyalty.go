package biz

import (
	"context"
	"errors"
	"time"
)

// ErrUserLoyaltyProfileNotFound is a specific error for when a user's loyalty profile is not found.
var ErrUserLoyaltyProfileNotFound = errors.New("user loyalty profile not found")

// ErrInsufficientPoints is a specific error for when a user tries to deduct more points than they have.
var ErrInsufficientPoints = errors.New("insufficient points")

// UserLoyaltyProfile represents a user's loyalty profile in the business layer.
type UserLoyaltyProfile struct {
	UserID          string
	CurrentPoints   int64
	LoyaltyLevel    string
	LastLevelUpdate time.Time
}

// PointsTransaction represents a points transaction in the business layer.
type PointsTransaction struct {
	ID           uint
	UserID       string
	PointsChange int64
	Reason       string
	OrderID      string
	CreatedAt    time.Time
}

// LoyaltyRepo defines the data storage interface for loyalty data.
// The business layer depends on this interface, not on a concrete data implementation.
type LoyaltyRepo interface {
	GetUserLoyaltyProfile(ctx context.Context, userID string) (*UserLoyaltyProfile, error)
	CreateUserLoyaltyProfile(ctx context.Context, profile *UserLoyaltyProfile) (*UserLoyaltyProfile, error)
	UpdateUserLoyaltyProfile(ctx context.Context, profile *UserLoyaltyProfile) (*UserLoyaltyProfile, error)
	AddPointsTransaction(ctx context.Context, transaction *PointsTransaction) (*PointsTransaction, error)
	ListPointsTransactions(ctx context.Context, userID string, pageSize, pageToken int32) ([]*PointsTransaction, int32, error)
}

// LoyaltyUsecase is the use case for loyalty-related operations.
// It orchestrates the business logic.
type LoyaltyUsecase struct {
	repo LoyaltyRepo
	// You can also inject other dependencies like a logger
}

// NewLoyaltyUsecase creates a new LoyaltyUsecase.
func NewLoyaltyUsecase(repo LoyaltyRepo) *LoyaltyUsecase {
	return &LoyaltyUsecase{repo: repo}
}

// GetUserLoyaltyProfile retrieves a user's loyalty profile.
func (uc *LoyaltyUsecase) GetUserLoyaltyProfile(ctx context.Context, userID string) (*UserLoyaltyProfile, error) {
	profile, err := uc.repo.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserLoyaltyProfileNotFound) {
			// If not found, create a default one
			defaultProfile := &UserLoyaltyProfile{
				UserID:        userID,
				CurrentPoints: 0,
				LoyaltyLevel:  "Bronze",
				LastLevelUpdate: time.Now(),
			}
			return uc.repo.CreateUserLoyaltyProfile(ctx, defaultProfile)
		}
		return nil, err
	}
	return profile, nil
}

// AddPoints adds points to a user's balance and records a transaction.
func (uc *LoyaltyUsecase) AddPoints(ctx context.Context, userID string, points int64, reason, orderID string) (*UserLoyaltyProfile, *PointsTransaction, error) {
	profile, err := uc.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	profile.CurrentPoints += points
	updatedProfile, err := uc.repo.UpdateUserLoyaltyProfile(ctx, profile)
	if err != nil {
		return nil, nil, err
	}

	transaction := &PointsTransaction{
		UserID:       userID,
		PointsChange: points,
		Reason:       reason,
		OrderID:      orderID,
		CreatedAt:    time.Now(),
	}
	addedTransaction, err := uc.repo.AddPointsTransaction(ctx, transaction)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Logic to update loyalty level based on new points

	return updatedProfile, addedTransaction, nil
}

// DeductPoints deducts points from a user's balance and records a transaction.
func (uc *LoyaltyUsecase) DeductPoints(ctx context.Context, userID string, points int64, reason, orderID string) (*UserLoyaltyProfile, *PointsTransaction, error) {
	profile, err := uc.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	if profile.CurrentPoints < points {
		return nil, nil, ErrInsufficientPoints
	}

	profile.CurrentPoints -= points
	updatedProfile, err := uc.repo.UpdateUserLoyaltyProfile(ctx, profile)
	if err != nil {
		return nil, nil, err
	}

	transaction := &PointsTransaction{
		UserID:       userID,
		PointsChange: -points,
		Reason:       reason,
		OrderID:      orderID,
		CreatedAt:    time.Now(),
	}
	addedTransaction, err := uc.repo.AddPointsTransaction(ctx, transaction)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Logic to update loyalty level based on new points

	return updatedProfile, addedTransaction, nil
}

// ListPointsTransactions lists a user's points transaction history.
func (uc *LoyaltyUsecase) ListPointsTransactions(ctx context.Context, userID string, pageSize, pageToken int32) ([]*PointsTransaction, int32, error) {
	return uc.repo.ListPointsTransactions(ctx, userID, pageSize, pageToken)
}

// UpdateUserLevel updates a user's loyalty level.
func (uc *LoyaltyUsecase) UpdateUserLevel(ctx context.Context, userID, newLevel, reason string) (*UserLoyaltyProfile, error) {
	profile, err := uc.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.LoyaltyLevel = newLevel
	profile.LastLevelUpdate = time.Now()
	return uc.repo.UpdateUserLoyaltyProfile(ctx, profile)
}
