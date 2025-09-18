package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init a new snowflake node.
// In a real distributed system, the nodeID should be fetched from a discovery service like etcd
// or passed via environment variables/configuration files to avoid conflicts.
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}

	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(machineID)
	return
}

// GenID generates a snowflake ID.
func GenID() int64 {
	return node.Generate().Int64()
}
