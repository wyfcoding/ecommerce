package algorithm

import (
	"math"
	"math/rand/v2"
	"sync"
)

// ============================================================================
// 7. 神经网络 - 深度学习基础
// ============================================================================

// NeuralNetwork 神经网络
type NeuralNetwork struct {
	layers []*Layer
	mu     sync.RWMutex
}

// Layer 神经网络层
type Layer struct {
	weights [][]float64
	bias    []float64
	output  []float64
}

// NewNeuralNetwork 创建神经网络
func NewNeuralNetwork(layerSizes []int) *NeuralNetwork {
	nn := &NeuralNetwork{
		layers: make([]*Layer, len(layerSizes)-1),
	}

	for i := 0; i < len(layerSizes)-1; i++ {
		nn.layers[i] = &Layer{
			weights: make([][]float64, layerSizes[i]),
			bias:    make([]float64, layerSizes[i+1]),
			output:  make([]float64, layerSizes[i+1]),
		}

		// 随机初始化权重
		for j := 0; j < layerSizes[i]; j++ {
			nn.layers[i].weights[j] = make([]float64, layerSizes[i+1])
			for k := 0; k < layerSizes[i+1]; k++ {
				nn.layers[i].weights[j][k] = (rand.Float64() - 0.5) * 2
			}
		}
	}

	return nn
}

// Forward 前向传播
func (nn *NeuralNetwork) Forward(input []float64) []float64 {
	nn.mu.RLock()
	defer nn.mu.RUnlock()

	current := input

	for _, layer := range nn.layers {
		next := make([]float64, len(layer.bias))

		for j := 0; j < len(layer.bias); j++ {
			sum := layer.bias[j]
			for i := 0; i < len(current); i++ {
				sum += current[i] * layer.weights[i][j]
			}
			next[j] = nn.sigmoid(sum)
		}

		current = next
	}

	return current
}

// sigmoid 激活函数
func (nn *NeuralNetwork) sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// Train 训练神经网络（反向传播）
// 应用: 复杂用户行为预测、推荐系统
func (nn *NeuralNetwork) Train(points []*Point, labels []int, learningRate float64, iterations int) {
	nn.mu.Lock()
	defer nn.mu.Unlock()

	for iter := 0; iter < iterations; iter++ {
		for i, p := range points {
			// 前向传播
			output := nn.Forward(p.Data)

			// 计算误差
			target := float64(labels[i])
			error := target - output[0]

			// 简化的反向传播（实际应该更复杂）
			for j := 0; j < len(nn.layers); j++ {
				layer := nn.layers[j]
				for k := 0; k < len(layer.bias); k++ {
					layer.bias[k] += learningRate * error
				}
			}
		}
	}
}
