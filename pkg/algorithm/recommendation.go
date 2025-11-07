package algorithm

import (
	"math"
	"sort"
	"sync"
	"time"
)

// RecommendationEngine 推荐引擎
type RecommendationEngine struct {
	userItemMatrix map[uint64]map[uint64]float64 // 用户-物品评分矩阵
	itemUserMatrix map[uint64]map[uint64]float64 // 物品-用户评分矩阵
	itemViews      map[uint64]int                // 物品浏览次数
	itemSales      map[uint64]int                // 物品销量
	itemScores     map[uint64]float64            // 物品综合得分
	lastUpdate     map[uint64]time.Time          // 物品最后更新时间
	mu             sync.RWMutex
}

// NewRecommendationEngine 创建推荐引擎
func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{
		userItemMatrix: make(map[uint64]map[uint64]float64),
		itemUserMatrix: make(map[uint64]map[uint64]float64),
		itemViews:      make(map[uint64]int),
		itemSales:      make(map[uint64]int),
		itemScores:     make(map[uint64]float64),
		lastUpdate:     make(map[uint64]time.Time),
	}
}

// AddRating 添加评分（浏览、购买、收藏等行为）
func (re *RecommendationEngine) AddRating(userID, itemID uint64, rating float64) {
	re.mu.Lock()
	defer re.mu.Unlock()

	if re.userItemMatrix[userID] == nil {
		re.userItemMatrix[userID] = make(map[uint64]float64)
	}
	re.userItemMatrix[userID][itemID] = rating

	if re.itemUserMatrix[itemID] == nil {
		re.itemUserMatrix[itemID] = make(map[uint64]float64)
	}
	re.itemUserMatrix[itemID][userID] = rating

	re.lastUpdate[itemID] = time.Now()
}

// AddView 添加浏览记录
func (re *RecommendationEngine) AddView(itemID uint64) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.itemViews[itemID]++
	re.lastUpdate[itemID] = time.Now()
}

// AddSale 添加销售记录
func (re *RecommendationEngine) AddSale(itemID uint64) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.itemSales[itemID]++
	re.lastUpdate[itemID] = time.Now()
}

// UserBasedCF 基于用户的协同过滤
func (re *RecommendationEngine) UserBasedCF(userID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// 找出相似用户
	similarities := make(map[uint64]float64)
	for otherUserID := range re.userItemMatrix {
		if otherUserID != userID {
			sim := re.cosineSimilarity(userID, otherUserID)
			if sim > 0 {
				similarities[otherUserID] = sim
			}
		}
	}

	// 计算物品的预测评分
	predictions := make(map[uint64]float64)
	userRatings := re.userItemMatrix[userID]

	for otherUserID, similarity := range similarities {
		for itemID, rating := range re.userItemMatrix[otherUserID] {
			// 跳过用户已评分的物品
			if _, exists := userRatings[itemID]; exists {
				continue
			}

			predictions[itemID] += similarity * rating
		}
	}

	// 排序并返回TopN
	return re.topNItems(predictions, topN)
}

// ItemBasedCF 基于物品的协同过滤
func (re *RecommendationEngine) ItemBasedCF(userID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	userRatings := re.userItemMatrix[userID]
	if len(userRatings) == 0 {
		return nil
	}

	// 计算物品相似度并预测评分
	predictions := make(map[uint64]float64)

	for ratedItemID := range userRatings {
		for candidateItemID := range re.itemUserMatrix {
			// 跳过已评分的物品
			if _, exists := userRatings[candidateItemID]; exists {
				continue
			}

			// 计算物品相似度
			similarity := re.itemSimilarity(ratedItemID, candidateItemID)
			if similarity > 0 {
				predictions[candidateItemID] += similarity * userRatings[ratedItemID]
			}
		}
	}

	return re.topNItems(predictions, topN)
}

// HotItems 热门商品推荐（带时间衰减）
func (re *RecommendationEngine) HotItems(topN int, decayHours float64) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	now := time.Now()
	scores := make(map[uint64]float64)

	for itemID := range re.itemUserMatrix {
		// 基础热度分数
		viewScore := float64(re.itemViews[itemID])
		saleScore := float64(re.itemSales[itemID]) * 5 // 销量权重更高

		baseScore := viewScore + saleScore

		// 时间衰减
		if lastUpdate, exists := re.lastUpdate[itemID]; exists {
			hoursPassed := now.Sub(lastUpdate).Hours()
			decayFactor := math.Exp(-hoursPassed / decayHours)
			scores[itemID] = baseScore * decayFactor
		} else {
			scores[itemID] = baseScore
		}
	}

	return re.topNItems(scores, topN)
}

// PersonalizedHot 个性化热门推荐（结合用户偏好和热度）
func (re *RecommendationEngine) PersonalizedHot(userID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// 获取用户偏好的类别（这里简化处理，实际应该从商品类别获取）
	userRatings := re.userItemMatrix[userID]
	
	scores := make(map[uint64]float64)
	now := time.Now()

	for itemID := range re.itemUserMatrix {
		// 跳过用户已评分的物品
		if _, exists := userRatings[itemID]; exists {
			continue
		}

		// 热度分数
		viewScore := float64(re.itemViews[itemID])
		saleScore := float64(re.itemSales[itemID]) * 5
		hotScore := viewScore + saleScore

		// 时间衰减
		decayFactor := 1.0
		if lastUpdate, exists := re.lastUpdate[itemID]; exists {
			hoursPassed := now.Sub(lastUpdate).Hours()
			decayFactor = math.Exp(-hoursPassed / 24.0) // 24小时衰减
		}

		// 用户偏好分数（基于物品相似度）
		preferenceScore := 0.0
		count := 0
		for ratedItemID := range userRatings {
			sim := re.itemSimilarity(ratedItemID, itemID)
			if sim > 0 {
				preferenceScore += sim
				count++
			}
		}
		if count > 0 {
			preferenceScore /= float64(count)
		}

		// 综合评分：热度50%，个人偏好30%，时间衰减20%
		scores[itemID] = 0.5*hotScore + 0.3*preferenceScore*100 + 0.2*decayFactor*100
	}

	return re.topNItems(scores, topN)
}

// SimilarItems 相似商品推荐
func (re *RecommendationEngine) SimilarItems(itemID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	similarities := make(map[uint64]float64)

	for candidateItemID := range re.itemUserMatrix {
		if candidateItemID == itemID {
			continue
		}

		sim := re.itemSimilarity(itemID, candidateItemID)
		if sim > 0 {
			similarities[candidateItemID] = sim
		}
	}

	return re.topNItems(similarities, topN)
}

// AlsoBought 购买了A的用户还购买了B
func (re *RecommendationEngine) AlsoBought(itemID uint64, topN int) []uint64 {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// 找出购买了该商品的用户
	users := re.itemUserMatrix[itemID]
	if len(users) == 0 {
		return nil
	}

	// 统计这些用户还购买了哪些商品
	itemCounts := make(map[uint64]int)
	for userID := range users {
		for otherItemID := range re.userItemMatrix[userID] {
			if otherItemID != itemID {
				itemCounts[otherItemID]++
			}
		}
	}

	// 转换为float64分数
	scores := make(map[uint64]float64)
	for itemID, count := range itemCounts {
		scores[itemID] = float64(count)
	}

	return re.topNItems(scores, topN)
}

// cosineSimilarity 计算用户余弦相似度
func (re *RecommendationEngine) cosineSimilarity(user1, user2 uint64) float64 {
	ratings1 := re.userItemMatrix[user1]
	ratings2 := re.userItemMatrix[user2]

	if len(ratings1) == 0 || len(ratings2) == 0 {
		return 0
	}

	var dotProduct, norm1, norm2 float64

	for itemID, rating1 := range ratings1 {
		if rating2, exists := ratings2[itemID]; exists {
			dotProduct += rating1 * rating2
		}
		norm1 += rating1 * rating1
	}

	for _, rating2 := range ratings2 {
		norm2 += rating2 * rating2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// itemSimilarity 计算物品相似度
func (re *RecommendationEngine) itemSimilarity(item1, item2 uint64) float64 {
	users1 := re.itemUserMatrix[item1]
	users2 := re.itemUserMatrix[item2]

	if len(users1) == 0 || len(users2) == 0 {
		return 0
	}

	var dotProduct, norm1, norm2 float64

	for userID, rating1 := range users1 {
		if rating2, exists := users2[userID]; exists {
			dotProduct += rating1 * rating2
		}
		norm1 += rating1 * rating1
	}

	for _, rating2 := range users2 {
		norm2 += rating2 * rating2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// topNItems 返回分数最高的N个物品
func (re *RecommendationEngine) topNItems(scores map[uint64]float64, topN int) []uint64 {
	type itemScore struct {
		itemID uint64
		score  float64
	}

	items := make([]itemScore, 0, len(scores))
	for itemID, score := range scores {
		items = append(items, itemScore{itemID, score})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	result := make([]uint64, 0, topN)
	for i := 0; i < len(items) && i < topN; i++ {
		result = append(result, items[i].itemID)
	}

	return result
}

// UpdateItemScore 更新物品综合得分（定期批量更新）
func (re *RecommendationEngine) UpdateItemScore(itemID uint64) {
	re.mu.Lock()
	defer re.mu.Unlock()

	viewScore := float64(re.itemViews[itemID])
	saleScore := float64(re.itemSales[itemID]) * 5

	// 时间衰减
	decayFactor := 1.0
	if lastUpdate, exists := re.lastUpdate[itemID]; exists {
		hoursPassed := time.Since(lastUpdate).Hours()
		decayFactor = math.Exp(-hoursPassed / 24.0)
	}

	re.itemScores[itemID] = (viewScore + saleScore) * decayFactor
}

// BatchUpdateScores 批量更新所有物品得分
func (re *RecommendationEngine) BatchUpdateScores() {
	re.mu.Lock()
	defer re.mu.Unlock()

	for itemID := range re.itemUserMatrix {
		viewScore := float64(re.itemViews[itemID])
		saleScore := float64(re.itemSales[itemID]) * 5

		decayFactor := 1.0
		if lastUpdate, exists := re.lastUpdate[itemID]; exists {
			hoursPassed := time.Since(lastUpdate).Hours()
			decayFactor = math.Exp(-hoursPassed / 24.0)
		}

		re.itemScores[itemID] = (viewScore + saleScore) * decayFactor
	}
}
