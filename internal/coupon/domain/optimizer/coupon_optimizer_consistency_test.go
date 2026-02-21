package optimizer

import (
	"sort"
	"testing"

	pkgopt "github.com/wyfcoding/pkg/algos/optimization"
)

func sampleCoupons() []Coupon {
	return []Coupon{
		{ID: 1, Type: CouponTypeCash, ReductionAmount: 1200, Threshold: 0, CanStack: true, Priority: 1},
		{ID: 2, Type: CouponTypeReduction, ReductionAmount: 2000, Threshold: 15000, CanStack: true, Priority: 2},
		{ID: 3, Type: CouponTypeDiscount, DiscountRate: 0.9, MaxDiscount: 1800, Threshold: 10000, CanStack: true, Priority: 3},
		{ID: 4, Type: CouponTypeCash, ReductionAmount: 800, Threshold: 0, CanStack: false, Priority: 4},
	}
}

func toPkgCoupons(in []Coupon) []pkgopt.Coupon {
	out := make([]pkgopt.Coupon, 0, len(in))
	for _, c := range in {
		out = append(out, pkgopt.Coupon{
			ID:              c.ID,
			DiscountRate:    c.DiscountRate,
			Threshold:       c.Threshold,
			ReductionAmount: c.ReductionAmount,
			MaxDiscount:     c.MaxDiscount,
			Type:            pkgopt.CouponType(c.Type),
			Priority:        c.Priority,
			CanStack:        c.CanStack,
		})
	}
	return out
}

func normalizeIDs(ids []uint64) []uint64 {
	cp := append([]uint64(nil), ids...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	return cp
}

func TestCouponOptimizerConsistency(t *testing.T) {
	price := int64(20000)
	local := NewCouponOptimizer()
	pkg := pkgopt.NewCouponOptimizer()
	coupons := sampleCoupons()
	pkgCoupons := toPkgCoupons(coupons)

	lIDs, lFinal, lDiscount := local.OptimalCombination(price, coupons)
	pIDs, pFinal, pDiscount := pkg.OptimalCombination(price, pkgCoupons)

	if lFinal != pFinal || lDiscount != pDiscount {
		t.Fatalf("optimal mismatch: local=(%d,%d) pkg=(%d,%d)", lFinal, lDiscount, pFinal, pDiscount)
	}

	lIDs = normalizeIDs(lIDs)
	pIDs = normalizeIDs(pIDs)
	if len(lIDs) != len(pIDs) {
		t.Fatalf("optimal id count mismatch: local=%v pkg=%v", lIDs, pIDs)
	}
	for i := range lIDs {
		if lIDs[i] != pIDs[i] {
			t.Fatalf("optimal ids mismatch: local=%v pkg=%v", lIDs, pIDs)
		}
	}

	lIDs, lFinal, lDiscount = local.GreedyOptimization(price, coupons)
	pIDs, pFinal, pDiscount = pkg.GreedyOptimization(price, pkgCoupons)
	if lFinal != pFinal || lDiscount != pDiscount {
		t.Fatalf("greedy mismatch: local=(%d,%d) pkg=(%d,%d)", lFinal, lDiscount, pFinal, pDiscount)
	}
	lIDs = normalizeIDs(lIDs)
	pIDs = normalizeIDs(pIDs)
	if len(lIDs) != len(pIDs) {
		t.Fatalf("greedy id count mismatch: local=%v pkg=%v", lIDs, pIDs)
	}
	for i := range lIDs {
		if lIDs[i] != pIDs[i] {
			t.Fatalf("greedy ids mismatch: local=%v pkg=%v", lIDs, pIDs)
		}
	}
}

func BenchmarkCouponOptimizerLocalOptimal(b *testing.B) {
	opt := NewCouponOptimizer()
	coupons := sampleCoupons()
	for i := 0; i < b.N; i++ {
		_, _, _ = opt.OptimalCombination(20000, coupons)
	}
}

func BenchmarkCouponOptimizerPkgOptimal(b *testing.B) {
	opt := pkgopt.NewCouponOptimizer()
	coupons := toPkgCoupons(sampleCoupons())
	for i := 0; i < b.N; i++ {
		_, _, _ = opt.OptimalCombination(20000, coupons)
	}
}
