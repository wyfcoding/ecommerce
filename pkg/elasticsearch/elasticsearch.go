package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

// Config 结构体用于 Elasticsearch 客户端配置。
type Config struct {
	Addresses     []string `toml:"addresses"`
	Username      string   `toml:"username"`
	Password      string   `toml:"password"`
	CloudID       string   `toml:"cloud_id"`
	APIKey        string   `toml:"api_key"`
	ServiceToken  string   `toml:"service_token"`
	CACert        string   `toml:"ca_cert"`
	RetryOnStatus []int    `toml:"retry_on_status"`
	MaxRetries    int      `toml:"max_retries"`
}

// NewElasticsearchClient 创建一个新的 Elasticsearch 客户端实例。
func NewElasticsearchClient(conf *Config) (*elasticsearch.Client, func(), error) {
	cfg := elasticsearch.Config{
		Addresses:     conf.Addresses,
		Username:      conf.Username,
		Password:      conf.Password,
		CloudID:       conf.CloudID,
		APIKey:        conf.APIKey,
		ServiceToken:  conf.ServiceToken,
		CACert:        []byte(conf.CACert),
		RetryOnStatus: conf.RetryOnStatus,
		MaxRetries:    conf.MaxRetries,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 检查连接
	res, err := es.Info()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, nil, fmt.Errorf("elasticsearch client info error: %s", res.String())
	}

	zap.S().Info("Successfully connected to Elasticsearch: ", res.String())

	// 返回一个空的清理函数以保持接口一致性
	cleanup := func() {}

	return es, cleanup, nil
}
