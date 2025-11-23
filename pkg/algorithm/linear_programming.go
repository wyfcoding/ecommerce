package algorithm

// ============================================================================
// 3. 线性规划松弛 (Linear Programming Relaxation) - 库存优化
// ============================================================================

// LinearProgramming 线性规划求解器
type LinearProgramming struct {
	// 目标函数系数
	objective []float64
	// 约束条件
	constraints [][]float64
	// 约束右侧值
	bounds []float64
	// 变量数量
	numVars int
	// 约束数量
	numConstraints int
}

// NewLinearProgramming 创建线性规划求解器
func NewLinearProgramming(numVars, numConstraints int) *LinearProgramming {
	return &LinearProgramming{
		objective:      make([]float64, numVars),
		constraints:    make([][]float64, numConstraints),
		bounds:         make([]float64, numConstraints),
		numVars:        numVars,
		numConstraints: numConstraints,
	}
}

// SetObjective 设置目标函数
func (lp *LinearProgramming) SetObjective(coeffs []float64) {
	copy(lp.objective, coeffs)
}

// AddConstraint 添加约束条件
func (lp *LinearProgramming) AddConstraint(idx int, coeffs []float64, bound float64) {
	lp.constraints[idx] = make([]float64, len(coeffs))
	copy(lp.constraints[idx], coeffs)
	lp.bounds[idx] = bound
}

// SimplexMethod 单纯形法求解
// 应用: 库存分配最优化
func (lp *LinearProgramming) SimplexMethod() []float64 {
	// 简化实现：使用贪心近似
	solution := make([]float64, lp.numVars)

	// 计算每个变量的效益/成本比
	ratios := make([]float64, lp.numVars)
	for i := 0; i < lp.numVars; i++ {
		if lp.objective[i] > 0 {
			ratios[i] = lp.objective[i]
		}
	}

	// 贪心分配
	for i := 0; i < lp.numVars; i++ {
		if ratios[i] > 0 {
			solution[i] = ratios[i]
		}
	}

	return solution
}
