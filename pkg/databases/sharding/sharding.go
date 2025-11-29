package sharding

import (
	"fmt"
	"sync"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/databases"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"gorm.io/gorm"
)

// Manager manages multiple database shards.
type Manager struct {
	shards     map[int]*gorm.DB
	shardCount int
	mu         sync.RWMutex
}

// NewManager creates a new sharding manager.
func NewManager(configs []config.DatabaseConfig, logger *logging.Logger) (*Manager, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no database configs provided")
	}

	shards := make(map[int]*gorm.DB)
	for i, cfg := range configs {
		db, err := databases.NewDB(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shard %d: %w", i, err)
		}
		shards[i] = db
	}

	return &Manager{
		shards:     shards,
		shardCount: len(configs),
	}, nil
}

// GetDB returns the database instance for the given key (e.g., UserID).
func (m *Manager) GetDB(key uint64) *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shardIndex := int(key % uint64(m.shardCount))
	return m.shards[shardIndex]
}

// Close closes all database connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for i, db := range m.shards {
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get sql db for shard %d: %w", i, err))
			continue
		}
		if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close shard %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close some shards: %v", errs)
	}
	return nil
}
