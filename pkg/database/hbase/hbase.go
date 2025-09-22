package hbase

import (
	"context"
	"fmt"
	"time"

	"github.com/tsuna/go-hbase"
	"github.com/tsuna/go-hbase/hrpc"
	"go.uber.org/zap"
)

// Config 结构体用于 HBase 客户端配置。
type Config struct {
	ZookeeperQuorum []string      `toml:"zookeeper_quorum"`
	RootZnode       string        `toml:"root_znode"`
	Timeout         time.Duration `toml:"timeout"`
}

// NewHBaseClient 创建一个新的 HBase 客户端实例。
func NewHBaseClient(conf *Config) (hbase.Client, func(), error) {
	client := hbase.NewClient(conf.ZookeeperQuorum,
		hbase.ZnodeParent(conf.RootZnode),
		hbase.RpcTimeout(conf.Timeout),
		hbase.ReadTimeout(conf.Timeout),
		hbase.WriteTimeout(conf.Timeout),
	)

	// Verify connection by trying to get a table list (simplified)
	// In a real scenario, you might want a more robust connection check.
	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cancel()
	_, err := client.Tables(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to HBase: %w", err)
	}

	zap.S().Infof("Successfully connected to HBase Zookeeper quorum: %v", conf.ZookeeperQuorum)

	cleanup := func() {
		if client != nil {
			zap.S().Info("closing HBase client...")
			client.Close() // go-hbase client has a Close method
		}
	}

	return client, cleanup, nil
}
