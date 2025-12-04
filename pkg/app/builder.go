package app

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/logging"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/wyfcoding/ecommerce/pkg/server"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// Builder 提供了构建 App 的灵活方式。
type Builder struct {
	serviceName    string      // 服务名称
	configInstance interface{} // 配置实例
	appOpts        []Option    // 应用程序选项

	initService  interface{} // 服务初始化函数
	registerGRPC interface{} // gRPC 注册函数
	registerGin  interface{} // Gin 注册函数

	metricsPort string // Metrics 端口

	// grpcInterceptors 用于收集gRPC一元拦截器。
	grpcInterceptors []grpc.UnaryServerInterceptor
	// ginMiddleware 用于收集Gin中间件。
	ginMiddleware []gin.HandlerFunc
}

// NewBuilder 创建一个新的应用构建器。
func NewBuilder(serviceName string) *Builder {
	return &Builder{serviceName: serviceName}
}

// WithConfig 设置配置实例。
// conf 应该是一个结构体的指针，用于加载配置。
func (b *Builder) WithConfig(conf interface{}) *Builder {
	b.configInstance = conf
	return b
}

// WithGRPC 注册 gRPC 服务器的创建逻辑。
// register 函数负责将具体的gRPC服务注册到 `*grpc.Server` 实例中。
func (b *Builder) WithGRPC(register func(*grpc.Server, interface{})) *Builder {
	b.registerGRPC = register
	return b
}

// WithGin 注册 Gin 服务器的创建逻辑。
// register 函数负责将HTTP路由注册到 `*gin.Engine` 实例中。
func (b *Builder) WithGin(register func(*gin.Engine, interface{})) *Builder {
	b.registerGin = register
	return b
}

// WithService 注册服务的初始化逻辑。
// init 函数接收配置和Metrics实例，并返回服务实例、清理函数和错误。
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
// 这些拦截器将在gRPC服务器创建时被链式应用。
func (b *Builder) WithGRPCInterceptor(interceptors ...grpc.UnaryServerInterceptor) *Builder {
	b.grpcInterceptors = append(b.grpcInterceptors, interceptors...)
	return b
}

// WithGinMiddleware 添加一个或多个 Gin 中间件。
// 这些中间件将在Gin引擎创建时被应用。
func (b *Builder) WithGinMiddleware(middleware ...gin.HandlerFunc) *Builder {
	b.ginMiddleware = append(b.ginMiddleware, middleware...)
	return b
}

// Build 构建最终的 App 实例。
// 它负责加载配置、初始化日志、Metrics、服务实例，并创建和注册gRPC和Gin服务器。
func (b *Builder) Build() *App {
	// 1. 加载配置。
	configPath := fmt.Sprintf("./configs/%s/config.toml", b.serviceName)
	var flagConfigPath string
	flag.StringVar(&flagConfigPath, "conf", configPath, "配置文件路径")
	flag.Parse()

	if err := config.Load(flagConfigPath, b.configInstance); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 2. 初始化日志。
	logger := logging.NewLogger(b.serviceName, "app")

	// 3. (可选) 初始化 Metrics。
	var metricsInstance *metrics.Metrics
	if b.metricsPort != "" {
		metricsInstance = metrics.NewMetrics(b.serviceName)
		metricsCleanup := metricsInstance.ExposeHttp(b.metricsPort)
		b.appOpts = append(b.appOpts, WithCleanup(metricsCleanup))
	}

	// 4. 依赖注入：初始化核心服务。
	// 使用类型断言调用注册的服务初始化函数。
	serviceInstance, cleanup, err := b.initService.(func(interface{}, *metrics.Metrics) (interface{}, func(), error))(b.configInstance, metricsInstance)
	if err != nil {
		logger.Logger.Error("初始化服务失败", "error", err)
		panic(err)
	}
	b.appOpts = append(b.appOpts, WithCleanup(cleanup))

	// 5. 创建服务器。
	// 使用反射从配置实例中获取 `config.Config` 结构。
	val := reflect.ValueOf(b.configInstance).Elem()
	cfgField := val.FieldByName("Config")
	if !cfgField.IsValid() {
		panic("配置实例必须包含一个名为 'Config' 的字段，类型为 config.Config")
	}
	cfg := cfgField.Interface().(config.Config)

	var servers []server.Server
	// 如果注册了gRPC服务，则创建gRPC服务器。
	if b.registerGRPC != nil {
		grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
		grpcSrv := server.NewGRPCServer(grpcAddr, logger.Logger, func(s *grpc.Server) {
			b.registerGRPC.(func(*grpc.Server, interface{}))(s, serviceInstance)
		}, b.grpcInterceptors...)
		servers = append(servers, grpcSrv)
	}
	// 如果注册了Gin服务，则创建Gin HTTP服务器。
	if b.registerGin != nil {
		httpAddr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
		ginEngine := server.NewDefaultGinEngine(logger.Logger, b.ginMiddleware...)
		b.registerGin.(func(*gin.Engine, interface{}))(ginEngine, serviceInstance)
		ginSrv := server.NewGinServer(ginEngine, httpAddr, logger.Logger)
		servers = append(servers, ginSrv)
	}
	b.appOpts = append(b.appOpts, WithServer(servers...))

	// 6. 创建应用实例。
	return New(b.serviceName, logger.Logger, b.appOpts...)
}
