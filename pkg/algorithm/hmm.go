package algorithm

import (
	"math"
	"sync"
)

// ============================================================================
// 4. 隐马尔可夫模型 (HMM) - 序列预测
// ============================================================================

// HiddenMarkovModel 隐马尔可夫模型
type HiddenMarkovModel struct {
	states       []string
	observations []string
	// 状态转移概率
	transitionProb map[string]map[string]float64
	// 观测概率
	emissionProb map[string]map[string]float64
	// 初始概率
	initialProb map[string]float64
	mu          sync.RWMutex
}

// NewHiddenMarkovModel 创建HMM
func NewHiddenMarkovModel(states, observations []string) *HiddenMarkovModel {
	return &HiddenMarkovModel{
		states:         states,
		observations:   observations,
		transitionProb: make(map[string]map[string]float64),
		emissionProb:   make(map[string]map[string]float64),
		initialProb:    make(map[string]float64),
	}
}

// SetTransitionProb 设置状态转移概率
func (hmm *HiddenMarkovModel) SetTransitionProb(from, to string, prob float64) {
	hmm.mu.Lock()
	defer hmm.mu.Unlock()

	if hmm.transitionProb[from] == nil {
		hmm.transitionProb[from] = make(map[string]float64)
	}
	hmm.transitionProb[from][to] = prob
}

// SetEmissionProb 设置观测概率
func (hmm *HiddenMarkovModel) SetEmissionProb(state, obs string, prob float64) {
	hmm.mu.Lock()
	defer hmm.mu.Unlock()

	if hmm.emissionProb[state] == nil {
		hmm.emissionProb[state] = make(map[string]float64)
	}
	hmm.emissionProb[state][obs] = prob
}

// SetInitialProb 设置初始概率
func (hmm *HiddenMarkovModel) SetInitialProb(state string, prob float64) {
	hmm.mu.Lock()
	defer hmm.mu.Unlock()

	hmm.initialProb[state] = prob
}

// Viterbi Viterbi算法（最可能的状态序列）
// 应用: 用户行为序列预测、订单异常检测
func (hmm *HiddenMarkovModel) Viterbi(observations []string) []string {
	hmm.mu.RLock()
	defer hmm.mu.RUnlock()

	n := len(observations)
	m := len(hmm.states)

	// DP表
	dp := make([][]float64, n)
	path := make([][]int, n)

	for i := 0; i < n; i++ {
		dp[i] = make([]float64, m)
		path[i] = make([]int, m)
	}

	// 初始化
	for j, state := range hmm.states {
		dp[0][j] = math.Log(hmm.initialProb[state]) + math.Log(hmm.emissionProb[state][observations[0]])
	}

	// 递推
	for i := 1; i < n; i++ {
		for j, state := range hmm.states {
			maxProb := math.Inf(-1)
			maxK := 0

			for k, prevState := range hmm.states {
				prob := dp[i-1][k] + math.Log(hmm.transitionProb[prevState][state])
				if prob > maxProb {
					maxProb = prob
					maxK = k
				}
			}

			dp[i][j] = maxProb + math.Log(hmm.emissionProb[state][observations[i]])
			path[i][j] = maxK
		}
	}

	// 回溯
	result := make([]string, n)
	maxProb := math.Inf(-1)
	maxIdx := 0

	for j := 0; j < m; j++ {
		if dp[n-1][j] > maxProb {
			maxProb = dp[n-1][j]
			maxIdx = j
		}
	}

	result[n-1] = hmm.states[maxIdx]
	for i := n - 1; i > 0; i-- {
		maxIdx = path[i][maxIdx]
		result[i-1] = hmm.states[maxIdx]
	}

	return result
}
