package neo4j

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

// Config 结构体用于 Neo4j 数据库配置。
type Config struct {
	URI      string `toml:"uri"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// NewNeo4jDriver 创建一个新的 Neo4j 驱动实例。
func NewNeo4jDriver(conf *Config) (neo4j.Driver, func(), error) {
	driver, err := neo4j.NewDriver(conf.URI, neo4j.BasicAuth(conf.Username, conf.Password, ""))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	// 验证连接
	if err := driver.VerifyConnectivity(); err != nil {
		return nil, nil, fmt.Errorf("failed to verify neo4j connectivity: %w", err)
	}

	cleanup := func() {
		if err := driver.Close(); err != nil {
		zap.S().Errorf("failed to close neo4j driver: %v", err)
		}
	}

	return driver, cleanup, nil
}
