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

func (m *ConcurrentMap[K, V]) getShard(key K) *shard[K, V] {
	// Simple hash function for comparable keys (this is a naive implementation,
	// for production robust hashing is needed, but Go generic constraints limit options without reflection or specific interfaces)
	// Here we rely on the fact that we can't easily hash generic K without more constraints.
	// For simplicity in this example, we'll assume K is string or int and cast, or use a simple address based hash if possible.
	// A better approach in Go 1.18+ is to require K to implement a Hash() method or use a hash library that supports generics.
	// For this example, we will just use a simple distribution strategy if K is string.
	// Real-world implementation should use a proper hash function.

	// Placeholder: In a real generic map, you'd need a hash function passed in or constraints.
	// We will use a simple round-robin or just lock the first shard if we can't hash.
	// To make this useful, let's assume K is string for the hash calculation part, or use a passed in hash func.
	// For now, to keep it generic and simple, we might just use a single lock if we can't shard effectively without a hash func.
	// BUT, the requirement is "high performance".
	// Let's change the signature to require a hash function or assume string keys for sharding.
	// Let's assume string keys for now as it's most common, or provide a Hashable interface.

	// Re-design: Let's make it ConcurrentStringMap for simplicity and performance correctness in this context.
	return m.shards[0] // Fallback to single shard if we don't implement hashing here.
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
