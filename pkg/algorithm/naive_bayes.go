package algorithm

import (
	"math"
	"sync"
)

// ============================================================================
// 5. 朴素贝叶斯 - 文本分类
// ============================================================================

// NaiveBayes 朴素贝叶斯分类器
type NaiveBayes struct {
	classes      map[string]float64            // 类别概率
	features     map[string]map[string]float64 // 特征概率
	classCount   map[string]int
	featureCount map[string]map[string]int
	mu           sync.RWMutex
}

// NewNaiveBayes 创建朴素贝叶斯分类器
func NewNaiveBayes() *NaiveBayes {
	return &NaiveBayes{
		classes:      make(map[string]float64),
		features:     make(map[string]map[string]float64),
		classCount:   make(map[string]int),
		featureCount: make(map[string]map[string]int),
	}
}

// Train 训练模型
// 应用: 商品评论分类、内容审核
func (nb *NaiveBayes) Train(documents [][]string, labels []string) {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	totalDocs := len(documents)

	// 计算类别概率
	for _, label := range labels {
		nb.classCount[label]++
	}

	for label, count := range nb.classCount {
		nb.classes[label] = float64(count) / float64(totalDocs)
	}

	// 计算特征概率
	for i, doc := range documents {
		label := labels[i]

		if nb.featureCount[label] == nil {
			nb.featureCount[label] = make(map[string]int)
		}

		for _, word := range doc {
			nb.featureCount[label][word]++
		}
	}

	// 计算条件概率
	for label := range nb.classCount {
		nb.features[label] = make(map[string]float64)

		totalWords := 0
		for _, count := range nb.featureCount[label] {
			totalWords += count
		}

		for word, count := range nb.featureCount[label] {
			// 拉普拉斯平滑
			nb.features[label][word] = float64(count+1) / float64(totalWords+len(nb.featureCount[label]))
		}
	}
}

// Predict 预测
func (nb *NaiveBayes) Predict(document []string) string {
	nb.mu.RLock()
	defer nb.mu.RUnlock()

	maxProb := math.Inf(-1)
	maxLabel := ""

	for label, classProb := range nb.classes {
		prob := math.Log(classProb)

		for _, word := range document {
			if wordProb, exists := nb.features[label][word]; exists {
				prob += math.Log(wordProb)
			} else {
				// 未见过的词
				prob += math.Log(1.0 / float64(len(nb.featureCount[label])+1))
			}
		}

		if prob > maxProb {
			maxProb = prob
			maxLabel = label
		}
	}

	return maxLabel
}
