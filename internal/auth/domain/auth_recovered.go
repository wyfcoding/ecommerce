//go:build ignore

package domain

import (
	"context"
	"errors"

	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/security"
	"gorm.io/gorm"
)

// Auth 认证聚合根（救回版本：包含密码校验、锁定逻辑）
type Auth struct {
	gorm.Model
	database.BaseEntity
	UserID           uint64 `gorm:"column:user_id;uniqueIndex;not null"`
	Username         string `gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	PasswordHash     string `gorm:"column:password_hash;type:varchar(255);not null"`
	Status           int    `gorm:"column:status;type:tinyint;default:0"`
	LoginFailedCount int    `gorm:"column:failed_count;default:0"`
}

func (a *Auth) CheckPassword(password string) error {
	if a.Status != 0 {
		return errors.New("account locked")
	}
	if !security.CheckPassword(password, a.PasswordHash) {
		a.LoginFailedCount++
		return errors.New("invalid password")
	}
	a.LoginFailedCount = 0
	return nil
}

type Repository interface {
	Save(ctx context.Context, auth *Auth) error
	FindByUsername(ctx context.Context, username string) (*Auth, error)
}
