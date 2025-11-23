package algorithm

import (
	"math"
	"sync"
)

// ============================================================================
// 1. KMeans聚类 - 用户分群
// ============================================================================

// Point 数据点
type Point struct {
	ID    uint64
	Data  []float64
	Label int
}

// KMeans K均值聚类
type KMeans struct {
	k         int
	points    []*Point
	centroids [][]float64
	maxIter   int
	tolerance float64
	mu        sync.RWMutex
}

// NewKMeans 创建KMeans聚类器
func NewKMeans(k, maxIter int, tolerance float64) *KMeans {
	return &KMeans{
		k:         k,
		maxIter:   maxIter,
		tolerance: tolerance,
	}
}

// Fit 训练模型
// 应用: 用户分群、商品分类
func (km *KMeans) Fit(points []*Point) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.points = points
	if len(points) == 0 {
		return
	}

	dim := len(points[0].Data)
	km.centroids = make([][]float64, km.k)

	// 随机初始化质心
	for i := 0; i < km.k; i++ {
		km.centroids[i] = make([]float64, dim)
		copy(km.centroids[i], points[i%len(points)].Data)
	}

	// 迭代优化
	for iter := 0; iter < km.maxIter; iter++ {
		// 分配点到最近的质心
		for _, p := range km.points {
			minDist := math.Inf(1)
			minLabel := 0
			for j, centroid := range km.centroids {
				dist := km.euclideanDistance(p.Data, centroid)
				if dist < minDist {
					minDist = dist
					minLabel = j
				}
			}
			p.Label = minLabel
		}

		// 更新质心
		oldCentroids := make([][]float64, km.k)
		for i := range oldCentroids {
			oldCentroids[i] = make([]float64, dim)
			copy(oldCentroids[i], km.centroids[i])
		}

		for j := 0; j < km.k; j++ {
			count := 0
			for d := 0; d < dim; d++ {
				km.centroids[j][d] = 0
			}

			for _, p := range km.points {
				if p.Label == j {
					for d := 0; d < dim; d++ {
						km.centroids[j][d] += p.Data[d]
					}
					count++
				}
			}

			if count > 0 {
				for d := 0; d < dim; d++ {
					km.centroids[j][d] /= float64(count)
				}
			}
		}

		// 检查收敛
		converged := true
		for j := 0; j < km.k; j++ {
			if km.euclideanDistance(oldCentroids[j], km.centroids[j]) > km.tolerance {
				converged = false
				break
			}
		}

		if converged {
			break
		}
	}
}

// Predict 预测标签
func (km *KMeans) Predict(data []float64) int {
	km.mu.RLock()
	defer km.mu.RUnlock()

	minDist := math.Inf(1)
	minLabel := 0

	for j, centroid := range km.centroids {
		dist := km.euclideanDistance(data, centroid)
		if dist < minDist {
			minDist = dist
			minLabel = j
		}
	}

	return minLabel
}

// euclideanDistance 欧几里得距离
func (km *KMeans) euclideanDistance(a, b []float64) float64 {
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
