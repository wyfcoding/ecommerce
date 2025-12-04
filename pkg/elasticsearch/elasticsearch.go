package elasticsearch

import (
	"fmt"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v9" // 导入Elasticsearch Go客户端库。
)

// Config 结构体用于 Elasticsearch 客户端配置。
// 它包含了连接和操作Elasticsearch集群所需的所有参数。
type Config struct {
	Addresses     []string `toml:"addresses"`       // Elasticsearch节点的地址列表，例如 ["http://localhost:9200"]。
	Username      string   `toml:"username"`        // 连接Elasticsearch的用户名。
	Password      string   `toml:"password"`        // 连接Elasticsearch的密码。
	CloudID       string   `toml:"cloud_id"`        // Elasticsearch Service (ESS) 的Cloud ID，用于云部署。
	APIKey        string   `toml:"api_key"`         // 用于认证的API Key。
	ServiceToken  string   `toml:"service_token"`   // 用于认证的服务Token。
	CACert        string   `toml:"ca_cert"`         // CA证书内容，用于TLS连接。
	RetryOnStatus []int    `toml:"retry_on_status"` // 需要重试的HTTP状态码列表。
	MaxRetries    int      `toml:"max_retries"`     // 最大重试次数。
}

// NewElasticsearchClient 创建一个新的 Elasticsearch 客户端实例。
// 它根据提供的配置建立连接，并返回 `*elasticsearch.Client` 实例和一个用于清理的空函数。
// conf: 包含Elasticsearch连接参数的配置结构体。
func NewElasticsearchClient(conf *Config) (*elasticsearch.Client, func(), error) {
	// 构建Elasticsearch客户端配置。
	cfg := elasticsearch.Config{
		Addresses:     conf.Addresses,
		Username:      conf.Username,
		Password:      conf.Password,
		CloudID:       conf.CloudID,
		APIKey:        conf.APIKey,
		ServiceToken:  conf.ServiceToken,
		CACert:        []byte(conf.CACert), // CA证书需要转换为字节切片。
		RetryOnStatus: conf.RetryOnStatus,
		MaxRetries:    conf.MaxRetries,
		// 还可以添加 Transport, Logger, Header 等更多高级配置。
	}

	// 使用配置创建新的Elasticsearch客户端。
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 尝试通过调用Info()方法检查与Elasticsearch集群的连接状态。
	res, err := es.Info()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get elasticsearch info: %w", err)
	}
	defer res.Body.Close() // 确保关闭响应体。

	if res.IsError() {
		// 如果Elasticsearch返回错误响应，则报告错误。
		return nil, nil, fmt.Errorf("elasticsearch client info error: %s", res.String())
	}

	slog.Info("Successfully connected to Elasticsearch", "info", res.String())

	// 返回一个空的清理函数。
	// 对于Elasticsearch客户端，通常不需要显式关闭连接，因为它是基于HTTP的，
	// 连接池由客户端库管理。提供一个空的清理函数是为了保持接口一致性。
	cleanup := func() {}

	return es, cleanup, nil
}
