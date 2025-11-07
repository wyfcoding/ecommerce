package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义标准的JWT错误
var (
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not valid yet")
	ErrTokenInvalid     = errors.New("token is invalid")
)

// MyCustomClaims 定义了JWT的自定义载荷 (Payload)
type MyCustomClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成一个JWT
func GenerateToken(userID uint64, username, secretKey, issuer string, expires time.Duration, method jwt.SigningMethod) (string, error) {
	// 如果没有指定签名方法，使用默认的HS256
	if method == nil {
		method = jwt.SigningMethodHS256
	}

	// 创建 Claims
	expireTime := time.Now().Add(expires)
	claims := MyCustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	// 使用提供的签名算法创建一个新的 Token 对象
	token := jwt.NewWithClaims(method, claims)

	// 使用提供的密钥签名并获取完整的编码后的字符串 token
	return token.SignedString([]byte(secretKey))
}

// ParseToken 解析JWT字符串
func ParseToken(tokenString string, secretKey string) (*MyCustomClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	// 处理可能发生的错误
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenExpired
		} else {
			return nil, ErrTokenInvalid
		}
	}

	// 校验token并返回自定义的Claims
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}
