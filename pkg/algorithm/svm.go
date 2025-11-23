package algorithm

import (
	"sync"
)

// ============================================================================
// 6. 支持向量机 (SVM) - 简化版
// ============================================================================

// SVM 支持向量机（简化实现）
type SVM struct {
	weights []float64
	bias    float64
	mu      sync.RWMutex
}

// NewSVM 创建SVM
func NewSVM(dim int) *SVM {
	return &SVM{
		weights: make([]float64, dim),
		bias:    0,
	}
}

// Train 训练SVM（梯度下降）
// 应用: 用户分类、异常检测
func (svm *SVM) Train(points []*Point, labels []int, learningRate float64, iterations int) {
	svm.mu.Lock()
	defer svm.mu.Unlock()

	dim := len(points[0].Data)

	for iter := 0; iter < iterations; iter++ {
		for i, p := range points {
			// 计算预测值
			pred := svm.bias
			for j := 0; j < dim; j++ {
				pred += svm.weights[j] * p.Data[j]
			}

			// 计算标签（-1或1）
			label := float64(labels[i])
			if label == 0 {
				label = -1
			}

			// 计算误差
			error := label - pred

			// 更新权重
			for j := 0; j < dim; j++ {
				svm.weights[j] += learningRate * error * p.Data[j]
			}

			// 更新偏置
			svm.bias += learningRate * error
		}
	}
}

// Predict 预测
func (svm *SVM) Predict(data []float64) int {
	svm.mu.RLock()
	defer svm.mu.RUnlock()

	pred := svm.bias
	for i := range data {
		pred += svm.weights[i] * data[i]
	}

	if pred >= 0 {
		return 1
	}
	return 0
}
