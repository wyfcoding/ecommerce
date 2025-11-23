package algorithm

import (
	"sort"
	"sync"
)

// ============================================================================
// 7. 后缀数组 (Suffix Array) - 搜索优化
// ============================================================================

// SuffixArray 后缀数组
type SuffixArray struct {
	text string
	sa   []int
	rank []int
	mu   sync.RWMutex
}

// NewSuffixArray 创建后缀数组
func NewSuffixArray(text string) *SuffixArray {
	sa := &SuffixArray{
		text: text,
		sa:   make([]int, len(text)),
		rank: make([]int, len(text)),
	}
	sa.build()
	return sa
}

// build 构建后缀数组
func (sa *SuffixArray) build() {
	n := len(sa.text)
	for i := 0; i < n; i++ {
		sa.sa[i] = i
		sa.rank[i] = int(sa.text[i])
	}

	// 倍增法构建
	for k := 1; k < n; k *= 2 {
		sort.Slice(sa.sa, func(i, j int) bool {
			a, b := sa.sa[i], sa.sa[j]
			if sa.rank[a] != sa.rank[b] {
				return sa.rank[a] < sa.rank[b]
			}
			ra := 0
			rb := 0
			if a+k < n {
				ra = sa.rank[a+k]
			}
			if b+k < n {
				rb = sa.rank[b+k]
			}
			return ra < rb
		})

		newRank := make([]int, n)
		newRank[sa.sa[0]] = 0
		for i := 1; i < n; i++ {
			newRank[sa.sa[i]] = newRank[sa.sa[i-1]]
			a, b := sa.sa[i-1], sa.sa[i]
			if sa.rank[a] != sa.rank[b] {
				newRank[sa.sa[i]]++
			} else {
				ra, rb := 0, 0
				if a+k < n {
					ra = sa.rank[a+k]
				}
				if b+k < n {
					rb = sa.rank[b+k]
				}
				if ra != rb {
					newRank[sa.sa[i]]++
				}
			}
		}
		sa.rank = newRank
	}
}

// Search 搜索模式
// 应用: 商品搜索优化
func (sa *SuffixArray) Search(pattern string) []int {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	results := make([]int, 0)
	for _, pos := range sa.sa {
		if pos+len(pattern) <= len(sa.text) {
			if sa.text[pos:pos+len(pattern)] == pattern {
				results = append(results, pos)
			}
		}
	}
	return results
}
