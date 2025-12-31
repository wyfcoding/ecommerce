package search

import (
	"github.com/wyfcoding/pkg/algorithm"
)

// ProductVector 带有向量的商品
type ProductVector struct {
	ID        uint64
	Embedding []float64
}

// VectorSearcher 基于 K-D 树的向量相似度搜索器
type VectorSearcher struct {
	tree *algorithm.KDTree
}

// Build 构建向量索引
func (s *VectorSearcher) Build(products []ProductVector) {
	points := make([]algorithm.KDPoint, len(products))
	for i, p := range products {
		points[i] = algorithm.KDPoint{ID: p.ID, Vector: p.Embedding}
	}
	s.tree = algorithm.NewKDTree(points)
}

// FindMostSimilar 找到最相似的商品
func (s *VectorSearcher) FindMostSimilar(embedding []float64) (uint64, float64) {
	if s.tree == nil {
		return 0, 0
	}
	point, dist := s.tree.Nearest(embedding)
	return point.ID, dist
}
