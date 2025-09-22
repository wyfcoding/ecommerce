package data

import (
	"context"
	"ecommerce/internal/user/biz"
	"ecommerce/internal/user/data/model"
)

// userRepo 是 biz.UserRepo 接口的具体实现。
type userRepo struct {
	*Data
}

// NewUserRepo 是 userRepo 的构造函数。
func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{
		Data: data,
	}
}

// toBizUser 将数据库模型 data.User 转换为业务领域模型 biz.User。
func (r *userRepo) toBizUser(user *model.User) *biz.User {
	if user == nil {
		return nil
	}
	return &biz.User{
		ID:        user.ID,
		Username:  user.Username,
		Password:  user.Password,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Phone:     user.Phone,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// CreateUser 实现了 biz.UserRepo 接口的 CreateUser 方法。
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &model.User{
		Username: u.Username,
		Password: u.Password,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Gender:   u.Gender,
		Birthday: u.Birthday,
		Phone:    u.Phone,
		Email:    u.Email,
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	u.ID = user.ID // Assign the generated ID back to biz.User
	return r.toBizUser(user), nil
}

// GetUserByUsername 实现了 biz.UserRepo 接口的 GetUserByUsername 方法。
func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

// GetUserByUserID 实现了 biz.UserRepo 接口的 GetUserByUserID 方法。
func (r *userRepo) GetUserByUserID(ctx context.Context, userID uint64) (*biz.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

// UpdateUser 实现了 biz.UserRepo 接口的 UpdateUser 方法。
func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", u.ID).First(&user).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if u.Nickname != "" {
		updates["nickname"] = u.Nickname
	}
	if u.Avatar != "" {
		updates["avatar"] = u.Avatar
	}
	if u.Gender != 0 {
		updates["gender"] = u.Gender
	}
	if !u.Birthday.IsZero() {
		updates["birthday"] = u.Birthday
	}
	if u.Phone != "" {
		updates["phone"] = u.Phone
	}
	if u.Email != "" {
		updates["email"] = u.Email
	}

	if len(updates) == 0 {
		return r.toBizUser(&user), nil
	}

	if err := r.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.toBizUser(&user), nil
}
