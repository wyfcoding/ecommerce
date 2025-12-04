// Package hash 提供了密码哈希和验证功能。
// 主要使用 bcrypt 算法来安全地存储和比较用户密码。
package hash

import "golang.org/x/crypto/bcrypt"

// HashPassword 对给定的明文密码生成 bcrypt 哈希值。
// bcrypt 是一种专门为密码存储设计的哈希算法，它通过迭代计算增加哈希的成本，
// 以抵御彩虹表攻击和暴力破解。
//
// password: 用户提供的明文密码。
// 返回: 密码的哈希字符串和可能发生的错误。
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost 是推荐的默认计算成本，可以根据CPU性能进行调整。
	// 成本越高，哈希越安全，但计算耗时越长。
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 比较一个明文密码和它对应的 bcrypt 哈希值。
// 此函数不会解密哈希，而是对明文密码进行哈希，然后比较两个哈希值是否匹配。
//
// password: 用户输入的明文密码。
// hash: 存储在数据库中的bcrypt哈希值。
// 返回: 如果密码匹配哈希，则返回 true；否则返回 false。
func CheckPasswordHash(password, hash string) bool {
	// bcrypt.CompareHashAndPassword 函数会处理哈希和明文的比较，
	// 如果匹配则返回 nil，否则返回错误。
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // 如果没有错误，则表示密码匹配。
}
