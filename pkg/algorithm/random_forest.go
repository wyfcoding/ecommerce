package algorithm

import (
	"math"
	"math/rand/v2"
	"sync"
)

// ============================================================================
// 3. 随机森林 - 复杂预测
// ============================================================================

// RandomForest 随机森林
type RandomForest struct {
	trees    []*DecisionTree
	numTrees int
	mu       sync.RWMutex
}

// NewRandomForest 创建随机森林
func NewRandomForest(numTrees int) *RandomForest {
	return &RandomForest{
		trees:    make([]*DecisionTree, numTrees),
		numTrees: numTrees,
	}
}

// Fit 训练随机森林
// 应用: 复杂用户行为预测、欺诈检测
func (rf *RandomForest) Fit(points []*Point, labels []int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	for i := 0; i < rf.numTrees; i++ {
		// Bootstrap采样
		bootstrapPoints := make([]*Point, len(points))
		bootstrapLabels := make([]int, len(labels))

		for j := 0; j < len(points); j++ {
			idx := int(math.Floor(rand.Float64() * float64(len(points))))
			bootstrapPoints[j] = points[idx]
			bootstrapLabels[j] = labels[idx]
		}

		// 训练决策树
		tree := NewDecisionTree(10, 5)
		tree.Fit(bootstrapPoints, bootstrapLabels)
		rf.trees[i] = tree
	}
}

// Predict 预测（投票）
func (rf *RandomForest) Predict(data []float64) int {
	rf.mu.RLock()
	defer rf.mu.RUnlock()

	votes := make(map[int]int)
	for _, tree := range rf.trees {
		label := tree.Predict(data)
		votes[label]++
	}

	maxVotes := 0
	maxLabel := 0
	for label, count := range votes {
		if count > maxVotes {
			maxVotes = count
			maxLabel = label
		}
	}

	return maxLabel
}
