package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not valid yet")
	ErrTokenInvalid     = errors.New("token is invalid")
)

// CustomClaims 定义了 JWT 的载荷
type MyCustomClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(userID uint64, username string, secretKey []byte, expires time.Duration) (string, int64, error) {
	// 创建 Claims
	expireTime := time.Now().Add(expires)
	claims := MyCustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "genesis-project",
		},
	}

	// 使用 HS256 签名算法创建一个新的 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用秘钥签名并获取完整的编码后的字符串 token
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireTime.Unix(), nil
}

// ParseToken 解析 JWT
func ParseToken(tokenString string, secretKey []byte) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
