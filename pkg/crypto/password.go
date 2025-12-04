package crypto

import (
	"golang.org/x/crypto/bcrypt" // 导入bcrypt密码哈希库。
)

// HashPassword 对给定的明文密码生成 bcrypt 哈希值。
// bcrypt 是一种安全的密码哈希算法，适用于存储用户密码。
// password: 用户的明文密码。
// 返回值：密码的哈希字符串和可能发生的错误。
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword 会对密码进行哈希，并自动添加一个随机的盐值。
	// bcrypt.DefaultCost 是推荐的计算成本，可以根据系统性能和安全需求进行调整。
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证一个明文密码是否与给定的 bcrypt 哈希值匹配。
// 此函数不会解密哈希，而是对提供的明文密码进行哈希，然后比较两个哈希值。
// password: 用户的明文密码。
// hash: 存储的 bcrypt 哈希值。
// 返回值：如果密码匹配哈希，则返回 true；否则返回 false。
func CheckPassword(password, hash string) bool {
	// bcrypt.CompareHashAndPassword 负责比较明文密码和哈希值。
	// 如果匹配，返回 nil；否则返回错误。
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // 如果没有错误，则表示密码验证成功。
}
