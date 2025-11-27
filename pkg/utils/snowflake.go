package utils

import (
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// initNode initializes node automatically if not initialized.
func initNode() error {
	var err error
	once.Do(func() {
		nodeID := int64(1) // default nodeID = 1

		// try to read from env: SNOWFLAKE_NODE_ID
		if v := os.Getenv("SNOWFLAKE_NODE_ID"); v != "" {
			if n, e := strconv.ParseInt(v, 10, 64); e == nil {
				nodeID = n
			}
		}

		node, err = snowflake.NewNode(nodeID)
	})

	return err
}

// GenerateID generates a unique ID safely.
func GenerateID() int64 {
	if node == nil {
		if err := initNode(); err != nil {
			panic("failed to init snowflake node: " + err.Error())
		}
	}
	return node.Generate().Int64()
}
