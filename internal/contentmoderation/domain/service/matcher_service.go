package service

// SensitiveMatcher 定义敏感词匹配接口
type SensitiveMatcher interface {
	// ContainsSensitivity 检查是否包含敏感词
	ContainsSensitivity(text string) bool
	// Reload 加载敏感词库
	Reload(keywords []string) error
}
