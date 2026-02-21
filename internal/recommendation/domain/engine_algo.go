// 变更说明：
// 从 pkg/algos/ml/recommendation_engine.go 迁移。
// 实现了基于用户 (User-Based) 和物品 (Item-Based) 的协同过滤，以及热门与个性化推荐。
package domain

import (
	"math"
	"slices"
	"sync"
)

type RecommendationEngine struct {
	userItemMatrix map[uint64]map[uint64]float64
	itemUserMatrix map[uint64]map[uint64]float64
	itemViews      map[uint64]int
	itemSales      map[uint64]int
	mu             sync.RWMutex
}

func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{
		userItemMatrix: make(map[uint64]map[uint64]float64),
		itemUserMatrix: make(map[uint64]map[uint64]float64),
		itemViews:      make(map[uint64]int),
		itemSales:      make(map[uint64]int),
	}
}

func (re *RecommendationEngine) UserBasedCF(userID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	userRatings := re.userItemMatrix[userID]
	if len(userRatings) == 0 {
		return nil
	}

	predictions := make(map[uint64]float64)
	for otherUID, otherRatings := range re.userItemMatrix {
		if otherUID == userID {
			continue
		}
		sim := re.cosineSimilarity(userID, otherUID)
		if sim <= 0.05 {
			continue
		}
		for itemID, r := range otherRatings {
			if _, ok := userRatings[itemID]; !ok {
				predictions[itemID] += sim * r
			}
		}
	}
	return re.topN(predictions, topN)
}

func (re *RecommendationEngine) ItemBasedCF(userID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	userRatings := re.userItemMatrix[userID]
	predictions := make(map[uint64]float64)
	for itemID, r := range userRatings {
		for candID := range re.itemUserMatrix {
			if _, ok := userRatings[candID]; ok {
				continue
			}
			sim := re.itemSimilarity(itemID, candID)
			predictions[candID] += sim * r
		}
	}
	return re.topN(predictions, topN)
}

func (re *RecommendationEngine) cosineSimilarity(u1, u2 uint64) float64 {
	r1, r2 := re.userItemMatrix[u1], re.userItemMatrix[u2]
	var dot, n1, n2 float64
	for id, v1 := range r1 {
		if v2, ok := r2[id]; ok {
			dot += v1 * v2
		}
		n1 += v1 * v1
	}
	for _, v2 := range r2 {
		n2 += v2 * v2
	}
	if n1*n2 == 0 {
		return 0
	}
	return dot / (math.Sqrt(n1) * math.Sqrt(n2))
}

func (re *RecommendationEngine) itemSimilarity(i1, i2 uint64) float64 {
	u1, u2 := re.itemUserMatrix[i1], re.itemUserMatrix[i2]
	var dot, n1, n2 float64
	for id, v1 := range u1 {
		if v2, ok := u2[id]; ok {
			dot += v1 * v2
		}
		n1 += v1 * v1
	}
	for _, v2 := range u2 {
		n2 += v2 * v2
	}
	if n1*n2 == 0 {
		return 0
	}
	return dot / (math.Sqrt(n1) * math.Sqrt(n2))
}

func (re *RecommendationEngine) topN(scores map[uint64]float64, n int) []uint64 {
	type item struct {
		id uint64
		s  float64
	}
	list := make([]item, 0)
	for id, s := range scores {
		list = append(list, item{id, s})
	}
	slices.SortFunc(list, func(a, b item) int {
		if a.s > b.s {
			return -1
		}
		return 1
	})
	res := make([]uint64, 0)
	for i := 0; i < len(list) && i < n; i++ {
		res = append(res, list[i].id)
	}
	return res
}
