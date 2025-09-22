package data

import (
	"context"
	"ecommerce/internal/auth/biz"
	"ecommerce/internal/auth/data/model"
	"time"

	"gorm.io/gorm"
)

type authRepo struct {
	data *Data
}

// NewAuthRepo creates a new AuthRepo.
func NewAuthRepo(data *Data) biz.AuthRepo {
	return &authRepo{data: data}
}

// CreateSession creates a new session record.
func (r *authRepo) CreateSession(ctx context.Context, session *biz.Session) (*biz.Session, error) {
	po := &model.Session{
		UserID:    session.UserID,
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	session.ID = po.ID
	return session, nil
}

// GetSessionByToken gets a session by token.
func (r *authRepo) GetSessionByToken(ctx context.Context, token string) (*biz.Session, error) {
	var po model.Session
	if err := r.data.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Session not found or expired
		}
		return nil, err
	}
	return &biz.Session{
		ID:        po.ID,
		UserID:    po.UserID,
		Token:     po.Token,
		ExpiresAt: po.ExpiresAt,
	}, nil
}

// DeleteSessionByToken deletes a session by token.
func (r *authRepo) DeleteSessionByToken(ctx context.Context, token string) error {
	return r.data.db.WithContext(ctx).Where("token = ?", token).Delete(&model.Session{}).Error
}

// CreateRefreshToken creates a new refresh token record.
func (r *authRepo) CreateRefreshToken(ctx context.Context, refreshToken *biz.RefreshToken) (*biz.RefreshToken, error) {
	po := &model.RefreshToken{
		UserID:    refreshToken.UserID,
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	refreshToken.ID = po.ID
	return refreshToken, nil
}

// GetRefreshTokenByToken gets a refresh token by token.
func (r *authRepo) GetRefreshTokenByToken(ctx context.Context, token string) (*biz.RefreshToken, error) {
	var po model.RefreshToken
	if err := r.data.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Refresh token not found or expired
		}
		return nil, err
	}
	return &biz.RefreshToken{
		ID:        po.ID,
		UserID:    po.UserID,
		Token:     po.Token,
		ExpiresAt: po.ExpiresAt,
	}, nil
}

// DeleteRefreshTokenByToken deletes a refresh token by token.
func (r *authRepo) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	return r.data.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
