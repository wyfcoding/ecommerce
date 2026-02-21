package routeopt

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"

	pkgopt "github.com/wyfcoding/pkg/algos/optimization"
)

func sampleRouteData() (Location, []Location) {
	start := Location{ID: 0, Lat: 31.2304, Lon: 121.4737, Demand: 0}
	dests := []Location{
		{ID: 1, Lat: 31.2400, Lon: 121.4800, Demand: 6},
		{ID: 2, Lat: 31.2600, Lon: 121.5000, Demand: 7},
		{ID: 3, Lat: 31.2800, Lon: 121.5200, Demand: 5},
		{ID: 4, Lat: 31.2100, Lon: 121.4500, Demand: 4},
	}
	return start, dests
}

func toPkgLocation(in Location) pkgopt.Location {
	return pkgopt.Location{
		Lat:    in.Lat,
		Lon:    in.Lon,
		Demand: in.Demand,
		ID:     in.ID,
	}
}

func toPkgLocations(in []Location) []pkgopt.Location {
	out := make([]pkgopt.Location, 0, len(in))
	for _, d := range in {
		out = append(out, toPkgLocation(d))
	}
	return out
}

func routeSignature(ids []uint64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatUint(id, 10)
	}
	return strings.Join(parts, "-")
}

func localRouteToSignature(r Route) string {
	ids := make([]uint64, 0, len(r.Locations))
	for _, l := range r.Locations {
		ids = append(ids, l.ID)
	}
	return routeSignature(ids)
}

func pkgRouteToSignature(r pkgopt.Route) string {
	ids := make([]uint64, 0, len(r.Locations))
	for _, l := range r.Locations {
		ids = append(ids, l.ID)
	}
	return routeSignature(ids)
}

func TestRouteOptimizerConsistency(t *testing.T) {
	start, dests := sampleRouteData()

	local := NewRouteOptimizer()
	pkg := pkgopt.NewRouteOptimizer()

	localRoute := local.OptimizeRoute(start, dests)
	pkgRoute := pkg.OptimizeRoute(toPkgLocation(start), toPkgLocations(dests))

	if len(localRoute.Locations) != len(pkgRoute.Locations) {
		t.Fatalf("optimize route length mismatch: local=%d pkg=%d", len(localRoute.Locations), len(pkgRoute.Locations))
	}
	for i := range localRoute.Locations {
		if localRoute.Locations[i].ID != pkgRoute.Locations[i].ID {
			t.Fatalf("optimize route id mismatch at %d: local=%d pkg=%d", i, localRoute.Locations[i].ID, pkgRoute.Locations[i].ID)
		}
	}
	if math.Abs(localRoute.Distance-pkgRoute.Distance) > 1e-9 {
		t.Fatalf("optimize route distance mismatch: local=%f pkg=%f", localRoute.Distance, pkgRoute.Distance)
	}

	localVRP := local.ClarkeWrightVRP(start, dests, 25)
	pkgVRP := pkg.ClarkeWrightVRP(toPkgLocation(start), toPkgLocations(dests), 25)

	if len(localVRP) != len(pkgVRP) {
		t.Fatalf("vrp route count mismatch: local=%d pkg=%d", len(localVRP), len(pkgVRP))
	}

	localMap := make(map[string]float64, len(localVRP))
	pkgMap := make(map[string]float64, len(pkgVRP))
	for _, r := range localVRP {
		localMap[localRouteToSignature(r)] = r.Distance
	}
	for _, r := range pkgVRP {
		pkgMap[pkgRouteToSignature(r)] = r.Distance
	}

	localKeys := make([]string, 0, len(localMap))
	for k := range localMap {
		localKeys = append(localKeys, k)
	}
	sort.Strings(localKeys)

	for _, k := range localKeys {
		pkgDist, ok := pkgMap[k]
		if !ok {
			t.Fatalf("vrp route missing in pkg result: key=%s", k)
		}
		if math.Abs(localMap[k]-pkgDist) > 1e-9 {
			t.Fatalf("vrp distance mismatch for %s: local=%f pkg=%f", k, localMap[k], pkgDist)
		}
	}
}

func BenchmarkRouteOptimizerLocalOptimizeRoute(b *testing.B) {
	start, dests := sampleRouteData()
	optimizer := NewRouteOptimizer()
	for i := 0; i < b.N; i++ {
		_ = optimizer.OptimizeRoute(start, dests)
	}
}

func BenchmarkRouteOptimizerPkgOptimizeRoute(b *testing.B) {
	start, dests := sampleRouteData()
	optimizer := pkgopt.NewRouteOptimizer()
	pkgStart := toPkgLocation(start)
	pkgDests := toPkgLocations(dests)
	for i := 0; i < b.N; i++ {
		_ = optimizer.OptimizeRoute(pkgStart, pkgDests)
	}
}
