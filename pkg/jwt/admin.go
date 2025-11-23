package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAdminToken 生成管理员JWT token
func GenerateAdminToken(adminID uint64, username, email, secret, issuer string, expireSeconds int64) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": adminID,
		"username": username,
		"email":    email,
		"iss":      issuer,
		"exp":      time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
