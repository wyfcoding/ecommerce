package matcher

import (
	"testing"
	"time"

	pkgopt "github.com/wyfcoding/pkg/algos/optimization"
)

func sampleGroupBuyGroups(base time.Time) []GroupBuyGroup {
	return []GroupBuyGroup{
		{
			ID:            101,
			ActivityID:    9001,
			LeaderID:      1,
			RequiredCount: 5,
			CurrentCount:  4,
			CreatedAt:     base.Add(-30 * time.Minute),
			ExpireAt:      base.Add(2 * time.Hour),
			Region:        "sh",
			Lat:           31.2304,
			Lon:           121.4737,
		},
		{
			ID:            102,
			ActivityID:    9001,
			LeaderID:      2,
			RequiredCount: 6,
			CurrentCount:  3,
			CreatedAt:     base.Add(-10 * time.Minute),
			ExpireAt:      base.Add(3 * time.Hour),
			Region:        "sh",
			Lat:           31.2400,
			Lon:           121.4800,
		},
		{
			ID:            103,
			ActivityID:    9001,
			LeaderID:      3,
			RequiredCount: 8,
			CurrentCount:  7,
			CreatedAt:     base.Add(-50 * time.Minute),
			ExpireAt:      base.Add(90 * time.Minute),
			Region:        "bj",
			Lat:           39.9042,
			Lon:           116.4074,
		},
		{
			ID:            104,
			ActivityID:    9001,
			LeaderID:      4,
			RequiredCount: 4,
			CurrentCount:  1,
			CreatedAt:     base.Add(-5 * time.Minute),
			ExpireAt:      base.Add(4 * time.Hour),
			Region:        "gz",
			Lat:           23.1291,
			Lon:           113.2644,
		},
		{
			ID:            201,
			ActivityID:    9999,
			LeaderID:      9,
			RequiredCount: 5,
			CurrentCount:  2,
			CreatedAt:     base.Add(-20 * time.Minute),
			ExpireAt:      base.Add(2 * time.Hour),
			Region:        "sh",
			Lat:           31.2200,
			Lon:           121.4700,
		},
	}
}

func toPkgGroupBuyGroups(in []GroupBuyGroup) []pkgopt.GroupBuyGroup {
	out := make([]pkgopt.GroupBuyGroup, 0, len(in))
	for _, g := range in {
		out = append(out, pkgopt.GroupBuyGroup{
			CreatedAt:     g.CreatedAt,
			ExpireAt:      g.ExpireAt,
			Region:        g.Region,
			Lat:           g.Lat,
			Lon:           g.Lon,
			ID:            g.ID,
			ActivityID:    g.ActivityID,
			LeaderID:      g.LeaderID,
			RequiredCount: g.RequiredCount,
			CurrentCount:  g.CurrentCount,
		})
	}
	return out
}

func TestGroupBuyMatcherConsistency(t *testing.T) {
	base := time.Now()
	groups := sampleGroupBuyGroups(base)
	pkgGroups := toPkgGroupBuyGroups(groups)

	local := NewGroupBuyMatcher()
	pkg := pkgopt.NewGroupBuyMatcher()

	strategies := []MatchStrategy{
		MatchStrategyFastest,
		MatchStrategyNearest,
		MatchStrategyNewest,
		MatchStrategyAlmostFull,
	}

	for _, st := range strategies {
		localBest := local.FindBestGroup(9001, 31.2310, 121.4740, "sh", groups, st)
		pkgBest := pkg.FindBestGroup(9001, 31.2310, 121.4740, "sh", pkgGroups, pkgopt.MatchStrategy(st))
		if (localBest == nil) != (pkgBest == nil) {
			t.Fatalf("strategy %v nil mismatch: local=%v pkg=%v", st, localBest == nil, pkgBest == nil)
		}
		if localBest != nil && localBest.ID != pkgBest.ID {
			t.Fatalf("strategy %v id mismatch: local=%d pkg=%d", st, localBest.ID, pkgBest.ID)
		}
	}

	localSmart := local.SmartMatch(9001, 31.2310, 121.4740, "sh", groups)
	pkgSmart := pkg.SmartMatch(9001, 31.2310, 121.4740, "sh", pkgGroups)
	if (localSmart == nil) != (pkgSmart == nil) {
		t.Fatalf("smart nil mismatch: local=%v pkg=%v", localSmart == nil, pkgSmart == nil)
	}
	if localSmart != nil && localSmart.ID != pkgSmart.ID {
		t.Fatalf("smart id mismatch: local=%d pkg=%d", localSmart.ID, pkgSmart.ID)
	}

	localUsers := []struct {
		Region string
		UserID uint64
		Lat    float64
		Lon    float64
	}{
		{Region: "sh", UserID: 10001, Lat: 31.2300, Lon: 121.4730},
		{Region: "sh", UserID: 10002, Lat: 31.2290, Lon: 121.4720},
		{Region: "bj", UserID: 10003, Lat: 39.9000, Lon: 116.4000},
	}
	pkgUsers := []struct {
		Region string
		UserID uint64
		Lat    float64
		Lon    float64
	}{
		{Region: "sh", UserID: 10001, Lat: 31.2300, Lon: 121.4730},
		{Region: "sh", UserID: 10002, Lat: 31.2290, Lon: 121.4720},
		{Region: "bj", UserID: 10003, Lat: 39.9000, Lon: 116.4000},
	}

	localBatch := local.BatchMatch(9001, localUsers, groups)
	pkgBatch := pkg.BatchMatch(9001, pkgUsers, pkgGroups)

	if len(localBatch) != len(pkgBatch) {
		t.Fatalf("batch size mismatch: local=%d pkg=%d", len(localBatch), len(pkgBatch))
	}
	for userID, localGroupID := range localBatch {
		pkgGroupID, ok := pkgBatch[userID]
		if !ok {
			t.Fatalf("batch user missing in pkg result: user=%d", userID)
		}
		if pkgGroupID != localGroupID {
			t.Fatalf("batch mismatch for user=%d: local=%d pkg=%d", userID, localGroupID, pkgGroupID)
		}
	}
}

func BenchmarkGroupBuyMatcherLocalSmartMatch(b *testing.B) {
	groups := sampleGroupBuyGroups(time.Now())
	matcher := NewGroupBuyMatcher()
	for i := 0; i < b.N; i++ {
		_ = matcher.SmartMatch(9001, 31.2310, 121.4740, "sh", groups)
	}
}

func BenchmarkGroupBuyMatcherPkgSmartMatch(b *testing.B) {
	groups := toPkgGroupBuyGroups(sampleGroupBuyGroups(time.Now()))
	matcher := pkgopt.NewGroupBuyMatcher()
	for i := 0; i < b.N; i++ {
		_ = matcher.SmartMatch(9001, 31.2310, 121.4740, "sh", groups)
	}
}
