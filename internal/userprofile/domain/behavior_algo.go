// 变更说明：
// 从 pkg/algos/ml/user_behavior.go 迁移。
// 实现了用户行为异常识别算法，检测高频重复、单一被动行为及敏感操作序列。
package domain

import "time"

type UserBehavior struct {
	Timestamp time.Time
	IP        string
	UserAgent string
	Action    string
	UserID    uint64
}

func (ub *UserBehavior) Analyze(actions []string) float64 {
	if len(actions) == 0 {
		return 0.0
	}
	score := 0.0
	if ub.detectRepetition(actions) {
		score += 0.4
	}
	if ub.detectPassiveOnly(actions) {
		score += 0.2
	}
	if ub.detectSensitiveSequence(actions) {
		score += 0.3
	}
	return min(1.0, score)
}

func (ub *UserBehavior) detectRepetition(actions []string) bool {
	if len(actions) < 5 {
		return false
	}
	last := actions[len(actions)-1]
	count := 0
	for i := len(actions) - 1; i >= 0; i-- {
		if actions[i] == last {
			count++
		} else {
			break
		}
	}
	return count >= 5
}

func (ub *UserBehavior) detectPassiveOnly(actions []string) bool {
	for _, a := range actions {
		if a != "view" && a != "scroll" {
			return false
		}
	}
	return true
}

func (ub *UserBehavior) detectSensitiveSequence(actions []string) bool {
	seq := []string{"login", "change_password", "withdraw"}
	idx := 0
	for _, a := range actions {
		if a == seq[idx] {
			idx++
			if idx == len(seq) {
				return true
			}
		}
	}
	return false
}
