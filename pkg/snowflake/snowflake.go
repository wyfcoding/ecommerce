package snowflake

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
)

var ErrNodeNotInitialized = errors.New("snowflake node not initialized")

// Config 结构体用于 Snowflake ID 生成器配置。
type Config struct {
	StartTime string `toml:"start_time"`
	MachineID int64  `toml:"machine_id"`
}

// SnowflakeNode 封装了 snowflake.Node 以提供方法。
type SnowflakeNode struct {
	n *snowflake.Node
}

// NewSnowflakeNode 初始化一个新的雪花节点。
//
// WARNING: The underlying bwmarrin/snowflake library uses a global `snowflake.Epoch` variable.
// This means that `conf.StartTime` should be consistent across all initializations within the same process.
// If multiple `NewSnowflakeNode` calls are made with different `conf.StartTime` values,
// the `snowflake.Epoch` will be overwritten, potentially leading to incorrect ID generation.
// Ensure `conf.StartTime` is set once and consistently for the entire application.
//
// 在真实的分布式系统中，nodeID 应该从服务发现（如 etcd）获取，
// 或者通过环境变量/配置文件传递，以避免冲突。
func NewSnowflakeNode(conf *Config) (*SnowflakeNode, func(), error) {
	var st time.Time
	st, err := time.Parse("2006-01-02", conf.StartTime)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	snowflake.Epoch = st.UnixNano() / 1000000
	n, err := snowflake.NewNode(conf.MachineID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	cleanup := func() {}

	return &SnowflakeNode{n: n}, cleanup, nil
}

// GenID 生成一个雪花 ID。
func (sn *SnowflakeNode) GenID() (uint64, error) {
	if sn == nil || sn.n == nil {
		return 0, ErrNodeNotInitialized
	}
	return uint64(sn.n.Generate().Int64()), nil
}
