package data

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a user session.
type Session struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"userId"`
	Token     string    `gorm:"uniqueIndex;not null;comment:会话令牌" json:"token"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expiresAt"`
	// Add other session-related fields like IP address, user agent, etc.
}

// RefreshToken represents a refresh token.
type RefreshToken struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"userId"`
	Token     string    `gorm:"uniqueIndex;not null;comment:刷新令牌" json:"token"`
	ExpiresAt time.Time `gorm:"not null;comment:过期时间" json:"expiresAt"`
}

// TableName specifies the table name for Session.
func (Session) TableName() string {
	return "sessions"
}

// TableName specifies the table name for RefreshToken.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
