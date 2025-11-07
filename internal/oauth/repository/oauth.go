package repository

import (
	"context"

	"gorm.io/gorm"

	"ecommerce/internal/oauth/model"
)

// OAuthRepo OAuth仓储接口
type OAuthRepo interface {
	// 用户OAuth
	CreateUserOAuth(ctx context.Context, userOAuth *model.UserOAuth) error
	UpdateUserOAuth(ctx context.Context, userOAuth *model.UserOAuth) error
	GetByUserIDAndProvider(ctx context.Context, userID uint64, provider model.OAuthProvider) (*model.UserOAuth, error)
	GetByProviderAndOpenID(ctx context.Context, provider model.OAuthProvider, openID string) (*model.UserOAuth, error)
	ListByUserID(ctx context.Context, userID uint64) ([]*model.UserOAuth, error)
	DeleteUserOAuth(ctx context.Context, id uint64) error
	
	// OAuth State
	CreateOAuthState(ctx context.Context, state *model.OAuthState) error
	GetOAuthStateByState(ctx context.Context, state string) (*model.OAuthState, error)
	DeleteOAuthState(ctx context.Context, id uint64) error
	DeleteExpiredStates(ctx context.Context) error
}

type oauthRepo struct {
	db *gorm.DB
}

// NewOAuthRepo 创建OAuth仓储实例
func NewOAuthRepo(db *gorm.DB) OAuthRepo {
	return &oauthRepo{db: db}
}

// CreateUserOAuth 创建用户OAuth绑定
func (r *oauthRepo) CreateUserOAuth(ctx context.Context, userOAuth *model.UserOAuth) error {
	return r.db.WithContext(ctx).Create(userOAuth).Error
}

// UpdateUserOAuth 更新用户OAuth绑定
func (r *oauthRepo) UpdateUserOAuth(ctx context.Context, userOAuth *model.UserOAuth) error {
	return r.db.WithContext(ctx).Save(userOAuth).Error
}

// GetByUserIDAndProvider 根据用户ID和提供商获取OAuth绑定
func (r *oauthRepo) GetByUserIDAndProvider(ctx context.Context, userID uint64, provider model.OAuthProvider) (*model.UserOAuth, error) {
	var userOAuth model.UserOAuth
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&userOAuth).Error
	if err != nil {
		return nil, err
	}
	return &userOAuth, nil
}

// GetByProviderAndOpenID 根据提供商和OpenID获取OAuth绑定
func (r *oauthRepo) GetByProviderAndOpenID(ctx context.Context, provider model.OAuthProvider, openID string) (*model.UserOAuth, error) {
	var userOAuth model.UserOAuth
	err := r.db.WithContext(ctx).
		Where("provider = ? AND open_id = ?", provider, openID).
		First(&userOAuth).Error
	if err != nil {
		return nil, err
	}
	return &userOAuth, nil
}

// ListByUserID 根据用户ID获取所有OAuth绑定
func (r *oauthRepo) ListByUserID(ctx context.Context, userID uint64) ([]*model.UserOAuth, error) {
	var userOAuths []*model.UserOAuth
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&userOAuths).Error
	if err != nil {
		return nil, err
	}
	return userOAuths, nil
}

// DeleteUserOAuth 删除用户OAuth绑定
func (r *oauthRepo) DeleteUserOAuth(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.UserOAuth{}, id).Error
}

// CreateOAuthState 创建OAuth状态
func (r *oauthRepo) CreateOAuthState(ctx context.Context, state *model.OAuthState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

// GetOAuthStateByState 根据state获取OAuth状态
func (r *oauthRepo) GetOAuthStateByState(ctx context.Context, state string) (*model.OAuthState, error) {
	var oauthState model.OAuthState
	err := r.db.WithContext(ctx).Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		return nil, err
	}
	return &oauthState, nil
}

// DeleteOAuthState 删除OAuth状态
func (r *oauthRepo) DeleteOAuthState(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.OAuthState{}, id).Error
}

// DeleteExpiredStates 删除过期的OAuth状态
func (r *oauthRepo) DeleteExpiredStates(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW()").
		Delete(&model.OAuthState{}).Error
}
