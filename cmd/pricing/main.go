package main

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/wyfcoding/pkg/grpcclient"

	pb "github.com/wyfcoding/ecommerce/go-api/pricing/v1"
	"github.com/wyfcoding/ecommerce/internal/pricing/application"
	"github.com/wyfcoding/ecommerce/internal/pricing/infrastructure/persistence"
	pricinggrpc "github.com/wyfcoding/ecommerce/internal/pricing/interfaces/grpc"
	pricinghttp "github.com/wyfcoding/ecommerce/internal/pricing/interfaces/http"
	"github.com/wyfcoding/pkg/app"
	configpkg "github.com/wyfcoding/pkg/config"
	"github.com/wyfcoding/pkg/databases"
	"github.com/wyfcoding/pkg/logging"
	"github.com/wyfcoding/pkg/metrics"
	"github.com/wyfcoding/pkg/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// AppContext 应用上下文，包含配置、服务实例和客户端依赖。
type AppContext struct {
	AppService *application.PricingService
	Config     *configpkg.Config
	Clients    *ServiceClients
}

// ServiceClients 包含所有下游服务的 gRPC 客户端连接。
type ServiceClients struct {
	// 如果需要，在此处添加依赖项
}

// BootstrapName 服务名称常量。
const BootstrapName = "pricing-service"

func main() {
	app.NewBuilder(BootstrapName).
		WithConfig(&configpkg.Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithGinMiddleware(middleware.CORS()).
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	pb.RegisterPricingServiceServer(s, pricinggrpc.NewServer(service))
	slog.Default().Info("gRPC server registered (DDD)", "service", BootstrapName)
}

func registerGin(e *gin.Engine, srv any) {
	ctx := srv.(*AppContext)
	service := ctx.AppService
	handler := pricinghttp.NewHandler(service, slog.Default())

	api := e.Group("/api/v1")
	handler.RegisterRoutes(api)

	slog.Default().Info("HTTP routes registered (DDD)", "service", BootstrapName)
}

func initService(cfg any, m *metrics.Metrics) (any, func(), error) {
	c := cfg.(*configpkg.Config)
	slog.Info("initializing service dependencies...", "service", BootstrapName)

	// 初始化日志
	logger := logging.NewLogger(BootstrapName, "app")

	// 初始化数据库
	db, err := databases.NewDB(c.Data.Database, logging.Default())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 基础设施层
	repo := persistence.NewPricingRepository(db)

	// 下游客户端
	clients, clientCleanups, err := initClients(c)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to init clients: %w", err)
	}

	// 应用层
	mgr := application.NewPricingManager(repo, logger.Logger)
	query := application.NewPricingQuery(repo)
	service := application.NewPricingService(mgr, query)

	cleanup := func() {
		slog.Info("cleaning up resources...", "service", BootstrapName)
		clientCleanups()
		sqlDB.Close()
	}

	return &AppContext{AppService: service, Config: c, Clients: clients}, cleanup, nil
}

// initClients 使用反射自动初始化下游 gRPC 客户端
// 它会遍历 ServiceClients 的字段，并在配置中查找同名字段
func initClients(cfg *configpkg.Config) (*ServiceClients, func(), error) {
	clients := &ServiceClients{}
	var conns []*grpc.ClientConn

	// 反射获取目标结构体（ServiceClients）的值和类型
	clientsVal := reflect.ValueOf(clients).Elem()
	// 反射获取配置源结构体（ServicesConfig）的值
	servicesCfgVal := reflect.ValueOf(cfg.Services)

	for i := 0; i < clientsVal.NumField(); i++ {
		field := clientsVal.Type().Field(i)
		fieldName := field.Name

		// 在配置中查找同名字段
		cfgField := servicesCfgVal.FieldByName(fieldName)
		if !cfgField.IsValid() {
			continue // 配置中没有这个服务，跳过
		}

		// 获取 ServiceAddr.GRPCAddr 字段的值
		// 假设 cfgField 是 configpkg.ServiceAddr 类型
		addrField := cfgField.FieldByName("GRPCAddr")
		if !addrField.IsValid() || addrField.String() == "" {
			continue // 地址为空，跳过
		}

		addr := addrField.String()
		slog.Debug("initializing grpc client", "service", fieldName, "addr", addr)

		conn, err := grpcclient.NewClient(grpcclient.ClientConfig{
			Target:      addr,
			Timeout:     5000,
			ConnTimeout: 5000,
		})
		if err != nil {
			// 关闭已打开的连接
			for _, c := range conns {
				c.Close()
			}
			return nil, nil, fmt.Errorf("failed to create client for %s: %w", fieldName, err)
		}

		conns = append(conns, conn)
		// 将连接赋值给 clients 结构体对应的字段
		clientsVal.Field(i).Set(reflect.ValueOf(conn))
	}

	cleanup := func() {
		for _, conn := range conns {
			if conn != nil {
				conn.Close()
			}
		}
	}
	return clients, cleanup, nil
}
