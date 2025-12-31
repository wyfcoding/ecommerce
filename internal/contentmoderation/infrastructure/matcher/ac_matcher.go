package matcher

import (
	"sync"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain/service"
	"github.com/wyfcoding/pkg/algorithm"
)

// ACMatcher 基于 Aho-Corasick 自动机的敏感词匹配器
// 适用于在一长段文本中高效查找多个模式串
type ACMatcher struct {
	ac *algorithm.AhoCorasick
	mu sync.RWMutex
}

// NewACMatcher 创建一个新的 ACMatcher
func NewACMatcher() *ACMatcher {
	return &ACMatcher{
		ac: algorithm.NewAhoCorasick(),
	}
}

// ContainsSensitivity 检查文本是否包含任何敏感词
// AC 自动机支持 O(n) 时间复杂度的文本扫描
func (am *ACMatcher) ContainsSensitivity(text string) bool {
	return am.ac.Contains(text)
}

// Reload 重新构建 AC 自动机
func (am *ACMatcher) Reload(keywords []string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	newAc := algorithm.NewAhoCorasick()
	newAc.AddPatterns(keywords...)
	newAc.Build()

	am.ac = newAc
	return nil
}

var _ service.SensitiveMatcher = (*ACMatcher)(nil)
