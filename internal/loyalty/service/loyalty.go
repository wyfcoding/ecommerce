package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/loyalty/model"
	"ecommerce/internal/loyalty/repository"
)

// ErrUserLoyaltyProfileNotFound is a specific error for when a user's loyalty profile is not found.
var ErrUserLoyaltyProfileNotFound = errors.New("user loyalty profile not found")

// ErrInsufficientPoints is a specific error for when a user tries to deduct more points than they have.
var ErrInsufficientPoints = errors.New("insufficient points")

// LoyaltyService is the use case for loyalty-related operations.
// It orchestrates the business logic.
type LoyaltyService struct {
	repo repository.LoyaltyRepo
	// You can also inject other dependencies like a logger
}

// NewLoyaltyService creates a new LoyaltyService.
func NewLoyaltyService(repo repository.LoyaltyRepo) *LoyaltyService {
	return &LoyaltyService{repo: repo}
}

// GetUserLoyaltyProfile retrieves a user's loyalty profile.
func (s *LoyaltyService) GetUserLoyaltyProfile(ctx context.Context, userID string) (*model.UserLoyaltyProfile, error) {
	profile, err := s.repo.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserLoyaltyProfileNotFound) {
			// If not found, create a default one
			defaultProfile := &model.UserLoyaltyProfile{
				UserID:          userID,
				CurrentPoints:   0,
				LoyaltyLevel:    "Bronze",
				LastLevelUpdate: time.Now(),
			}
			return s.repo.CreateUserLoyaltyProfile(ctx, defaultProfile)
		}
		return nil, err
	}
	return profile, nil
}

// AddPoints adds points to a user's balance and records a transaction.
func (s *LoyaltyService) AddPoints(ctx context.Context, userID string, points int64, reason, orderID string) (*model.UserLoyaltyProfile, *model.PointsTransaction, error) {
	profile, err := s.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	profile.CurrentPoints += points
	updatedProfile, err := s.repo.UpdateUserLoyaltyProfile(ctx, profile)
	if err != nil {
		return nil, nil, err
	}

	transaction := &model.PointsTransaction{
		UserID:       userID,
		PointsChange: points,
		Reason:       reason,
		OrderID:      orderID,
		CreatedAt:    time.Now(),
	}
	addedTransaction, err := s.repo.AddPointsTransaction(ctx, transaction)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Logic to update loyalty level based on new points

	return updatedProfile, addedTransaction, nil
}

// DeductPoints deducts points from a user's balance and records a transaction.
func (s *LoyaltyService) DeductPoints(ctx context.Context, userID string, points int64, reason, orderID string) (*model.UserLoyaltyProfile, *model.PointsTransaction, error) {
	profile, err := s.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	if profile.CurrentPoints < points {
		return nil, nil, ErrInsufficientPoints
	}

	profile.CurrentPoints -= points
	updatedProfile, err := s.repo.UpdateUserLoyaltyProfile(ctx, profile)
	if err != nil {
		return nil, nil, err
	}

	transaction := &model.PointsTransaction{
		UserID:       userID,
		PointsChange: -points,
		Reason:       reason,
		OrderID:      orderID,
		CreatedAt:    time.Now(),
	}
	addedTransaction, err := s.repo.AddPointsTransaction(ctx, transaction)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Logic to update loyalty level based on new points

	return updatedProfile, addedTransaction, nil
}

// ListPointsTransactions lists a user's points transaction history.
func (s *LoyaltyService) ListPointsTransactions(ctx context.Context, userID string, pageSize, pageToken int32) ([]*model.PointsTransaction, int32, error) {
	return s.repo.ListPointsTransactions(ctx, userID, pageSize, pageToken)
}

// UpdateUserLevel updates a user's loyalty level.
func (s *LoyaltyService) UpdateUserLevel(ctx context.Context, userID, newLevel, reason string) (*model.UserLoyaltyProfile, error) {
	profile, err := s.GetUserLoyaltyProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.LoyaltyLevel = newLevel
	profile.LastLevelUpdate = time.Now()
	return s.repo.UpdateUserLoyaltyProfile(ctx, profile)
}
