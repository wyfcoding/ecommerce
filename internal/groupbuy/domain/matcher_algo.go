// 变更说明：
// 从 pkg/algos/optimization/group_buy_matcher.go 迁移。
// 实现了拼团活动的多种匹配策略（快速成团、邻近地域、智能评分等）。
package domain

import (
	"math"
	"slices"
	"time"

	algomath "github.com/wyfcoding/pkg/algos/math"
)

type MatchStrategy int

const (
	MatchStrategyFastest MatchStrategy = iota + 1
	MatchStrategyNearest
	MatchStrategyNewest
	MatchStrategyAlmostFull
)

type GroupBuyGroup struct {
	CreatedAt     time.Time
	ExpireAt      time.Time
	Region        string
	Lat           float64
	Lon           float64
	ID            uint64
	ActivityID    uint64
	LeaderID      uint64
	RequiredCount int
	CurrentCount  int
}

type GroupBuyMatcher struct{}

func NewGroupBuyMatcher() *GroupBuyMatcher {
	return &GroupBuyMatcher{}
}

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
	available := make([]GroupBuyGroup, 0)
	now := time.Now()
	for _, g := range groups {
		if g.ActivityID == activityID && g.CurrentCount < g.RequiredCount && g.ExpireAt.After(now) {
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

func (m *GroupBuyMatcher) matchFastest(groups []GroupBuyGroup) *GroupBuyGroup {
	slices.SortFunc(groups, func(a, b GroupBuyGroup) int {
		remA := a.RequiredCount - a.CurrentCount
		remB := b.RequiredCount - b.CurrentCount
		if remA != remB {
			return remA - remB
		}
		if a.ExpireAt.After(b.ExpireAt) {
			return -1
		}
		return 1
	})
	return &groups[0]
}

func (m *GroupBuyMatcher) matchAlmostFull(groups []GroupBuyGroup) *GroupBuyGroup {
	slices.SortFunc(groups, func(a, b GroupBuyGroup) int {
		rA := float64(a.CurrentCount) / float64(a.RequiredCount)
		rB := float64(b.CurrentCount) / float64(b.RequiredCount)
		if rA > rB {
			return -1
		}
		return 1
	})
	return &groups[0]
}

func (m *GroupBuyMatcher) matchNewest(groups []GroupBuyGroup) *GroupBuyGroup {
	slices.SortFunc(groups, func(a, b GroupBuyGroup) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		return 1
	})
	return &groups[0]
}

func (m *GroupBuyMatcher) matchNearest(groups []GroupBuyGroup, lat, lon float64, region string) *GroupBuyGroup {
	var best *GroupBuyGroup
	minDist := math.MaxFloat64
	for i := range groups {
		dist := algomath.HaversineDistance(lat, lon, groups[i].Lat, groups[i].Lon)
		if groups[i].Region == region {
			dist *= 0.5 // 地域优先权
		}
		if dist < minDist {
			minDist = dist
			best = &groups[i]
		}
	}
	return best
}
