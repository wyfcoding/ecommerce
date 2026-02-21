package allocator

import (
	"math"
	"sort"
	"testing"

	pkgopt "github.com/wyfcoding/pkg/algos/optimization"
)

type normalizedItem struct {
	skuID uint64
	qty   int32
}

type normalizedAllocation struct {
	warehouseID uint64
	totalCost   int64
	distance    float64
	items       []normalizedItem
}

func sampleAllocatorInput() (float64, float64, []OrderItem, map[uint64]map[uint64]*WarehouseInfo) {
	userLat, userLon := 31.2304, 121.4737
	items := []OrderItem{
		{SkuID: 1001, Quantity: 4},
		{SkuID: 1002, Quantity: 3},
	}
	warehouses := map[uint64]map[uint64]*WarehouseInfo{
		1: {
			1001: {ID: 1, Lat: 31.2400, Lon: 121.4800, ShipCost: 5, Stock: 3, Priority: 1},
			1002: {ID: 1, Lat: 31.2400, Lon: 121.4800, ShipCost: 6, Stock: 2, Priority: 1},
		},
		2: {
			1001: {ID: 2, Lat: 31.2800, Lon: 121.5000, ShipCost: 4, Stock: 5, Priority: 2},
			1002: {ID: 2, Lat: 31.2800, Lon: 121.5000, ShipCost: 5, Stock: 5, Priority: 2},
		},
		3: {
			1002: {ID: 3, Lat: 30.9000, Lon: 121.3000, ShipCost: 9, Stock: 8, Priority: 3},
		},
	}
	return userLat, userLon, items, warehouses
}

func toPkgOrderItems(in []OrderItem) []pkgopt.OrderItem {
	out := make([]pkgopt.OrderItem, 0, len(in))
	for _, it := range in {
		out = append(out, pkgopt.OrderItem{SkuID: it.SkuID, Quantity: it.Quantity})
	}
	return out
}

func toPkgWarehouses(in map[uint64]map[uint64]*WarehouseInfo) map[uint64]map[uint64]*pkgopt.WarehouseInfo {
	out := make(map[uint64]map[uint64]*pkgopt.WarehouseInfo, len(in))
	for wID, skuMap := range in {
		copied := make(map[uint64]*pkgopt.WarehouseInfo, len(skuMap))
		for skuID, info := range skuMap {
			copied[skuID] = &pkgopt.WarehouseInfo{
				Lat:      info.Lat,
				Lon:      info.Lon,
				ID:       info.ID,
				ShipCost: info.ShipCost,
				Stock:    info.Stock,
				Priority: info.Priority,
			}
		}
		out[wID] = copied
	}
	return out
}

func normalizeLocalAllocations(in []AllocationResult) []normalizedAllocation {
	out := make([]normalizedAllocation, 0, len(in))
	for _, a := range in {
		items := make([]normalizedItem, 0, len(a.Items))
		for _, it := range a.Items {
			items = append(items, normalizedItem{skuID: it.SkuID, qty: it.Quantity})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].skuID < items[j].skuID })

		out = append(out, normalizedAllocation{
			warehouseID: a.WarehouseID,
			totalCost:   a.TotalCost,
			distance:    a.Distance,
			items:       items,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].warehouseID < out[j].warehouseID })
	return out
}

func normalizePkgAllocations(in []pkgopt.AllocationResult) []normalizedAllocation {
	out := make([]normalizedAllocation, 0, len(in))
	for _, a := range in {
		items := make([]normalizedItem, 0, len(a.Items))
		for _, it := range a.Items {
			items = append(items, normalizedItem{skuID: it.SkuID, qty: it.Quantity})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].skuID < items[j].skuID })

		out = append(out, normalizedAllocation{
			warehouseID: a.WarehouseID,
			totalCost:   a.TotalCost,
			distance:    a.Distance,
			items:       items,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].warehouseID < out[j].warehouseID })
	return out
}

func assertAllocationsEqual(t *testing.T, title string, local, pkg []normalizedAllocation) {
	t.Helper()

	if len(local) != len(pkg) {
		t.Fatalf("%s allocation count mismatch: local=%d pkg=%d", title, len(local), len(pkg))
	}
	for i := range local {
		if local[i].warehouseID != pkg[i].warehouseID {
			t.Fatalf("%s warehouse mismatch at %d: local=%d pkg=%d", title, i, local[i].warehouseID, pkg[i].warehouseID)
		}
		if local[i].totalCost != pkg[i].totalCost {
			t.Fatalf("%s totalCost mismatch at %d: local=%d pkg=%d", title, i, local[i].totalCost, pkg[i].totalCost)
		}
		if math.Abs(local[i].distance-pkg[i].distance) > 1e-9 {
			t.Fatalf("%s distance mismatch at %d: local=%f pkg=%f", title, i, local[i].distance, pkg[i].distance)
		}
		if len(local[i].items) != len(pkg[i].items) {
			t.Fatalf("%s items count mismatch at %d: local=%d pkg=%d", title, i, len(local[i].items), len(pkg[i].items))
		}
		for j := range local[i].items {
			if local[i].items[j] != pkg[i].items[j] {
				t.Fatalf("%s item mismatch at %d:%d: local=%+v pkg=%+v", title, i, j, local[i].items[j], pkg[i].items[j])
			}
		}
	}
}

func TestWarehouseAllocatorConsistency(t *testing.T) {
	userLat, userLon, items, warehouses := sampleAllocatorInput()

	local := NewWarehouseAllocator()
	pkg := pkgopt.NewWarehouseAllocator()

	localOptimal := local.AllocateOptimal(userLat, userLon, items, warehouses)
	pkgOptimal := pkg.AllocateOptimal(userLat, userLon, toPkgOrderItems(items), toPkgWarehouses(warehouses))
	assertAllocationsEqual(t, "optimal", normalizeLocalAllocations(localOptimal), normalizePkgAllocations(pkgOptimal))

	localByDistance := local.AllocateByDistance(userLat, userLon, items, warehouses)
	pkgByDistance := pkg.AllocateByDistance(userLat, userLon, toPkgOrderItems(items), toPkgWarehouses(warehouses))
	assertAllocationsEqual(t, "distance", normalizeLocalAllocations(localByDistance), normalizePkgAllocations(pkgByDistance))
}

func BenchmarkWarehouseAllocatorLocalOptimal(b *testing.B) {
	userLat, userLon, items, warehouses := sampleAllocatorInput()
	allocator := NewWarehouseAllocator()
	for i := 0; i < b.N; i++ {
		_ = allocator.AllocateOptimal(userLat, userLon, items, warehouses)
	}
}

func BenchmarkWarehouseAllocatorPkgOptimal(b *testing.B) {
	userLat, userLon, items, warehouses := sampleAllocatorInput()
	allocator := pkgopt.NewWarehouseAllocator()
	pkgItems := toPkgOrderItems(items)
	pkgWarehouses := toPkgWarehouses(warehouses)
	for i := 0; i < b.N; i++ {
		_ = allocator.AllocateOptimal(userLat, userLon, pkgItems, pkgWarehouses)
	}
}
