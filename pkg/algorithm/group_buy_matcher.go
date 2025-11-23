package algorithm

import (
	"sort"
	"time"
)

// GroupBuyGroup 拼团信息
type GroupBuyGroup struct {
	ID            uint64
	ActivityID    uint64
	LeaderID      uint64
	RequiredCount int // 需要人数
	CurrentCount  int // 当前人数
	CreatedAt     time.Time
	ExpireAt      time.Time
	Region        string // 地区
	Lat           float64
	Lon           float64
}

// GroupBuyMatcher 拼团匹配器
type GroupBuyMatcher struct{}

// NewGroupBuyMatcher 创建拼团匹配器
func NewGroupBuyMatcher() *GroupBuyMatcher {
	return &GroupBuyMatcher{}
}

// MatchStrategy 匹配策略
type MatchStrategy int

const (
	MatchStrategyFastest    MatchStrategy = iota + 1 // 最快成团
	MatchStrategyNearest                             // 最近地域
	MatchStrategyNewest                              // 最新创建
	MatchStrategyAlmostFull                          // 即将成团
)

// FindBestGroup 找到最佳拼团
func (m *GroupBuyMatcher) FindBestGroup(
	activityID uint64,
	userLat, userLon float64,
	userRegion string,
	groups []GroupBuyGroup,
	strategy MatchStrategy,
) *GroupBuyGroup {

	if len(groups) == 0 {
		return nil
	}

	// 过滤：只保留未满且未过期的团
	available := make([]GroupBuyGroup, 0)
	now := time.Now()

	for _, g := range groups {
		if g.ActivityID == activityID &&
			g.CurrentCount < g.RequiredCount &&
			g.ExpireAt.After(now) {
			available = append(available, g)
		}
	}

	if len(available) == 0 {
		return nil
	}

	switch strategy {
	case MatchStrategyFastest:
		return m.matchFastest(available)
	case MatchStrategyNearest:
		return m.matchNearest(available, userLat, userLon, userRegion)
	case MatchStrategyNewest:
		return m.matchNewest(available)
	case MatchStrategyAlmostFull:
		return m.matchAlmostFull(available)
	default:
		return m.matchFastest(available)
	}
}

// matchFastest 最快成团策略（优先选择即将成团的）
func (m *GroupBuyMatcher) matchFastest(groups []GroupBuyGroup) *GroupBuyGroup {
	sort.Slice(groups, func(i, j int) bool {
		// 按剩余人数升序，剩余时间降序
		remainI := groups[i].RequiredCount - groups[i].CurrentCount
		remainJ := groups[j].RequiredCount - groups[j].CurrentCount

		if remainI != remainJ {
			return remainI < remainJ
		}

		return groups[i].ExpireAt.After(groups[j].ExpireAt)
	})

	return &groups[0]
}

// matchNearest 最近地域策略
func (m *GroupBuyMatcher) matchNearest(
	groups []GroupBuyGroup,
	userLat, userLon float64,
	userRegion string,
) *GroupBuyGroup {

	// 优先匹配同地区
	sameRegion := make([]GroupBuyGroup, 0)
	otherRegion := make([]GroupBuyGroup, 0)

	for _, g := range groups {
		if g.Region == userRegion {
			sameRegion = append(sameRegion, g)
		} else {
			otherRegion = append(otherRegion, g)
		}
	}

	// 如果有同地区的，优先选择
	if len(sameRegion) > 0 {
		// 按距离排序
		sort.Slice(sameRegion, func(i, j int) bool {
			distI := haversineDistance(userLat, userLon, sameRegion[i].Lat, sameRegion[i].Lon)
			distJ := haversineDistance(userLat, userLon, sameRegion[j].Lat, sameRegion[j].Lon)
			return distI < distJ
		})
		return &sameRegion[0]
	}

	// 否则选择其他地区最近的
	if len(otherRegion) > 0 {
		sort.Slice(otherRegion, func(i, j int) bool {
			distI := haversineDistance(userLat, userLon, otherRegion[i].Lat, otherRegion[i].Lon)
			distJ := haversineDistance(userLat, userLon, otherRegion[j].Lat, otherRegion[j].Lon)
			return distI < distJ
		})
		return &otherRegion[0]
	}

	return nil
}

// matchNewest 最新创建策略
func (m *GroupBuyMatcher) matchNewest(groups []GroupBuyGroup) *GroupBuyGroup {
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].CreatedAt.After(groups[j].CreatedAt)
	})

	return &groups[0]
}

// matchAlmostFull 即将成团策略
func (m *GroupBuyMatcher) matchAlmostFull(groups []GroupBuyGroup) *GroupBuyGroup {
	sort.Slice(groups, func(i, j int) bool {
		// 按完成度降序
		rateI := float64(groups[i].CurrentCount) / float64(groups[i].RequiredCount)
		rateJ := float64(groups[j].CurrentCount) / float64(groups[j].RequiredCount)
		return rateI > rateJ
	})

	return &groups[0]
}

// SmartMatch 智能匹配（综合多个因素）
func (m *GroupBuyMatcher) SmartMatch(
	activityID uint64,
	userLat, userLon float64,
	userRegion string,
	groups []GroupBuyGroup,
) *GroupBuyGroup {

	if len(groups) == 0 {
		return nil
	}

	// 过滤可用的团
	available := make([]GroupBuyGroup, 0)
	now := time.Now()

	for _, g := range groups {
		if g.ActivityID == activityID &&
			g.CurrentCount < g.RequiredCount &&
			g.ExpireAt.After(now) {
			available = append(available, g)
		}
	}

	if len(available) == 0 {
		return nil
	}

	// 计算综合评分
	type groupScore struct {
		group *GroupBuyGroup
		score float64
	}

	scores := make([]groupScore, 0)

	for i := range available {
		g := &available[i]

		// 1. 完成度评分（0-1）
		completionRate := float64(g.CurrentCount) / float64(g.RequiredCount)

		// 2. 时间紧迫度评分（0-1）
		remainTime := g.ExpireAt.Sub(now).Minutes()
		urgencyScore := 1.0 / (1.0 + remainTime/60.0) // 时间越短分数越高

		// 3. 地域评分（0-1）
		regionScore := 0.0
		if g.Region == userRegion {
			regionScore = 1.0
		} else {
			distance := haversineDistance(userLat, userLon, g.Lat, g.Lon)
			regionScore = 1.0 / (1.0 + distance/10000.0) // 距离越近分数越高
		}

		// 综合评分：完成度40%，时间30%，地域30%
		totalScore := 0.4*completionRate + 0.3*urgencyScore + 0.3*regionScore

		scores = append(scores, groupScore{g, totalScore})
	}

	// 按评分降序排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	return scores[0].group
}

// BatchMatch 批量匹配（为多个用户同时匹配）
func (m *GroupBuyMatcher) BatchMatch(
	activityID uint64,
	users []struct {
		UserID uint64
		Lat    float64
		Lon    float64
		Region string
	},
	groups []GroupBuyGroup,
) map[uint64]uint64 { // userID -> groupID

	result := make(map[uint64]uint64)

	// 复制groups避免修改原数据
	availableGroups := make([]GroupBuyGroup, len(groups))
	copy(availableGroups, groups)

	// 为每个用户匹配
	for _, user := range users {
		bestGroup := m.SmartMatch(activityID, user.Lat, user.Lon, user.Region, availableGroups)

		if bestGroup != nil {
			result[user.UserID] = bestGroup.ID

			// 更新该团的当前人数
			for i := range availableGroups {
				if availableGroups[i].ID == bestGroup.ID {
					availableGroups[i].CurrentCount++
					break
				}
			}
		}
	}

	return result
}
