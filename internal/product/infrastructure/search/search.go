// Package search 提供了商品搜索的多种基础设施实现，包括基于后缀数组的内存模糊搜索及基于 KD-Tree 的向量相似度检索。
package search

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
)

// --- Memory Searcher (基于 SuffixArray 的高性能文本子串检索) ---

// ProductSearchEntry 描述了搜索索引中的单条简易记录。
type ProductSearchEntry struct {
	ID   uint64 // 商品 ID
	Name string // 商品名称（检索主键）
}

// MemorySearcher 利用后缀数组算法实现极速的内存级子串搜索。
type MemorySearcher struct {
	sa      *algorithm.SuffixArray // 核心后缀数组实例
	entries []ProductSearchEntry   // 原始记录快照
	rawText string                 // 拼接后的原始索引文本
	offsets []int                  // 每条记录在原始文本中的起始偏移量
	mu      sync.RWMutex           // 保证重建索引时的并发安全
}

// Build 根据传入的商品列表全量重构内存索引。
// 流程：拼接字符串 -> 记录偏移量 -> 构造后缀数组。
func (s *MemorySearcher) Build(entries []ProductSearchEntry) {
	start := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	var sb strings.Builder
	offsets := make([]int, len(entries))
	separator := "\x01" // 使用不可见字符作为记录分隔符

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

	slog.Info("memory search index built successfully",
		"entries_count", len(entries),
		"text_size", len(rawText),
		"duration", time.Since(start),
	)
}

// Search 执行模糊子串搜索并返回去重后的匹配结果。
func (s *MemorySearcher) Search(query string, limit int) []ProductSearchEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sa == nil {
		return nil
	}

	// 利用后缀数组在 O(m log n) 时间内找到所有匹配位置
	positions := s.sa.Search(query)
	if len(positions) == 0 {
		return nil
	}

	resultSet := make(map[uint64]ProductSearchEntry)
	for _, pos := range positions {
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

// findEntryIndex 使用二分查找定位位置 pos 属于哪一个商品记录。
func (s *MemorySearcher) findEntryIndex(pos int) int {
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

// --- Vector Searcher (基于 KD-Tree 的空间度量衡检索，用于推荐或相似搜索) ---

// ProductVector 描述商品的特征向量。
type ProductVector struct {
	ID        uint64
	Embedding []float64 // 多维特征嵌入
}

// VectorSearcher 提供了基于空间坐标的最近邻检索能力。
type VectorSearcher struct {
	tree *algorithm.KDTree // 底层 KD-Tree 实例
}

// Build 构造向量空间索引。
func (s *VectorSearcher) Build(products []ProductVector) {
	start := time.Now()
	points := make([]algorithm.KDPoint, len(products))
	for i, p := range products {
		points[i] = algorithm.KDPoint{ID: p.ID, Vector: p.Embedding}
	}
	tree, err := algorithm.NewKDTree(points)
	if err != nil {
		slog.Error("failed to build vector search index", "error", err)
		s.tree = nil
		return
	}
	s.tree = tree

	slog.Info("vector search index built successfully",
		"points_count", len(products),
		"duration", time.Since(start),
	)
}

// FindMostSimilar 寻找在特征空间中距离给定向量最近的商品。
func (s *VectorSearcher) FindMostSimilar(embedding []float64) (uint64, float64) {
	if s.tree == nil {
		return 0, 0
	}
	point, dist := s.tree.Nearest(embedding)
	return point.ID, dist
}
