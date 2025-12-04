package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5" // 导入JWT库，版本v5。
)

// GenerateAdminToken 生成一个用于管理员用户的JWT (JSON Web Token)。
// 这个token包含了管理员的基本信息，并设置了签发者和过期时间，用于认证和授权。
// adminID: 管理员用户的唯一ID。
// username: 管理员的用户名。
// email: 管理员的邮箱地址。
// secret: 用于签名JWT的密钥。
// issuer: JWT的签发者（例如，"your-service-name"）。
// expireSeconds: JWT的过期时间，以秒为单位。
// 返回值：生成的JWT字符串和可能发生的错误。
func GenerateAdminToken(adminID uint64, username, email, secret, issuer string, expireSeconds int64) (string, error) {
	// 创建一个jwt.MapClaims，用于存储JWT的Payload（有效载荷）。
	// MapClaims 是一种通用的Map类型，可以存储自定义的Claim。
	claims := jwt.MapClaims{
		"admin_id": adminID,                                                           // 管理员ID，自定义Claim。
		"username": username,                                                          // 用户名，自定义Claim。
		"email":    email,                                                             // 邮箱，自定义Claim。
		"iss":      issuer,                                                            // 签发者 (Issuer)，标准Claim。
		"exp":      time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix(), // 过期时间 (Expiration Time)，标准Claim。
		"iat":      time.Now().Unix(),                                                 // 签发时间 (Issued At)，标准Claim。
	}

	// 使用HS256签名方法和自定义Claims创建一个新的Token。
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用秘密密钥对Token进行签名，并返回签名的JWT字符串。
	return token.SignedString([]byte(secret))
}
