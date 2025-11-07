package validator

import (
	"regexp"
	"strings"
)

// 常用正则表达式
var (
	phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{4,20}$`)
	passwordRegex = regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$`)
)

// IsValidPhone 验证手机号
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsValidEmail 验证邮箱
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidUsername 验证用户名
func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// IsValidPassword 验证密码强度
func IsValidPassword(password string) bool {
	return passwordRegex.MatchString(password)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsValidLength 检查字符串长度
func IsValidLength(s string, min, max int) bool {
	length := len([]rune(s))
	return length >= min && length <= max
}

// IsPositive 检查数字是否为正数
func IsPositive(n int64) bool {
	return n > 0
}

// IsNonNegative 检查数字是否为非负数
func IsNonNegative(n int64) bool {
	return n >= 0
}

// IsInRange 检查数字是否在范围内
func IsInRange(n, min, max int64) bool {
	return n >= min && n <= max
}
