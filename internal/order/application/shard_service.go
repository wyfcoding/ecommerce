package application

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

type ShardRouter struct {
	config      *domain.ShardConfig
	shardNodes  map[int]*ShardNode
	loadMonitor *ShardLoadMonitor
	mu          sync.RWMutex
}

type ShardNode struct {
	Index      int
	Address    string
	Weight     int
	IsActive   bool
	LoadFactor float64
}

type ShardLoadMonitor struct {
	shardLoads map[int]int64
	mu         sync.RWMutex
}

func NewShardLoadMonitor() *ShardLoadMonitor {
	return &ShardLoadMonitor{
		shardLoads: make(map[int]int64),
	}
}

func (m *ShardLoadMonitor) IncrementLoad(shardIndex int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shardLoads[shardIndex]++
}

func (m *ShardLoadMonitor) DecrementLoad(shardIndex int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shardLoads[shardIndex] > 0 {
		m.shardLoads[shardIndex]--
	}
}

func (m *ShardLoadMonitor) GetLoad(shardIndex int) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shardLoads[shardIndex]
}

func (m *ShardLoadMonitor) GetTotalLoad() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var total int64
	for _, load := range m.shardLoads {
		total += load
	}
	return total
}

func NewShardRouter(config *domain.ShardConfig) *ShardRouter {
	return &ShardRouter{
		config:      config,
		shardNodes:  make(map[int]*ShardNode),
		loadMonitor: NewShardLoadMonitor(),
	}
}

func (r *ShardRouter) AddShardNode(index int, address string, weight int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shardNodes[index] = &ShardNode{
		Index:    index,
		Address:  address,
		Weight:   weight,
		IsActive: true,
	}
}

func (r *ShardRouter) RemoveShardNode(index int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.shardNodes, index)
}

func (r *ShardRouter) SetNodeActive(index int, active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if node, exists := r.shardNodes[index]; exists {
		node.IsActive = active
	}
}

func (r *ShardRouter) Route(order *domain.Order) (*ShardNode, error) {
	shardIndex := order.GetShardIndex(r.config)

	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.shardNodes[shardIndex]
	if !exists {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	if !node.IsActive {
		return nil, fmt.Errorf("shard %d is not active", shardIndex)
	}

	return node, nil
}

func (r *ShardRouter) RouteByUserID(userID uint64) (*ShardNode, error) {
	shardIndex := r.config.GetShardIndex(userID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.shardNodes[shardIndex]
	if !exists {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	if !node.IsActive {
		return nil, fmt.Errorf("shard %d is not active", shardIndex)
	}

	return node, nil
}

func (r *ShardRouter) RouteByOrderID(orderID uint64) (*ShardNode, error) {
	shardIndex := r.config.GetShardIndex(orderID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.shardNodes[shardIndex]
	if !exists {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	return node, nil
}

func (r *ShardRouter) RouteByHash(key string) (*ShardNode, error) {
	shardIndex := r.config.GetShardIndexByString(key)

	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.shardNodes[shardIndex]
	if !exists {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	return node, nil
}

func (r *ShardRouter) GetShardIndex(order *domain.Order) int {
	return order.GetShardIndex(r.config)
}

func (r *ShardRouter) GetShardNode(index int) (*ShardNode, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	node, exists := r.shardNodes[index]
	return node, exists
}

func (r *ShardRouter) GetAllActiveShards() []*ShardNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*ShardNode, 0)
	for _, node := range r.shardNodes {
		if node.IsActive {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (r *ShardRouter) GetShardStats() map[int]*ShardStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[int]*ShardStats)
	for index, node := range r.shardNodes {
		stats[index] = &ShardStats{
			Index:      index,
			Address:    node.Address,
			IsActive:   node.IsActive,
			LoadFactor: r.loadMonitor.GetLoad(index),
		}
	}
	return stats
}

type ShardStats struct {
	Index      int    `json:"index"`
	Address    string `json:"address"`
	IsActive   bool   `json:"is_active"`
	LoadFactor int64  `json:"load_factor"`
}

type ConsistentHashRouter struct {
	virtualNodes int
	ring         map[uint32]*ShardNode
	sortedKeys   []uint32
	mu           sync.RWMutex
}

func NewConsistentHashRouter(virtualNodes int) *ConsistentHashRouter {
	return &ConsistentHashRouter{
		virtualNodes: virtualNodes,
		ring:         make(map[uint32]*ShardNode),
	}
}

func (r *ConsistentHashRouter) AddNode(node *ShardNode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.virtualNodes; i++ {
		key := r.hashKey(fmt.Sprintf("%d#%d", node.Index, i))
		r.ring[key] = node
	}

	r.updateSortedKeys()
}

func (r *ConsistentHashRouter) RemoveNode(index int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, node := range r.ring {
		if node.Index == index {
			delete(r.ring, key)
		}
	}

	r.updateSortedKeys()
}

func (r *ConsistentHashRouter) GetNode(key string) (*ShardNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ring) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	hash := r.hashKey(key)

	for _, k := range r.sortedKeys {
		if k >= hash {
			return r.ring[k], nil
		}
	}

	return r.ring[r.sortedKeys[0]], nil
}

func (r *ConsistentHashRouter) hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (r *ConsistentHashRouter) updateSortedKeys() {
	r.sortedKeys = make([]uint32, 0, len(r.ring))
	for k := range r.ring {
		r.sortedKeys = append(r.sortedKeys, k)
	}

	for i := 0; i < len(r.sortedKeys)-1; i++ {
		for j := i + 1; j < len(r.sortedKeys); j++ {
			if r.sortedKeys[j] < r.sortedKeys[i] {
				r.sortedKeys[i], r.sortedKeys[j] = r.sortedKeys[j], r.sortedKeys[i]
			}
		}
	}
}

type OrderShardService struct {
	router         *ShardRouter
	consistentHash *ConsistentHashRouter
	loadBalancer   *ShardLoadBalancer
}

func NewOrderShardService(config *domain.ShardConfig) *OrderShardService {
	return &OrderShardService{
		router:         NewShardRouter(config),
		consistentHash: NewConsistentHashRouter(150),
		loadBalancer:   NewShardLoadBalancer(),
	}
}

func (s *OrderShardService) RouteOrder(order *domain.Order) (*ShardNode, error) {
	return s.router.Route(order)
}

func (s *OrderShardService) RouteByUserID(userID uint64) (*ShardNode, error) {
	return s.router.RouteByUserID(userID)
}

func (s *OrderShardService) RouteByOrderID(orderID uint64) (*ShardNode, error) {
	return s.router.RouteByOrderID(orderID)
}

func (s *OrderShardService) RouteWithLoadBalance(order *domain.Order) (*ShardNode, error) {
	preferredNode, err := s.router.Route(order)
	if err != nil {
		return nil, err
	}

	if preferredNode.LoadFactor > 0.8 {
		alternativeNode := s.loadBalancer.SelectLeastLoaded(s.router.GetAllActiveShards())
		if alternativeNode != nil {
			return alternativeNode, nil
		}
	}

	return preferredNode, nil
}

func (s *OrderShardService) AddShardNode(index int, address string, weight int) {
	s.router.AddShardNode(index, address, weight)

	node := &ShardNode{
		Index:    index,
		Address:  address,
		Weight:   weight,
		IsActive: true,
	}
	s.consistentHash.AddNode(node)
}

func (s *OrderShardService) RemoveShardNode(index int) {
	s.router.RemoveShardNode(index)
	s.consistentHash.RemoveNode(index)
}

func (s *OrderShardService) GetShardStats() map[int]*ShardStats {
	return s.router.GetShardStats()
}

type ShardLoadBalancer struct{}

func NewShardLoadBalancer() *ShardLoadBalancer {
	return &ShardLoadBalancer{}
}

func (b *ShardLoadBalancer) SelectLeastLoaded(nodes []*ShardNode) *ShardNode {
	if len(nodes) == 0 {
		return nil
	}

	var selected *ShardNode
	minLoad := float64(1.1)

	for _, node := range nodes {
		if node.IsActive && node.LoadFactor < minLoad {
			minLoad = node.LoadFactor
			selected = node
		}
	}

	return selected
}

func (b *ShardLoadBalancer) SelectRoundRobin(nodes []*ShardNode, currentIndex int) *ShardNode {
	if len(nodes) == 0 {
		return nil
	}

	for i := range nodes {
		idx := (currentIndex + i) % len(nodes)
		if nodes[idx].IsActive {
			return nodes[idx]
		}
	}

	return nil
}

func (b *ShardLoadBalancer) SelectWeighted(nodes []*ShardNode) *ShardNode {
	if len(nodes) == 0 {
		return nil
	}

	var totalWeight int
	for _, node := range nodes {
		if node.IsActive {
			totalWeight += node.Weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	target := int(time.Now().UnixNano() % int64(totalWeight))
	var currentWeight int

	for _, node := range nodes {
		if node.IsActive {
			currentWeight += node.Weight
			if currentWeight >= target {
				return node
			}
		}
	}

	return nil
}

type ShardMigrationService struct {
	router   *ShardRouter
	migrator *DataMigrator
	progress map[string]*MigrationProgress
	mu       sync.RWMutex
}

type MigrationProgress struct {
	FromShard    int        `json:"from_shard"`
	ToShard      int        `json:"to_shard"`
	TotalRecords int64      `json:"total_records"`
	Migrated     int64      `json:"migrated"`
	Status       string     `json:"status"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
}

type DataMigrator interface {
	Migrate(ctx context.Context, fromShard, toShard int, userIDRange [2]uint64) error
	Verify(ctx context.Context, fromShard, toShard int) (bool, error)
}

func NewShardMigrationService(router *ShardRouter) *ShardMigrationService {
	return &ShardMigrationService{
		router:   router,
		progress: make(map[string]*MigrationProgress),
	}
}

func (s *ShardMigrationService) StartMigration(ctx context.Context, migrationID string, fromShard, toShard int, totalRecords int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.progress[migrationID]; exists {
		return fmt.Errorf("migration %s already exists", migrationID)
	}

	s.progress[migrationID] = &MigrationProgress{
		FromShard:    fromShard,
		ToShard:      toShard,
		TotalRecords: totalRecords,
		Status:       "running",
		StartTime:    time.Now(),
	}

	return nil
}

func (s *ShardMigrationService) UpdateProgress(migrationID string, migrated int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, exists := s.progress[migrationID]; exists {
		p.Migrated = migrated
		if p.Migrated >= p.TotalRecords {
			p.Status = "completed"
			now := time.Now()
			p.EndTime = &now
		}
	}
}

func (s *ShardMigrationService) GetProgress(migrationID string) (*MigrationProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.progress[migrationID]
	return p, exists
}

func (s *ShardMigrationService) CancelMigration(migrationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, exists := s.progress[migrationID]; exists {
		p.Status = "cancelled"
		now := time.Now()
		p.EndTime = &now
	}
}
