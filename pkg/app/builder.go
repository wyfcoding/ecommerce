package app

import (
	"flag"
	"fmt"
	"reflect"

	"ecommerce/pkg/config"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/server"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Builder provides a flexible way to construct an App.
type Builder struct {
	serviceName    string
	configInstance interface{}
	appOpts        []Option

	initService  interface{}
	registerGRPC interface{}
	registerGin  interface{}

	metricsPort string
}

// NewBuilder creates a new application builder.
func NewBuilder(serviceName string) *Builder {
	return &Builder{serviceName: serviceName}
}

// WithConfig sets the configuration instance.
func (b *Builder) WithConfig(conf interface{}) *Builder {
	b.configInstance = conf
	return b
}

// WithGRPC registers the gRPC server creation logic.
func (b *Builder) WithGRPC(register func(*grpc.Server, interface{})) *Builder {
	b.registerGRPC = register
	return b
}

// WithGin registers the Gin server creation logic.
func (b *Builder) WithGin(register func(*gin.Engine, interface{})) *Builder {
	b.registerGin = register
	return b
}

// WithService registers the service initialization logic.
// The init function now receives a *metrics.Metrics instance.
func (b *Builder) WithService(init func(interface{}, *metrics.Metrics) (interface{}, func(), error)) *Builder {
	b.initService = init
	return b
}

// WithMetrics enables the Prometheus metrics server on the given port.
func (b *Builder) WithMetrics(port string) *Builder {
	b.metricsPort = port
	return b
}

// Build constructs the final App.
func (b *Builder) Build() *App {
	// 1. 加载配置
	configPath := fmt.Sprintf("./configs/%s.toml", b.serviceName)
	var flagConfigPath string
	flag.StringVar(&flagConfigPath, "conf", configPath, "config file path")
	flag.Parse()

	if err := config.Load(flagConfigPath, b.configInstance); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	val := reflect.ValueOf(b.configInstance).Elem()
	cfgField := val.FieldByName("Config")
	if !cfgField.IsValid() {
		panic("config instance must embed config.Config")
	}
	cfg := cfgField.Interface().(config.Config)

	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. (Optional) 初始化 Metrics
	var metricsInstance *metrics.Metrics
	if b.metricsPort != "" {
		metricsInstance = metrics.NewMetrics(b.serviceName)
		metricsCleanup := metricsInstance.ExposeHttp(b.metricsPort)
		b.appOpts = append(b.appOpts, WithCleanup(metricsCleanup))
	}

	// 4. 依赖注入
	serviceInstance, cleanup, err := b.initService.(func(interface{}, *metrics.Metrics) (interface{}, func(), error))(b.configInstance, metricsInstance)
	if err != nil {
		zap.S().Fatalf("failed to init service: %v", err)
	}
	b.appOpts = append(b.appOpts, WithCleanup(cleanup))

	// 5. 创建服务器
	var servers []server.Server
	if b.registerGRPC != nil {
		grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
		grpcSrv := server.NewGRPCServer(grpcAddr, func(s *grpc.Server) {
			b.registerGRPC.(func(*grpc.Server, interface{}))(s, serviceInstance)
		})
		servers = append(servers, grpcSrv)
	}
	if b.registerGin != nil {
		httpAddr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
		ginEngine := server.NewDefaultGinEngine()
		b.registerGin.(func(*gin.Engine, interface{}))(ginEngine, serviceInstance)
		ginSrv := server.NewGinServer(ginEngine, httpAddr)
		servers = append(servers, ginSrv)
	}
	b.appOpts = append(b.appOpts, WithServer(servers...))

	// 6. 创建应用
	return New(b.appOpts...)
}
