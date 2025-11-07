package algorithm

import (
	"sync"
	"time"
)

// UserBehavior 用户行为
type UserBehavior struct {
	UserID    uint64
	IP        string
	UserAgent string
	Timestamp time.Time
	Action    string
}

// AntiBotDetector 防刷检测器
type AntiBotDetector struct {
	// 滑动窗口：记录用户最近的请求时间
	userRequests map[uint64][]time.Time
	ipRequests   map[string][]time.Time
	
	// 行为特征
	userBehaviors map[uint64][]UserBehavior
	
	mu sync.RWMutex
}

// NewAntiBotDetector 创建防刷检测器
func NewAntiBotDetector() *AntiBotDetector {
	detector := &AntiBotDetector{
		userRequests:  make(map[uint64][]time.Time),
		ipRequests:    make(map[string][]time.Time),
		userBehaviors: make(map[uint64][]UserBehavior),
	}

	// 启动清理goroutine
	go detector.cleanup()

	return detector
}

// IsBot 判断是否为机器人
func (d *AntiBotDetector) IsBot(behavior UserBehavior) (bool, string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 检查请求频率（滑动窗口）
	if isHighFrequency, reason := d.checkFrequency(behavior); isHighFrequency {
		return true, reason
	}

	// 2. 检查行为模式
	if isSuspicious, reason := d.checkBehaviorPattern(behavior); isSuspicious {
		return true, reason
	}

	// 3. 检查IP异常
	if isAbnormalIP, reason := d.checkIPAbnormal(behavior); isAbnormalIP {
		return true, reason
	}

	// 记录行为
	d.recordBehavior(behavior)

	return false, ""
}

// checkFrequency 检查请求频率
func (d *AntiBotDetector) checkFrequency(behavior UserBehavior) (bool, string) {
	now := behavior.Timestamp
	windowSize := 10 * time.Second

	// 检查用户请求频率
	if requests, exists := d.userRequests[behavior.UserID]; exists {
		// 清理过期请求
		validRequests := make([]time.Time, 0)
		for _, t := range requests {
			if now.Sub(t) <= windowSize {
				validRequests = append(validRequests, t)
			}
		}
		d.userRequests[behavior.UserID] = validRequests

		// 10秒内超过20次请求
		if len(validRequests) >= 20 {
			return true, "用户请求频率过高"
		}
	}

	// 检查IP请求频率
	if requests, exists := d.ipRequests[behavior.IP]; exists {
		validRequests := make([]time.Time, 0)
		for _, t := range requests {
			if now.Sub(t) <= windowSize {
				validRequests = append(validRequests, t)
			}
		}
		d.ipRequests[behavior.IP] = validRequests

		// 10秒内超过50次请求（同一IP可能有多个用户）
		if len(validRequests) >= 50 {
			return true, "IP请求频率过高"
		}
	}

	return false, ""
}

// checkBehaviorPattern 检查行为模式
func (d *AntiBotDetector) checkBehaviorPattern(behavior UserBehavior) (bool, string) {
	behaviors, exists := d.userBehaviors[behavior.UserID]
	if !exists || len(behaviors) < 5 {
		return false, ""
	}

	// 取最近10个行为
	recentBehaviors := behaviors
	if len(behaviors) > 10 {
		recentBehaviors = behaviors[len(behaviors)-10:]
	}

	// 1. 检查时间间隔是否过于规律（机器人特征）
	if d.isRegularInterval(recentBehaviors) {
		return true, "请求时间间隔过于规律"
	}

	// 2. 检查是否缺少正常浏览行为（直接秒杀）
	if d.isDirectKill(recentBehaviors, behavior) {
		return true, "缺少正常浏览行为"
	}

	// 3. 检查UserAgent是否异常
	if d.isAbnormalUserAgent(behavior.UserAgent) {
		return true, "UserAgent异常"
	}

	return false, ""
}

// isRegularInterval 检查时间间隔是否过于规律
func (d *AntiBotDetector) isRegularInterval(behaviors []UserBehavior) bool {
	if len(behaviors) < 3 {
		return false
	}

	intervals := make([]float64, 0)
	for i := 1; i < len(behaviors); i++ {
		interval := behaviors[i].Timestamp.Sub(behaviors[i-1].Timestamp).Seconds()
		intervals = append(intervals, interval)
	}

	// 计算方差
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	variance := 0.0
	for _, interval := range intervals {
		variance += (interval - mean) * (interval - mean)
	}
	variance /= float64(len(intervals))

	// 方差小于0.1说明时间间隔非常规律（机器人特征）
	return variance < 0.1 && mean < 2.0
}

// isDirectKill 检查是否直接秒杀（缺少浏览行为）
func (d *AntiBotDetector) isDirectKill(behaviors []UserBehavior, current UserBehavior) bool {
	if current.Action != "kill" {
		return false
	}

	// 检查最近是否有浏览商品详情的行为
	hasView := false
	for i := len(behaviors) - 1; i >= 0 && i >= len(behaviors)-5; i-- {
		if behaviors[i].Action == "view" {
			hasView = true
			break
		}
	}

	return !hasView
}

// isAbnormalUserAgent 检查UserAgent是否异常
func (d *AntiBotDetector) isAbnormalUserAgent(userAgent string) bool {
	// 简单检查：是否包含常见浏览器标识
	normalAgents := []string{"Mozilla", "Chrome", "Safari", "Firefox", "Edge"}
	
	for _, agent := range normalAgents {
		if contains(userAgent, agent) {
			return false
		}
	}

	return true
}

// checkIPAbnormal 检查IP异常
func (d *AntiBotDetector) checkIPAbnormal(behavior UserBehavior) (bool, string) {
	// 检查同一IP下的用户数量
	ipUsers := make(map[uint64]bool)
	
	for userID, behaviors := range d.userBehaviors {
		for _, b := range behaviors {
			if b.IP == behavior.IP {
				ipUsers[userID] = true
			}
		}
	}

	// 同一IP超过10个用户
	if len(ipUsers) > 10 {
		return true, "同一IP用户数过多"
	}

	return false, ""
}

// recordBehavior 记录行为
func (d *AntiBotDetector) recordBehavior(behavior UserBehavior) {
	// 记录用户请求时间
	d.userRequests[behavior.UserID] = append(d.userRequests[behavior.UserID], behavior.Timestamp)
	
	// 记录IP请求时间
	d.ipRequests[behavior.IP] = append(d.ipRequests[behavior.IP], behavior.Timestamp)
	
	// 记录用户行为
	d.userBehaviors[behavior.UserID] = append(d.userBehaviors[behavior.UserID], behavior)
	
	// 限制记录数量
	if len(d.userBehaviors[behavior.UserID]) > 100 {
		d.userBehaviors[behavior.UserID] = d.userBehaviors[behavior.UserID][50:]
	}
}

// cleanup 定期清理过期数据
func (d *AntiBotDetector) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		
		now := time.Now()
		expireTime := 5 * time.Minute

		// 清理用户请求记录
		for userID, requests := range d.userRequests {
			validRequests := make([]time.Time, 0)
			for _, t := range requests {
				if now.Sub(t) <= expireTime {
					validRequests = append(validRequests, t)
				}
			}
			if len(validRequests) == 0 {
				delete(d.userRequests, userID)
			} else {
				d.userRequests[userID] = validRequests
			}
		}

		// 清理IP请求记录
		for ip, requests := range d.ipRequests {
			validRequests := make([]time.Time, 0)
			for _, t := range requests {
				if now.Sub(t) <= expireTime {
					validRequests = append(validRequests, t)
				}
			}
			if len(validRequests) == 0 {
				delete(d.ipRequests, ip)
			} else {
				d.ipRequests[ip] = validRequests
			}
		}

		// 清理用户行为记录
		for userID, behaviors := range d.userBehaviors {
			validBehaviors := make([]UserBehavior, 0)
			for _, b := range behaviors {
				if now.Sub(b.Timestamp) <= expireTime {
					validBehaviors = append(validBehaviors, b)
				}
			}
			if len(validBehaviors) == 0 {
				delete(d.userBehaviors, userID)
			} else {
				d.userBehaviors[userID] = validBehaviors
			}
		}

		d.mu.Unlock()
	}
}

// GetRiskScore 获取风险评分（0-100，分数越高风险越大）
func (d *AntiBotDetector) GetRiskScore(behavior UserBehavior) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	score := 0

	// 1. 请求频率评分（0-40分）
	if requests, exists := d.userRequests[behavior.UserID]; exists {
		windowSize := 10 * time.Second
		now := behavior.Timestamp
		
		count := 0
		for _, t := range requests {
			if now.Sub(t) <= windowSize {
				count++
			}
		}
		
		if count > 20 {
			score += 40
		} else if count > 10 {
			score += 20
		} else if count > 5 {
			score += 10
		}
	}

	// 2. 行为模式评分（0-30分）
	if behaviors, exists := d.userBehaviors[behavior.UserID]; exists && len(behaviors) >= 3 {
		if d.isRegularInterval(behaviors) {
			score += 30
		}
		if d.isDirectKill(behaviors, behavior) {
			score += 20
		}
	}

	// 3. UserAgent评分（0-20分）
	if d.isAbnormalUserAgent(behavior.UserAgent) {
		score += 20
	}

	// 4. IP评分（0-10分）
	if isAbnormal, _ := d.checkIPAbnormal(behavior); isAbnormal {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
