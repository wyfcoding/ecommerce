package algorithm

import (
	"container/heap"
	"math"
)

// Point 地理坐标点
type Point struct {
	Lat float64
	Lon float64
}

// DeliveryPoint 配送点
type DeliveryPoint struct {
	ID       uint64
	Point    Point
	Priority int // 优先级（1-10，数字越大越优先）
	TimeWindow struct {
		Start int // 时间窗口开始（分钟，0-1440）
		End   int // 时间窗口结束
	}
}

// RouteOptimizer 路径优化器
type RouteOptimizer struct{}

// NewRouteOptimizer 创建路径优化器
func NewRouteOptimizer() *RouteOptimizer {
	return &RouteOptimizer{}
}

// Graph 图结构（用于Dijkstra算法）
type Graph struct {
	nodes map[uint64]*Node
	edges map[uint64]map[uint64]float64 // from -> to -> distance
}

type Node struct {
	ID    uint64
	Point Point
}

// NewGraph 创建图
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[uint64]*Node),
		edges: make(map[uint64]map[uint64]float64),
	}
}

// AddNode 添加节点
func (g *Graph) AddNode(id uint64, point Point) {
	g.nodes[id] = &Node{ID: id, Point: point}
	if g.edges[id] == nil {
		g.edges[id] = make(map[uint64]float64)
	}
}

// AddEdge 添加边
func (g *Graph) AddEdge(from, to uint64, distance float64) {
	if g.edges[from] == nil {
		g.edges[from] = make(map[uint64]float64)
	}
	g.edges[from][to] = distance
}

// Dijkstra 最短路径算法
func (ro *RouteOptimizer) Dijkstra(graph *Graph, start, end uint64) ([]uint64, float64) {
	dist := make(map[uint64]float64)
	prev := make(map[uint64]uint64)
	visited := make(map[uint64]bool)

	// 初始化距离
	for id := range graph.nodes {
		dist[id] = math.Inf(1)
	}
	dist[start] = 0

	// 优先队列
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Item{value: start, priority: 0})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		current := item.value.(uint64)

		if visited[current] {
			continue
		}
		visited[current] = true

		if current == end {
			break
		}

		// 遍历邻居
		for neighbor, edgeDist := range graph.edges[current] {
			if visited[neighbor] {
				continue
			}

			newDist := dist[current] + edgeDist
			if newDist < dist[neighbor] {
				dist[neighbor] = newDist
				prev[neighbor] = current
				heap.Push(pq, &Item{value: neighbor, priority: newDist})
			}
		}
	}

	// 重建路径
	if dist[end] == math.Inf(1) {
		return nil, -1 // 无法到达
	}

	path := make([]uint64, 0)
	for at := end; at != start; at = prev[at] {
		path = append([]uint64{at}, path...)
	}
	path = append([]uint64{start}, path...)

	return path, dist[end]
}

// AStar A*算法（启发式搜索）
func (ro *RouteOptimizer) AStar(graph *Graph, start, end uint64) ([]uint64, float64) {
	startNode := graph.nodes[start]
	endNode := graph.nodes[end]

	if startNode == nil || endNode == nil {
		return nil, -1
	}

	gScore := make(map[uint64]float64) // 从起点到当前点的实际距离
	fScore := make(map[uint64]float64) // gScore + 启发式估计距离
	prev := make(map[uint64]uint64)
	visited := make(map[uint64]bool)

	// 初始化
	for id := range graph.nodes {
		gScore[id] = math.Inf(1)
		fScore[id] = math.Inf(1)
	}
	gScore[start] = 0
	fScore[start] = ro.heuristic(startNode.Point, endNode.Point)

	// 优先队列
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Item{value: start, priority: fScore[start]})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		current := item.value.(uint64)

		if current == end {
			break
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		// 遍历邻居
		for neighbor, edgeDist := range graph.edges[current] {
			if visited[neighbor] {
				continue
			}

			tentativeGScore := gScore[current] + edgeDist

			if tentativeGScore < gScore[neighbor] {
				prev[neighbor] = current
				gScore[neighbor] = tentativeGScore
				fScore[neighbor] = gScore[neighbor] + ro.heuristic(graph.nodes[neighbor].Point, endNode.Point)
				heap.Push(pq, &Item{value: neighbor, priority: fScore[neighbor]})
			}
		}
	}

	// 重建路径
	if gScore[end] == math.Inf(1) {
		return nil, -1
	}

	path := make([]uint64, 0)
	for at := end; at != start; at = prev[at] {
		path = append([]uint64{at}, path...)
	}
	path = append([]uint64{start}, path...)

	return path, gScore[end]
}

// heuristic 启发式函数（欧几里得距离）
func (ro *RouteOptimizer) heuristic(p1, p2 Point) float64 {
	return haversineDistance(p1.Lat, p1.Lon, p2.Lat, p2.Lon)
}

// TSP 旅行商问题（贪心近似算法）
func (ro *RouteOptimizer) TSPGreedy(points []DeliveryPoint, startPoint Point) []uint64 {
	if len(points) == 0 {
		return nil
	}

	visited := make(map[uint64]bool)
	route := make([]uint64, 0, len(points))
	
	// 从起点开始
	currentPoint := startPoint
	
	for len(route) < len(points) {
		// 找最近的未访问点
		var nearestID uint64
		minDist := math.Inf(1)
		
		for _, p := range points {
			if visited[p.ID] {
				continue
			}
			
			dist := haversineDistance(currentPoint.Lat, currentPoint.Lon, p.Point.Lat, p.Point.Lon)
			
			// 考虑优先级（优先级高的点距离打折）
			adjustedDist := dist / float64(p.Priority)
			
			if adjustedDist < minDist {
				minDist = adjustedDist
				nearestID = p.ID
			}
		}
		
		if nearestID == 0 {
			break
		}
		
		route = append(route, nearestID)
		visited[nearestID] = true
		
		// 更新当前位置
		for _, p := range points {
			if p.ID == nearestID {
				currentPoint = p.Point
				break
			}
		}
	}
	
	return route
}

// TSPWithTimeWindow TSP问题（考虑时间窗口）
func (ro *RouteOptimizer) TSPWithTimeWindow(points []DeliveryPoint, startPoint Point, currentTime int) []uint64 {
	if len(points) == 0 {
		return nil
	}

	visited := make(map[uint64]bool)
	route := make([]uint64, 0, len(points))
	
	currentPoint := startPoint
	time := currentTime
	
	for len(route) < len(points) {
		var bestID uint64
		bestScore := math.Inf(-1)
		
		for _, p := range points {
			if visited[p.ID] {
				continue
			}
			
			dist := haversineDistance(currentPoint.Lat, currentPoint.Lon, p.Point.Lat, p.Point.Lon)
			travelTime := int(dist / 500) // 假设速度500米/分钟
			arrivalTime := time + travelTime
			
			// 检查时间窗口
			if arrivalTime > p.TimeWindow.End {
				continue // 无法在时间窗口内到达
			}
			
			// 计算评分：距离越近越好，优先级越高越好，时间窗口越紧越优先
			distScore := 1.0 / (1.0 + dist/1000.0)
			priorityScore := float64(p.Priority) / 10.0
			
			// 时间紧迫度
			urgency := 0.0
			if arrivalTime < p.TimeWindow.Start {
				// 到达太早，需要等待
				waitTime := p.TimeWindow.Start - arrivalTime
				urgency = 1.0 / (1.0 + float64(waitTime)/60.0)
			} else {
				// 在时间窗口内
				remaining := p.TimeWindow.End - arrivalTime
				urgency = 1.0 / (1.0 + float64(remaining)/60.0)
			}
			
			score := 0.4*distScore + 0.3*priorityScore + 0.3*urgency
			
			if score > bestScore {
				bestScore = score
				bestID = p.ID
			}
		}
		
		if bestID == 0 {
			break
		}
		
		route = append(route, bestID)
		visited[bestID] = true
		
		// 更新当前位置和时间
		for _, p := range points {
			if p.ID == bestID {
				dist := haversineDistance(currentPoint.Lat, currentPoint.Lon, p.Point.Lat, p.Point.Lon)
				travelTime := int(dist / 500)
				time += travelTime
				
				// 如果到达太早，等待到时间窗口开始
				if time < p.TimeWindow.Start {
					time = p.TimeWindow.Start
				}
				
				currentPoint = p.Point
				break
			}
		}
	}
	
	return route
}

// OptimizeRoute 2-opt优化（改进TSP解）
func (ro *RouteOptimizer) OptimizeRoute(route []uint64, points map[uint64]Point) []uint64 {
	if len(route) < 4 {
		return route
	}

	improved := true
	bestRoute := make([]uint64, len(route))
	copy(bestRoute, route)

	for improved {
		improved = false

		for i := 1; i < len(bestRoute)-2; i++ {
			for j := i + 1; j < len(bestRoute)-1; j++ {
				// 计算当前距离
				currentDist := ro.routeDistance(bestRoute, points)

				// 尝试2-opt交换
				newRoute := ro.twoOptSwap(bestRoute, i, j)
				newDist := ro.routeDistance(newRoute, points)

				if newDist < currentDist {
					bestRoute = newRoute
					improved = true
				}
			}
		}
	}

	return bestRoute
}

// twoOptSwap 2-opt交换
func (ro *RouteOptimizer) twoOptSwap(route []uint64, i, j int) []uint64 {
	newRoute := make([]uint64, len(route))
	
	// 复制前半部分
	copy(newRoute[:i], route[:i])
	
	// 反转中间部分
	for k := i; k <= j; k++ {
		newRoute[k] = route[j-(k-i)]
	}
	
	// 复制后半部分
	copy(newRoute[j+1:], route[j+1:])
	
	return newRoute
}

// routeDistance 计算路径总距离
func (ro *RouteOptimizer) routeDistance(route []uint64, points map[uint64]Point) float64 {
	if len(route) < 2 {
		return 0
	}

	totalDist := 0.0
	for i := 0; i < len(route)-1; i++ {
		p1 := points[route[i]]
		p2 := points[route[i+1]]
		totalDist += haversineDistance(p1.Lat, p1.Lon, p2.Lat, p2.Lon)
	}

	return totalDist
}

// PriorityQueue 优先队列（用于Dijkstra和A*）
type PriorityQueue []*Item

type Item struct {
	value    interface{}
	priority float64
	index    int
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
