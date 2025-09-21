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
		UserID:   user.UserID,
		Username: user.Username,
		Password: user.Password,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
	}
}

// CreateUser 实现了 biz.UserRepo 接口的 CreateUser 方法。
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &model.User{
		UserID:   u.UserID,
		Username: u.Username,
		Password: u.Password,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Gender:   u.Gender,
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
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
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

// UpdateUser 实现了 biz.UserRepo 接口的 UpdateUser 方法。
func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("user_id = ?", u.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	// 使用指针类型后，我们可以精确地判断哪些字段需要更新。
	updates := make(map[string]interface{})
	if u.Nickname != nil {
		updates["nickname"] = u.Nickname
	}
	if u.Avatar != nil {
		updates["avatar"] = u.Avatar
	}
	if u.Gender != nil {
		updates["gender"] = u.Gender
	}

	if len(updates) == 0 {
		return r.toBizUser(&user), nil
	}

	if err := r.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.toBizUser(&user), nil
}
