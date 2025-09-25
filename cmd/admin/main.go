package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	v1 "ecommerce/api/admin/v1"
	couponv1 "ecommerce/api/coupon/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	reviewv1 "ecommerce/api/review/v1"
	userv1 "ecommerce/api/user/v1"
	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/data"
	adminmodel "ecommerce/internal/admin/data/model"
	couponbiz "ecommerce/internal/coupon/biz"
	coupondata "ecommerce/internal/coupon/data"
	orderbiz "ecommerce/internal/order/biz"
	orderdata "ecommerce/internal/order/data"
	productbiz "ecommerce/internal/product/biz"
	productdata "ecommerce/internal/product/data"
	reviewbiz "ecommerce/internal/review/biz"
	reviewdata "ecommerce/internal/review/data"
	"ecommerce/internal/admin/service"
	userbiz "ecommerce/internal/user/biz"
	userdata "ecommerce/internal/user/data"
	"ecommerce/pkg/logging"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string `toml:"addr"`
			Port    int    `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database struct {
			DSN string `toml:"dsn"`
		} `toml:"database"`
		ProductService struct {
			Addr string `toml:"addr"`		} `toml:"product_service"`
		OrderService struct {
			Addr string `toml:"addr"`
		} `toml:"order_service"`
		UserService struct {
			Addr string `toml:"addr"`
		} `toml:"user_service"`
		ReviewService struct {
			Addr string `toml:"addr"`
		} `toml:"review_service"`
		CouponService struct {
			Addr string `toml:"addr"`
		} `toml:"coupon_service"`
	} `toml:"data"`
	JWT struct {
		Secret string `toml:"secret"`
		Issuer string `toml:"issuer"`
		Expire time.Duration `toml:"expire_duration"`
	} `toml:"jwt"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`		Output string `toml:"output"`
	} `toml:"log"`
}

// loadConfig 从指定路径加载并解析 TOML 配置文件
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/admin.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	config, err := loadConfig(configPath)
	if err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化依赖
	// 数据库连接
	db, err := gorm.Open(mysql.Open(config.Data.Database.DSN), &gorm.Config{})
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&adminmodel.AdminUser{}, &adminmodel.Role{}, &adminmodel.Permission{}, &adminmodel.AdminUserRole{}); err != nil {
		zap.S().Fatalf("failed to migrate database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Fatalf("failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// gRPC 客户端连接
	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Product Service Client
	productServiceConn, err := grpc.DialContext(dialCtx, config.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to product service: %v", err)
	}
	defer productServiceConn.Close()
	productClient := productv1.NewProductClient(productServiceConn)
	productData := productdata.NewData(db) // Assuming product service has its own data layer
	productRepo := productdata.NewProductRepo(productData)
	productUsecase := productbiz.NewProductUsecase(productRepo)

	// Order Service Client
	orderServiceConn, err := grpc.DialContext(dialCtx, config.Data.OrderService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to order service: %v", err)
	}
	defer orderServiceConn.Close()
	orderClient := orderv1.NewOrderClient(orderServiceConn)
	orderData := orderdata.NewData(db) // Assuming order service has its own data layer
	orderRepo := orderdata.NewOrderRepo(orderData)
	orderUsecase := orderbiz.NewOrderUsecase(orderRepo)

	// User Service Client
	userServiceConn, err := grpc.DialContext(dialCtx, config.Data.UserService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to user service: %v", err)
	}
	defer userServiceConn.Close()
	userClient := userv1.NewUserClient(userServiceConn)
	userData := userdata.NewData(db) // Assuming user service has its own data layer
	userRepo := userdata.NewUserRepo(userData)
	userUsecase := userbiz.NewUserUsecase(userRepo)

	// Review Service Client
	reviewServiceConn, err := grpc.DialContext(dialCtx, config.Data.ReviewService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to review service: %v", err)
	}
	defer reviewServiceConn.Close()
	reviewClient := reviewv1.NewReviewClient(reviewServiceConn)
	reviewData := reviewdata.NewData(db) // Assuming review service has its own data layer
	reviewRepo := reviewdata.NewReviewRepo(reviewData)
	reviewUsecase := reviewbiz.NewReviewUsecase(reviewRepo)

	// Coupon Service Client
	couponServiceConn, err := grpc.DialContext(dialCtx, config.Data.CouponService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to coupon service: %v", err)
	}
	defer couponServiceConn.Close()
	couponClient := couponv1.NewCouponClient(couponServiceConn)
	couponData := coupondata.NewData(db) // Assuming coupon service has its own data layer
	couponRepo := coupondata.NewCouponRepo(couponData)
	couponUsecase := couponbiz.NewCouponUsecase(couponRepo)

	// 4. 依赖注入 (DI)
	adminData := data.NewData(db)
	adminRepo := data.NewAdminRepo(adminData)
	authUsecase := biz.NewAuthUsecase(adminRepo, config.JWT.Secret, config.JWT.Issuer, config.JWT.Expire)
	adminUsecase := biz.NewAdminUsecase(adminRepo, authUsecase, productUsecase, userUsecase, orderUsecase, reviewUsecase, couponUsecase)

	adminService := service.NewAdminService(authUsecase, productUsecase, userUsecase, orderUsecase, reviewUsecase, couponUsecase) // 整合所有 usecase

	// 创建权限拦截器
	authInterceptor := service.NewAuthInterceptor(authUsecase, config.JWT.Secret)

	// 5. 启动 gRPC 和 HTTP Gateway
	grpcServer, grpcErrChan := startGRPCServer(adminService, authInterceptor, config.Server.GRPC.Addr, config.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}
	httpServer, httpErrChan := startHTTPServer(context.Background(), config.Server.GRPC.Addr, config.Server.GRPC.Port, config.Server.HTTP.Addr, config.Server.HTTP.Port)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 6. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down admin service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down admin service due to gRPC error...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down admin service due to HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("HTTP server shutdown error: %v", err)
	}
}

// startGRPCServer 启动 gRPC 服务器
func startGRPCServer(adminService *service.AdminService, authInterceptor *service.AuthInterceptor, addr string, port int) (*grpc.Server, chan error) {
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		errChan <- fmt.Errorf("failed to listen: %w", err)
		return nil, errChan
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Auth),
	)
	v1.RegisterAdminServer(s, adminService)

	zap.S().Infof("gRPC server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
		}
		close(errChan)
	}()
	return s, errChan
}

// startHTTPServer 启动 HTTP Gateway，它会将 HTTP 请求代理到 gRPC 服务
func startHTTPServer(ctx context.Context, grpcAddr string, grpcPort int, httpAddr string, httpPort int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("%s:%d", grpcAddr, grpcPort)

	err := v1.RegisterAdminHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		errChan <- fmt.Errorf("failed to register gRPC gateway for AdminService: %w", err)
		return nil, errChan
	}

	r := gin.Default()
	r.Use(gin.Recovery())
	// Add service-specific Gin routes here
	// For example:
	// r.GET("/admin/users/:id", handler.GetAdminUser)
	// r.POST("/admin/users", handler.CreateAdminUser)

	r.Any("/*any", gin.WrapH(mux))

	httpEndpoint := fmt.Sprintf("%s:%d", httpAddr, httpPort)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}
