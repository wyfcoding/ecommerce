package search

import (
	"strings"
	"sync"

	"github.com/wyfcoding/pkg/algorithm"
)

// --- Memory Searcher (Substring) ---

type ProductSearchEntry struct {
	ID   uint64
	Name string
}

type MemorySearcher struct {
	sa      *algorithm.SuffixArray
	entries []ProductSearchEntry
	rawText string
	offsets []int
	mu      sync.RWMutex
}

func (s *MemorySearcher) Build(entries []ProductSearchEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var sb strings.Builder
	offsets := make([]int, len(entries))
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

// --- Vector Searcher (Similarity) ---

type ProductVector struct {
	ID        uint64
	Embedding []float64
}

type VectorSearcher struct {
	tree *algorithm.KDTree
}

func (s *VectorSearcher) Build(products []ProductVector) {
	points := make([]algorithm.KDPoint, len(products))
	for i, p := range products {
		points[i] = algorithm.KDPoint{ID: p.ID, Vector: p.Embedding}
	}
	s.tree = algorithm.NewKDTree(points)
}

func (s *VectorSearcher) FindMostSimilar(embedding []float64) (uint64, float64) {
	if s.tree == nil {
		return 0, 0
	}
	point, dist := s.tree.Nearest(embedding)
	return point.ID, dist
}
