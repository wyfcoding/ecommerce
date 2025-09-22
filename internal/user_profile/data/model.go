package data

import (
	"time"

	"github.com/tsuna/go-hbase/hrpc" // For HBase specific types if needed
)

// UserProfile represents a user's profile in HBase.
// In HBase, data is stored in rows, with column families and columns.
// This struct represents a flattened view of a user's profile.
type UserProfile struct {
	UserID          string            `hbase:"key"` // Row key
	Gender          string            `hbase:"cf:gender"`
	AgeGroup        string            `hbase:"cf:age_group"`
	City            string            `hbase:"cf:city"`
	Interests       []string          `hbase:"cf:interests,json"` // Stored as JSON array
	RecentCategories []string         `hbase:"cf:recent_categories,json"`
	RecentBrands    []string          `hbase:"cf:recent_brands,json"`
	TotalSpent      uint64            `hbase:"cf:total_spent"`
	OrderCount      uint32            `hbase:"cf:order_count"`
	LastActiveTime  time.Time         `hbase:"cf:last_active_time"`
	CustomTags      map[string]string `hbase:"cf:custom_tags,json"`
	UpdatedAt       time.Time         `hbase:"cf:updated_at"` // Last update timestamp
}

// UserBehaviorEvent represents a user behavior event to be stored in HBase or processed.
type UserBehaviorEvent struct {
	UserID       string    `hbase:"key"` // Row key (composite key with timestamp)
	BehaviorType string    `hbase:"cf:behavior_type"`
	ItemID       string    `hbase:"cf:item_id"`
	Properties   map[string]string `hbase:"cf:properties,json"`
	EventTime    time.Time `hbase:"cf:event_time"`
}
