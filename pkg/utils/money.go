package utils

import "fmt"

// FenToYuan 分转元
func FenToYuan(fen int64) float64 {
	return float64(fen) / 100.0
}

// YuanToFen 元转分
func YuanToFen(yuan float64) int64 {
	return int64(yuan * 100)
}

// FormatMoney 格式化金额（分）为字符串
func FormatMoney(fen int64) string {
	yuan := FenToYuan(fen)
	return fmt.Sprintf("%.2f", yuan)
}
