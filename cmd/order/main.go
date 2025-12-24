package order

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/infrastructure/persistence"
	ordergrpc "github.com/wyfcoding/ecommerce/internal/order/interfaces/grpc"
	httpServer "github.com/wyfcoding/ecommerce/internal/order/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	"github.com/wyfcoding/pkg/cache"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/databases/sharding"
	"github.com/wyfcoding/pkg/grpcclient"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"
)

// BootstrapName 服务名称常量。
const BootstrapName = "order"

// Config 扩展了基础配置，增加了服务特定的设置
type Config struct {
	configpkg.Config `mapstructure:",squash"`
	DTM              struct {
		Server string `mapstructure:"server"`
	} `mapstructure:"dtm"`
	Warehouse struct {
		GrpcAddr string `mapstructure:"grpc_addr"`
	} `mapstructure:"warehouse"`
}

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	Config     *Config
	AppService *application.OrderService
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	User      *grpc.ClientConn
	Product   *grpc.ClientConn
	Inventory *grpc.ClientConn
	Pricing   *grpc.ClientConn
	Payment   *grpc.ClientConn
	Marketing *grpc.ClientConn
	Logistics *grpc.ClientConn
	Cart      *grpc.ClientConn
}

func main() {
	if err := app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run(); err != nil {
		slog.Error("application run failed", "error", err)
	}
}

func registerGRPC(s *grpc.Server, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterOrderServer(s, ordergrpc.NewServer(service))
	slog.Default().Info("gRPC server registered", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	handler := httpServer.NewHandler(ctx.AppService, slog.Default())
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	// 1. 数据库分片管理器
	slog.Info("initializing database sharding manager...")
	shardingManager, err := sharding.NewManager(c.Data.Shards, logging.Default())
	if err != nil {
		if len(c.Data.Shards) == 0 {
			slog.Info("No shards configured, using single database connection")
			db, err := databases.NewDB(c.Data.Database, logging.Default())
			if err != nil {
				return nil, nil, fmt.Errorf("failed to initialize single db manager: %w", err)
			}
			shardingManager, err = sharding.NewManager([]configpkg.DatabaseConfig{c.Data.Database}, logging.Default())
			if err != nil {
				sqlDB, _ := db.DB()
				sqlDB.Close()
				return nil, nil, fmt.Errorf("failed to initialize single db manager: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to initialize sharding manager: %w", err)
		}
	}

	// 2. Redis 缓存
	redisCache, err := cache.NewRedisCache(c.Data.Redis)
	if err != nil {
		shardingManager.Close()
		return nil, nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	// 3. ID 生成器
	idGen, err := idgen.NewSnowflakeGenerator(c.Snowflake)
	if err != nil {
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("failed to init id generator: %w", err)
	}

	// 4. Kafka 生产者
	kafkaProducer := kafka.NewProducer(c.MessageQueue.Kafka, logging.Default())

	// 下游客户端
	clients := &ServiceClients{}
	clientCleanup, err := grpcclient.InitServiceClients(c.Services, clients)
	if err != nil {
		kafkaProducer.Close()
		redisCache.Close()
		shardingManager.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 5. 基础设施与应用层
	repo := persistence.NewOrderRepository(shardingManager)

	orderManager := application.NewOrderManager(repo, idGen, kafkaProducer, slog.Default(), c.DTM.Server, c.Warehouse.GrpcAddr, m)
	orderQuery := application.NewOrderQuery(repo)

	service := application.NewOrderService(orderManager, orderQuery, slog.Default())

	cleanup := func() {
		slog.Info("cleaning up resources...", "service", BootstrapName)
		clientCleanup()
		kafkaProducer.Close()
		redisCache.Close()
		shardingManager.Close()
	}

	return &AppContext{AppService: service, Config: c, Clients: clients}, cleanup, nil
}
