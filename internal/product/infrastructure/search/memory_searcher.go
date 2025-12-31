package search

import (
	"strings"
	"sync"

	"github.com/wyfcoding/pkg/algorithm"
)

// ProductSearchEntry 商品搜索条目
type ProductSearchEntry struct {
	ID   uint64
	Name string
}

// MemorySearcher 基于后缀数组的内存高性能子串搜索器
type MemorySearcher struct {
	sa      *algorithm.SuffixArray
	entries []ProductSearchEntry
	rawText string
	offsets []int // 记录每个 Entry 在 rawText 中的起始位置
	mu      sync.RWMutex
}

// Build 构建搜索索引
func (s *MemorySearcher) Build(entries []ProductSearchEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sb strings.Builder
	offsets := make([]int, len(entries))

	// 使用特殊字符分隔，避免跨商品匹配
	separator := "\x01"

	for i, e := range entries {
		offsets[i] = sb.Len()
		sb.WriteString(e.Name)
		sb.WriteString(separator)
	}

	rawText := sb.String()
	s.rawText = rawText
	s.offsets = offsets
	s.entries = entries
	s.sa = algorithm.NewSuffixArray(rawText)
}

// Search 子串查找
func (s *MemorySearcher) Search(query string, limit int) []ProductSearchEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sa == nil {
		return nil
	}

	positions := s.sa.Search(query)
	if len(positions) == 0 {
		return nil
	}

	// 将查找到的起始位置映射回商品对象
	resultSet := make(map[uint64]ProductSearchEntry)
	for _, pos := range positions {
		// 二分查找 pos 属于哪个 Entry
		idx := s.findEntryIndex(pos)
		if idx != -1 {
			e := s.entries[idx]
			resultSet[e.ID] = e
		}
		if len(resultSet) >= limit {
			break
		}
	}

	results := make([]ProductSearchEntry, 0, len(resultSet))
	for _, v := range resultSet {
		results = append(results, v)
	}
	return results
}

func (s *MemorySearcher) findEntryIndex(pos int) int {
	// 在 offsets 中寻找最后一个 <= pos 的位置
	l, r := 0, len(s.offsets)-1
	ans := -1
	for l <= r {
		mid := l + (r-l)/2
		if s.offsets[mid] <= pos {
			ans = mid
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return ans
}
