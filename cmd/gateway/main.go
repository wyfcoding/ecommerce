
package main

import (
	"flag"
	"fmt"

	"ecommerce/internal/gateway/handler"
	"ecommerce/pkg/app"
	"ecommerce/pkg/server"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Services struct {
		ProductService struct {
			Addr string `toml:"addr"`
		} `toml:"product_service"`
		OrderService struct {
			Addr string `toml:"addr"`
		} `toml:"order_service"`
		UserService struct {
			Addr string `toml:"addr"`
		} `toml:"user_service"`
	} `toml:"services"`
}

func main() {
	// 1. 加载配置
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/gateway.toml", "config file path")
	flag.Parse()

	var config Config
	if err := configpkg.Load(configPath, &config); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化下游服务客户端
	clients, cleanup, err := initClients(&config)
	if err != nil {
		zap.S().Fatalf("failed to init clients: %v", err)
	}
	defer cleanup()

	// 4. 创建 Gin Engine
	engine := gin.Default()

	// 5. 注册路由
	h := handler.NewHandler(clients)
	h.RegisterRoutes(engine)

	// 6. 创建并运行应用
	httpAddr := fmt.Sprintf("%s:%d", config.Server.HTTP.Addr, config.Server.HTTP.Port)
	httpSrv := server.NewGinServer(engine, httpAddr)

	application := app.New(
		app.WithServer(httpSrv),
	)

	if err := application.Run(); err != nil {
		zap.S().Fatalf("failed to run app: %v", err)
	}
}

func initClients(config *Config) (*handler.Clients, func(), error) {
	productConn, err := grpc.Dial(config.Services.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	orderConn, err := grpc.Dial(config.Services.OrderService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		productConn.Close()
		return nil, nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
    
    userConn, err := grpc.Dial(config.Services.UserService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		productConn.Close()
		orderConn.Close()
		return nil, nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	clients := handler.NewClients(productConn, orderConn, userConn)

	cleanup := func() {
		productConn.Close()
		orderConn.Close()
        userConn.Close()
	}

	return clients, cleanup, nil
}
