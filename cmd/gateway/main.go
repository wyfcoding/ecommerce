
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	gatewayclient "ecommerce/internal/gateway/client"
	gatewayhandler "ecommerce/internal/gateway/handler"
	"ecommerce/internal/gateway/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/server"
	"ecommerce/pkg/tracing"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.ServerConfig `toml:"server"`
	configpkg.LogConfig    `toml:"log"`
	configpkg.TraceConfig  `toml:"trace"`
	Metrics                struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
	Services struct {
		AuthService struct {
			Addr string `toml:"addr"`
		} `toml:"auth_service"`
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

	var cfg Config
	if err := configpkg.LoadConfig(configPath, &cfg); err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// 3. 初始化追踪
	_, cleanupTracing, err := tracing.InitTracer(&cfg.Trace)
	if err != nil {
		zap.S().Fatalf("failed to init tracing: %v", err)
	}
	defer cleanupTracing()

	// 4. 初始化指标暴露
	cleanupMetrics := metrics.ExposeHttp(cfg.Metrics.Port)
	defer cleanupMetrics()

	// 5. 初始化下游服务客户端
	clients, cleanupClients, err := gatewayclient.NewClients(&cfg.Services.AuthService.Addr, &cfg.Services.ProductService.Addr, &cfg.Services.OrderService.Addr, &cfg.Services.UserService.Addr)
	if err != nil {
		zap.S().Fatalf("failed to init clients: %v", err)
	}
	defer cleanupClients()

	// 6. 依赖注入 (DI)
	gatewayService := service.NewGatewayService(clients.AuthClient)

	// 7. 创建 Gin Engine
	engine := gin.Default()

	// 8. 注册路由
	h := gatewayhandler.NewHandler(gatewayService)
	h.RegisterRoutes(engine)

	// 9. 创建并运行应用
	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	httpSrv := server.NewGinServer(engine, httpAddr)

	application := app.New(
		app.WithServer(httpSrv),
	)

	if err := application.Run(); err != nil {
		zap.S().Fatalf("failed to run app: %v", err)
	}
}
