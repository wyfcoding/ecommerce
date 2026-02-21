package domain

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/security"
)

// Auth 认证聚合根：物理实装功能完整版。
type Auth struct {
	gorm.Model
	database.BaseEntity
	UserID          uint64        `gorm:"column:user_id;uniqueIndex;not null"`
	Username        string        `gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	PasswordHash    string        `gorm:"column:password_hash;type:varchar(255);not null"`
	Status          int           `gorm:"column:status;type:tinyint;default:0"`
	LoginFailedCount int          `gorm:"column:failed_count;default:0"`
	LastLoginAt     *time.Time    `gorm:"column:last_login_at"`
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
