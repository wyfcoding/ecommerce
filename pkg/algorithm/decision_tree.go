package algorithm

import (
	"math"
	"sort"
	"sync"
)

// ============================================================================
// 2. 决策树 - 用户行为预测
// ============================================================================

// TreeNode 决策树节点
type TreeNode struct {
	Feature   int
	Threshold float64
	Left      *TreeNode
	Right     *TreeNode
	Label     int
	IsLeaf    bool
}

// DecisionTree 决策树
type DecisionTree struct {
	root      *TreeNode
	maxDepth  int
	minSample int
	mu        sync.RWMutex
}

// NewDecisionTree 创建决策树
func NewDecisionTree(maxDepth, minSample int) *DecisionTree {
	return &DecisionTree{
		maxDepth:  maxDepth,
		minSample: minSample,
	}
}

// Fit 训练决策树
// 应用: 用户行为预测、风险评估
func (dt *DecisionTree) Fit(points []*Point, labels []int) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	dt.root = dt.buildTree(points, labels, 0)
}

// buildTree 递归构建树
func (dt *DecisionTree) buildTree(points []*Point, labels []int, depth int) *TreeNode {
	if len(points) < dt.minSample || depth >= dt.maxDepth {
		// 创建叶子节点
		labelCount := make(map[int]int)
		for _, label := range labels {
			labelCount[label]++
		}

		maxCount := 0
		maxLabel := 0
		for label, count := range labelCount {
			if count > maxCount {
				maxCount = count
				maxLabel = label
			}
		}

		return &TreeNode{
			IsLeaf: true,
			Label:  maxLabel,
		}
	}

	// 找最佳分割
	bestGain := 0.0
	bestFeature := 0
	bestThreshold := 0.0

	dim := len(points[0].Data)
	for feature := 0; feature < dim; feature++ {
		// 获取该特征的所有值
		values := make([]float64, len(points))
		for i, p := range points {
			values[i] = p.Data[feature]
		}
		sort.Float64s(values)

		// 尝试不同的分割点
		for i := 0; i < len(values)-1; i++ {
			threshold := (values[i] + values[i+1]) / 2

			// 分割数据
			leftPoints := make([]*Point, 0)
			leftLabels := make([]int, 0)
			rightPoints := make([]*Point, 0)
			rightLabels := make([]int, 0)

			for j, p := range points {
				if p.Data[feature] <= threshold {
					leftPoints = append(leftPoints, p)
					leftLabels = append(leftLabels, labels[j])
				} else {
					rightPoints = append(rightPoints, p)
					rightLabels = append(rightLabels, labels[j])
				}
			}

			if len(leftPoints) == 0 || len(rightPoints) == 0 {
				continue
			}

			// 计算信息增益
			gain := dt.calculateGain(labels, leftLabels, rightLabels)
			if gain > bestGain {
				bestGain = gain
				bestFeature = feature
				bestThreshold = threshold
			}
		}
	}

	// 分割数据
	leftPoints := make([]*Point, 0)
	leftLabels := make([]int, 0)
	rightPoints := make([]*Point, 0)
	rightLabels := make([]int, 0)

	for i, p := range points {
		if p.Data[bestFeature] <= bestThreshold {
			leftPoints = append(leftPoints, p)
			leftLabels = append(leftLabels, labels[i])
		} else {
			rightPoints = append(rightPoints, p)
			rightLabels = append(rightLabels, labels[i])
		}
	}

	return &TreeNode{
		Feature:   bestFeature,
		Threshold: bestThreshold,
		Left:      dt.buildTree(leftPoints, leftLabels, depth+1),
		Right:     dt.buildTree(rightPoints, rightLabels, depth+1),
		IsLeaf:    false,
	}
}

// calculateGain 计算信息增益
func (dt *DecisionTree) calculateGain(parent, left, right []int) float64 {
	parentEntropy := dt.entropy(parent)

	leftWeight := float64(len(left)) / float64(len(parent))
	rightWeight := float64(len(right)) / float64(len(parent))

	childEntropy := leftWeight*dt.entropy(left) + rightWeight*dt.entropy(right)

	return parentEntropy - childEntropy
}

// entropy 计算熵
func (dt *DecisionTree) entropy(labels []int) float64 {
	if len(labels) == 0 {
		return 0
	}

	labelCount := make(map[int]int)
	for _, label := range labels {
		labelCount[label]++
	}

	var entropy float64
	for _, count := range labelCount {
		p := float64(count) / float64(len(labels))
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// Predict 预测
func (dt *DecisionTree) Predict(data []float64) int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	return dt.predictNode(dt.root, data)
}

// predictNode 递归预测
func (dt *DecisionTree) predictNode(node *TreeNode, data []float64) int {
	if node.IsLeaf {
		return node.Label
	}

	if data[node.Feature] <= node.Threshold {
		return dt.predictNode(node.Left, data)
	}
	return dt.predictNode(node.Right, data)
}
