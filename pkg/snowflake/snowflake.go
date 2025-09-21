package snowflake

import (
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init 初始化一个新的雪花节点。
// 在真实的分布式系统中，nodeID 应该从服务发现（如 etcd）获取，
// 或者通过环境变量/配置文件传递，以避免冲突。
func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}

	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		return fmt.Errorf("failed to create snowflake node: %w", err)
	}
	return nil
}

// GenID 生成一个雪花 ID。
func GenID() uint64 {
	if node == nil {
		panic("snowflake node not initialized, call Init() first")
	}
	return uint64(node.Generate().Int64())
}
