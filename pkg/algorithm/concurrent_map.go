package algorithm

import (
	"sync"
)

// ConcurrentMap implements a thread-safe map using sharding to reduce lock contention.
type ConcurrentMap[K comparable, V any] struct {
	shards     []*shard[K, V]
	shardCount uint64
}

type shard[K comparable, V any] struct {
	sync.RWMutex
	items map[K]V
}

// NewConcurrentMap creates a new ConcurrentMap.
func NewConcurrentMap[K comparable, V any](shardCount int) *ConcurrentMap[K, V] {
	if shardCount <= 0 {
		shardCount = 32
	}
	m := &ConcurrentMap[K, V]{
		shards:     make([]*shard[K, V], shardCount),
		shardCount: uint64(shardCount),
	}
	for i := 0; i < shardCount; i++ {
		m.shards[i] = &shard[K, V]{items: make(map[K]V)}
	}
	return m
}

// ConcurrentStringMap is a specialized version for string keys to allow sharding.
type ConcurrentStringMap[V any] struct {
	shards     []*shard[string, V]
	shardCount uint64
}

func NewConcurrentStringMap[V any](shardCount int) *ConcurrentStringMap[V] {
	if shardCount <= 0 {
		shardCount = 32
	}
	m := &ConcurrentStringMap[V]{
		shards:     make([]*shard[string, V], shardCount),
		shardCount: uint64(shardCount),
	}
	for i := 0; i < shardCount; i++ {
		m.shards[i] = &shard[string, V]{items: make(map[string]V)}
	}
	return m
}

func (m *ConcurrentStringMap[V]) getShard(key string) *shard[string, V] {
	hash := fnv32(key)
	return m.shards[uint64(hash)%m.shardCount]
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = 16777619
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (m *ConcurrentStringMap[V]) Set(key string, value V) {
	shard := m.getShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

func (m *ConcurrentStringMap[V]) Get(key string) (V, bool) {
	shard := m.getShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m *ConcurrentStringMap[V]) Remove(key string) {
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}
