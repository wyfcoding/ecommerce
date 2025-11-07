package utils

import "time"

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// ParseDate 解析日期字符串
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// StartOfDay 获取一天的开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}
