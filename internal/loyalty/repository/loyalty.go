package data

import (
	"context"
	"ecommerce/internal/loyalty/biz"

	"gorm.io/gorm"
)

// loyaltyRepo is the data layer implementation for LoyaltyRepo.
type loyaltyRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.UserLoyaltyProfile model to a biz.UserLoyaltyProfile entity.
func (p *UserLoyaltyProfile) toBiz() *biz.UserLoyaltyProfile {
	if p == nil {
		return nil
	}
	return &biz.UserLoyaltyProfile{
		UserID:          p.UserID,
		CurrentPoints:   p.CurrentPoints,
		LoyaltyLevel:    p.LoyaltyLevel,
		LastLevelUpdate: p.LastLevelUpdate,
	}
}

// fromBiz converts a biz.UserLoyaltyProfile entity to a data.UserLoyaltyProfile model.
func fromBizUserLoyaltyProfile(b *biz.UserLoyaltyProfile) *UserLoyaltyProfile {
	if b == nil {
		return nil
	}
	return &UserLoyaltyProfile{
		UserID:          b.UserID,
		CurrentPoints:   b.CurrentPoints,
		LoyaltyLevel:    b.LoyaltyLevel,
		LastLevelUpdate: b.LastLevelUpdate,
	}
}

// toBiz converts a data.PointsTransaction model to a biz.PointsTransaction entity.
func (t *PointsTransaction) toBiz() *biz.PointsTransaction {
	if t == nil {
		return nil
	}
	return &biz.PointsTransaction{
		ID:           t.ID,
		UserID:       t.UserID,
		PointsChange: t.PointsChange,
		Reason:       t.Reason,
		OrderID:      t.OrderID,
		CreatedAt:    t.CreatedAt,
	}
}

// fromBiz converts a biz.PointsTransaction entity to a data.PointsTransaction model.
func fromBizPointsTransaction(b *biz.PointsTransaction) *PointsTransaction {
	if b == nil {
		return nil
	}
	return &PointsTransaction{
		UserID:       b.UserID,
		PointsChange: b.PointsChange,
		Reason:       b.Reason,
		OrderID:      b.OrderID,
		CreatedAt:    b.CreatedAt,
	}
}

// GetUserLoyaltyProfile retrieves a user's loyalty profile from the database.
func (r *loyaltyRepo) GetUserLoyaltyProfile(ctx context.Context, userID string) (*biz.UserLoyaltyProfile, error) {
	var profile UserLoyaltyProfile
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrUserLoyaltyProfileNotFound
		}
		return nil, err
	}
	return profile.toBiz(), nil
}

// CreateUserLoyaltyProfile creates a new user loyalty profile in the database.
func (r *loyaltyRepo) CreateUserLoyaltyProfile(ctx context.Context, b *biz.UserLoyaltyProfile) (*biz.UserLoyaltyProfile, error) {
	profile := fromBizUserLoyaltyProfile(b)
	if err := r.data.db.WithContext(ctx).Create(profile).Error; err != nil {
		return nil, err
	}
	return profile.toBiz(), nil
}

// UpdateUserLoyaltyProfile updates an existing user loyalty profile in the database.
func (r *loyaltyRepo) UpdateUserLoyaltyProfile(ctx context.Context, b *biz.UserLoyaltyProfile) (*biz.UserLoyaltyProfile, error) {
	profile := fromBizUserLoyaltyProfile(b)
	if err := r.data.db.WithContext(ctx).Save(profile).Error; err != nil {
		return nil, err
	}
	return profile.toBiz(), nil
}

// AddPointsTransaction adds a new points transaction to the database.
func (r *loyaltyRepo) AddPointsTransaction(ctx context.Context, b *biz.PointsTransaction) (*biz.PointsTransaction, error) {
	transaction := fromBizPointsTransaction(b)
	if err := r.data.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return nil, err
	}
	return transaction.toBiz(), nil
}

// ListPointsTransactions lists a user's points transaction history from the database.
func (r *loyaltyRepo) ListPointsTransactions(ctx context.Context, userID string, pageSize, pageToken int32) ([]*biz.PointsTransaction, int32, error) {
	var transactions []PointsTransaction
	var totalCount int32

	query := r.data.db.WithContext(ctx).Where("user_id = ?", userID)

	// Get total count
	query.Model(&PointsTransaction{}).Count(int64(&totalCount))

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(int(pageSize)).Offset(int(pageToken * pageSize))
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	bizTransactions := make([]*biz.PointsTransaction, len(transactions))
	for i, tx := range transactions {
		bizTransactions[i] = tx.toBiz()
	}

	return bizTransactions, totalCount, nil
}
