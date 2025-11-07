package proxy

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name    string
	Address string
	Timeout time.Duration
}

// GRPCProxy gRPC 代理
type GRPCProxy struct {
	connections map[string]*grpc.ClientConn
	configs     map[string]*ServiceConfig
	mu          sync.RWMutex
}

// NewGRPCProxy 创建 gRPC 代理
func NewGRPCProxy() *GRPCProxy {
	return &GRPCProxy{
		connections: make(map[string]*grpc.ClientConn),
		configs:     make(map[string]*ServiceConfig),
	}
}

// RegisterService 注册服务
func (p *GRPCProxy) RegisterService(config *ServiceConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.configs[config.Name]; exists {
		return fmt.Errorf("service %s already registered", config.Name)
	}

	p.configs[config.Name] = config
	zap.S().Infof("registered service: %s at %s", config.Name, config.Address)
	return nil
}

// GetConnection 获取服务连接
func (p *GRPCProxy) GetConnection(serviceName string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, exists := p.connections[serviceName]
	config, configExists := p.configs[serviceName]
	p.mu.RUnlock()

	if !configExists {
		return nil, fmt.Errorf("service %s not registered", serviceName)
	}

	// 如果连接存在且有效，直接返回
	if exists && conn.GetState().String() != "SHUTDOWN" {
		return conn, nil
	}

	// 创建新连接
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查
	conn, exists = p.connections[serviceName]
	if exists && conn.GetState().String() != "SHUTDOWN" {
		return conn, nil
	}

	// 建立新连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newConn, err := grpc.DialContext(
		ctx,
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serviceName, err)
	}

	p.connections[serviceName] = newConn
	zap.S().Infof("established connection to service: %s", serviceName)
	return newConn, nil
}

// Close 关闭所有连接
func (p *GRPCProxy) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, conn := range p.connections {
		if err := conn.Close(); err != nil {
			zap.S().Errorf("failed to close connection to %s: %v", name, err)
		}
	}
	p.connections = make(map[string]*grpc.ClientConn)
}

// ProxyRequest 代理请求到 gRPC 服务
func (p *GRPCProxy) ProxyRequest(c *gin.Context, serviceName, method string, req, resp interface{}) error {
	// 获取连接
	conn, err := p.GetConnection(serviceName)
	if err != nil {
		zap.S().Errorf("failed to get connection for %s: %v", serviceName, err)
		return err
	}

	// 获取配置
	p.mu.RLock()
	config := p.configs[serviceName]
	p.mu.RUnlock()

	// 创建上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
	defer cancel()

	// 传递 metadata
	md := metadata.New(map[string]string{})
	if userID := c.GetHeader("x-user-id"); userID != "" {
		md.Set("x-user-id", userID)
	}
	if username := c.GetHeader("x-username"); username != "" {
		md.Set("x-username", username)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 调用 gRPC 方法
	err = conn.Invoke(ctx, method, req, resp)
	if err != nil {
		zap.S().Errorf("failed to invoke %s.%s: %v", serviceName, method, err)
		return err
	}

	return nil
}

// HTTPToGRPCMiddleware HTTP 到 gRPC 的转换中间件
func (p *GRPCProxy) HTTPToGRPCMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查服务是否可用
		_, err := p.GetConnection(serviceName)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": fmt.Sprintf("service %s unavailable", serviceName),
			})
			c.Abort()
			return
		}

		// 将服务名存入上下文，供后续处理器使用
		c.Set("service_name", serviceName)
		c.Next()
	}
}

// HealthCheck 健康检查
func (p *GRPCProxy) HealthCheck() map[string]bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	health := make(map[string]bool)
	for name, conn := range p.connections {
		state := conn.GetState()
		health[name] = state.String() == "READY"
	}
	return health
}
