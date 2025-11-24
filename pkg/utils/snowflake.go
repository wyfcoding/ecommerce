package utils

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// InitSnowflake initializes the snowflake node.
func InitSnowflake(nodeID int64) error {
	var err error
	once.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// GenerateID generates a new snowflake ID.
func GenerateID() int64 {
	if node == nil {
		// Fallback or panic? For safety, let's try to init with default if not initialized
		// But ideally InitSnowflake should be called at startup.
		// If not initialized, we can't guarantee uniqueness across nodes.
		// Let's panic to enforce initialization.
		panic("snowflake node not initialized")
	}
	return node.Generate().Int64()
}
