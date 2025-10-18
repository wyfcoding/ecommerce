package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "ecommerce/api/cart/v1"
	"ecommerce/internal/cart/config"
	"ecommerce/internal/cart/model"
	"ecommerce/internal/cart/repository"
	"ecommerce/internal/cart/service"
	"ecommerce/pkg/database"
	"ecommerce/pkg/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	gorm_logger "gorm.io/gorm/logger"
)

var (
	// configPath 配置文件路径，通过命令行参数传入。
	configPath string
)

func init() {
	// 注册命令行参数。
	flag.StringVar(&configPath, "conf", "configs/cart.toml", "config path, eg: -conf configs/cart.toml")
}

func main() {
	flag.Parse()

	// 1. 加载配置
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	logger, err := log.NewLogger(&conf.Log)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger) // 将新创建的 logger 设置为全局 logger

	zap.S().Infof("cart service starting with config: %+v", conf)

	// 3. 初始化数据库
	db, err := database.NewDB(&conf.Data.Database, gorm_logger.Default.LogMode(gorm_logger.Info))
	if err != nil {
		zap.S().Fatalf("failed to connect to database: %v", err)
	}

	// 自动迁移数据库表结构
	// 注意：生产环境中，数据库迁移通常通过独立的工具或流程来管理，而不是在服务启动时自动执行。
	// 这里为了简化示例，直接在启动时执行。
	err = db.AutoMigrate(
		&model.Cart{},
		&model.CartItem{},
	)
	if err != nil {
		zap.S().Fatalf("failed to auto migrate database: %v", err)
	}
	zap.S().Info("database auto migration completed")

	// 4. 初始化 Repository 层
	cartRepo := repository.NewCartRepo(db)
	cartItemRepo := repository.NewCartItemRepo(db)

	// 5. 初始化 Service 层
	cartService := service.NewCartService(
		cartRepo,
		cartItemRepo,
		conf.Business.MaxCartItemQuantity,
		conf.Business.CartExpirationHours,
	)

	// 创建一个上下文，用于控制 gRPC 和 HTTP 服务器的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6. 启动 gRPC 服务器
	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", conf.Server.Grpc.Addr, conf.Server.Grpc.Port)
		// 创建 gRPC 服务器
		s := grpc.NewServer()
		// 注册 CartService 到 gRPC 服务器
		v1.RegisterCartServer(s, cartService)
		// 注册反射服务，用于 gRPC 客户端工具（如grpcurl）发现服务
		reflection.Register(s)

		zap.S().Infof("gRPC server listening on %s", grpcAddr)
		if err := s.Serve(log.NewGRPCListener(grpcAddr)); err != nil {
			zap.S().Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 7. 启动 HTTP Gateway (gRPC-Gateway)
	go func() {
		httpAddr := fmt.Sprintf("%s:%d", conf.Server.Http.Addr, conf.Server.Http.Port)
		grpcAddr := fmt.Sprintf("%s:%d", conf.Server.Grpc.Addr, conf.Server.Grpc.Port)

		// 创建 gRPC 客户端连接到 gRPC 服务器
		conn, err := grpc.DialContext(
			ctx,
			grpcAddr,
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
			grpc.WithBlock(),
		)
		if err != nil {
			zap.S().Fatalf("failed to dial gRPC server: %v", err)
		}
		defer conn.Close()

		// 注册 gRPC-Gateway mux
		gwmux := runtime.NewServeMux()
		err = v1.RegisterCartHandler(ctx, gwmux, conn)
		if err != nil {
			zap.S().Fatalf("failed to register gateway: %v", err)
		}

		httpServer := &http.Server{
			Addr:    httpAddr,
			Handler: gwmux,
		}

		zap.S().Infof("HTTP gateway server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("failed to serve HTTP gateway: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("shutting down server...")

	// 给服务器一个关闭的宽限期
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 在这里可以添加 gRPC 和 HTTP 服务器的优雅关闭逻辑
	// 例如：grpcServer.GracefulStop()
	//       httpServer.Shutdown(shutdownCtx)

	zap.S().Info("server exited")
}
