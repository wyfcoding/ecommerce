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

// Builder 提供了构建 App 的灵活方式。
type Builder struct {
	serviceName    string
	configInstance interface{}
	appOpts        []Option

	initService  interface{}
	registerGRPC interface{}
	registerGin  interface{}

	metricsPort string

	// 新增字段以收集拦截器和中间件
	grpcInterceptors []grpc.UnaryServerInterceptor
	ginMiddleware    []gin.HandlerFunc
}

// NewBuilder 创建一个新的应用构建器。
func NewBuilder(serviceName string) *Builder {
	return &Builder{serviceName: serviceName}
}

// WithConfig 设置配置实例。
func (b *Builder) WithConfig(conf interface{}) *Builder {
	b.configInstance = conf
	return b
}

// WithGRPC 注册 gRPC 服务器的创建逻辑。
func (b *Builder) WithGRPC(register func(*grpc.Server, interface{})) *Builder {
	b.registerGRPC = register
	return b
}

// WithGin 注册 Gin 服务器的创建逻辑。
func (b *Builder) WithGin(register func(*gin.Engine, interface{})) *Builder {
	b.registerGin = register
	return b
}

// WithService 注册服务的初始化逻辑。
func (b *Builder) WithService(init func(interface{}, *metrics.Metrics) (interface{}, func(), error)) *Builder {
	b.initService = init
	return b
}

// WithMetrics 在指定端口上启用 Prometheus 指标服务器。
func (b *Builder) WithMetrics(port string) *Builder {
	b.metricsPort = port
	return b
}

// WithGRPCInterceptor 添加一个或多个 gRPC 一元拦截器。
func (b *Builder) WithGRPCInterceptor(interceptors ...grpc.UnaryServerInterceptor) *Builder {
	b.grpcInterceptors = append(b.grpcInterceptors, interceptors...)
	return b
}

// WithGinMiddleware 添加一个或多个 Gin 中间件。
func (b *Builder) WithGinMiddleware(middleware ...gin.HandlerFunc) *Builder {
	b.ginMiddleware = append(b.ginMiddleware, middleware...)
	return b
}

// Build 构建最终的 App。
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

	// 3. (可选) 初始化 Metrics
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
		// 将收集到的拦截器传递给 gRPC 服务器构造函数
		grpcSrv := server.NewGRPCServer(grpcAddr, func(s *grpc.Server) {
			b.registerGRPC.(func(*grpc.Server, interface{}))(s, serviceInstance)
		}, b.grpcInterceptors...)
		servers = append(servers, grpcSrv)
	}
	if b.registerGin != nil {
		httpAddr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
		// 将收集到的中间件传递给 Gin 引擎构造函数
		ginEngine := server.NewDefaultGinEngine(b.ginMiddleware...)
		b.registerGin.(func(*gin.Engine, interface{}))(ginEngine, serviceInstance)
		ginSrv := server.NewGinServer(ginEngine, httpAddr)
		servers = append(servers, ginSrv)
	}
	b.appOpts = append(b.appOpts, WithServer(servers...))

	// 6. 创建应用
	return New(b.appOpts...)
}