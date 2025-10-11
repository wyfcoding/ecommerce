package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Project packages
	"ecommerce/pkg/config"
	"ecommerce/pkg/database/mysql"
	"ecommerce/pkg/logging"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	// Service-specific imports

	couponv1 "ecommerce/api/coupon/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	reviewv1 "ecommerce/api/review/v1"
	userv1 "ecommerce/api/user/v1"
	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/data"
	adminhandler "ecommerce/internal/admin/handler"
	"ecommerce/internal/admin/service"
	couponbiz "ecommerce/internal/coupon/biz"
	coupondata "ecommerce/internal/coupon/data"
	orderbiz "ecommerce/internal/order/biz"
	orderdata "ecommerce/internal/order/data"
	productbiz "ecommerce/internal/product/biz"
	productdata "ecommerce/internal/product/data"
	reviewbiz "ecommerce/internal/review/biz"
	reviewdata "ecommerce/internal/review/data"
	userbiz "ecommerce/internal/user/biz"
	userdata "ecommerce/internal/user/data"
)

// Config 结构体用于映射 TOML 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
		GRPC struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"grpc"`
	} `toml:"server"`
	Data struct {
		Database mysql.Config `toml:"database"` // 使用 pkg/database/mysql.Config
		// ... 其他服务客户端配置
		ProductService struct {
			Addr string `toml:"addr"`
		} `toml:"product_service"`
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
		Secret string        `toml:"secret"`
		Issuer string        `toml:"issuer"`
		Expire time.Duration `toml:"expire_duration"`
	} `toml:"jwt"`
	Log     logging.Config `toml:"log"`   // 使用 pkg/logging.Config
	Trace   tracing.Config `toml:"trace"` // 使用 pkg/tracing.Config
	Metrics struct {
		Port string `toml:"port"`
	} `toml:"metrics"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "./configs/admin.toml", "config file path")
	flag.Parse()

	// 1. 加载配置
	var cfg Config
	if err := config.LoadConfig(configPath, &cfg); err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // 刷新任何缓冲的日志条目

	// 3. 初始化追踪
	_, cleanupTracing, err := tracing.InitTracer(&cfg.Trace)
	if err != nil {
		zap.S().Fatalf("failed to init tracing: %v", err)
	}
	defer cleanupTracing()

	// 4. 初始化指标暴露
	cleanupMetrics := metrics.ExposeHttp(cfg.Metrics.Port)
	defer cleanupMetrics()

	// 5. 初始化依赖
	// 数据库连接
	db, cleanupDB, err := mysql.NewGORMDB(&cfg.Data.Database)
	if err != nil {
		zap.S().Fatalf("failed to connect database: %v", err)
	}
	defer cleanupDB()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&adminmodel.AdminUser{}, &adminmodel.Role{}, &adminmodel.Permission{}, &adminmodel.AdminUserRole{}); err != nil {
		zap.S().Fatalf("failed to migrate database: %v", err)
	}

	// gRPC 客户端连接
	// ... (现有的 gRPC 客户端连接，确保延迟清理)
	// 商品服务客户端
	productServiceConn, err := grpc.DialContext(context.Background(), cfg.Data.ProductService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to product service: %v", err)
	}
	defer productServiceConn.Close()
	productClient := productv1.NewProductClient(productServiceConn)
	productData := productdata.NewData(db) // 假设商品服务有自己的数据层
	productRepo := productdata.NewProductRepo(productData)
	productUsecase := productbiz.NewProductUsecase(productRepo)

	// 订单服务客户端
	orderServiceConn, err := grpc.DialContext(context.Background(), cfg.Data.OrderService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to order service: %v", err)
	}
	defer orderServiceConn.Close()
	orderClient := orderv1.NewOrderClient(orderServiceConn)
	orderData := orderdata.NewData(db) // 假设订单服务有自己的数据层
	orderRepo := orderdata.NewOrderRepo(orderData)
	orderUsecase := orderbiz.NewOrderUsecase(orderRepo)

	// 用户服务客户端
	userServiceConn, err := grpc.DialContext(context.Background(), cfg.Data.UserService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to user service: %v", err)
	}
	defer userServiceConn.Close()
	userClient := userv1.NewUserClient(userServiceConn)
	userData := userdata.NewData(db) // 假设用户服务有自己的数据层
	userRepo := userdata.NewUserRepo(userData)
	userUsecase := userbiz.NewUserUsecase(userRepo)

	// 评论服务客户端
	reviewServiceConn, err := grpc.DialContext(context.Background(), cfg.Data.ReviewService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to review service: %v", err)
	}
	defer reviewServiceConn.Close()
	reviewClient := reviewv1.NewReviewClient(reviewServiceConn)
	reviewData := reviewdata.NewData(db) // 假设评论服务有自己的数据层
	reviewRepo := reviewdata.NewReviewRepo(reviewData)
	reviewUsecase := reviewbiz.NewReviewUsecase(reviewRepo)

	// 优惠券服务客户端
	couponServiceConn, err := grpc.DialContext(context.Background(), cfg.Data.CouponService.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		zap.S().Fatalf("failed to connect to coupon service: %v", err)
	}
	defer couponServiceConn.Close()
	couponClient := couponv1.NewCouponClient(couponServiceConn)
	couponData := coupondata.NewData(db) // 假设优惠券服务有自己的数据层
	couponRepo := coupondata.NewCouponRepo(couponData)
	couponUsecase := couponbiz.NewCouponUsecase(couponRepo)

	// 6. 依赖注入 (DI)
	adminData := data.NewData(db)
	adminRepo := data.NewAdminRepo(adminData)
	authUsecase := biz.NewAuthUsecase(adminRepo, cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Expire)
	adminUsecase := biz.NewAdminUsecase(adminRepo, authUsecase, productUsecase, userUsecase, orderUsecase, reviewUsecase, couponUsecase)

	adminService := service.NewAdminService(authUsecase, productUsecase, userUsecase, orderUsecase, reviewUsecase, couponUsecase) // 整合所有 usecase

	// 创建权限拦截器
	authInterceptor := service.NewAuthInterceptor(authUsecase, cfg.JWT.Secret)

	// 7. 启动 gRPC 服务器
	grpcServer, grpcErrChan := adminhandler.StartGRPCServer(adminService, authInterceptor, cfg.Server.GRPC.Addr, cfg.Server.GRPC.Port)
	if grpcServer == nil {
		zap.S().Fatalf("failed to start gRPC server: %v", <-grpcErrChan)
	}

	// 8. 启动 Gin HTTP 服务器
	ginServer, ginErrChan := adminhandler.StartGinServer(adminService, cfg.Server.HTTP.Addr, cfg.Server.HTTP.Port)
	if ginServer == nil {
		zap.S().Fatalf("failed to start Gin HTTP server: %v", <-ginErrChan)
	}

	// 9. 等待中断信号或服务器错误以实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down admin service...")
	case err := <-grpcErrChan:
		zap.S().Errorf("gRPC server error: %v", err)
		zap.S().Info("Shutting down admin service due to gRPC error...")
	case err := <-ginErrChan:
		zap.S().Errorf("Gin HTTP server error: %v", err)
		zap.S().Info("Shutting down admin service due to Gin HTTP error...")
	}

	// 优雅地关闭服务器
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ginServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("Gin HTTP server shutdown error: %v", err)
	}
}

// startGRPCServer 启动 gRPC 服务器

// startGinServer 启动 Gin HTTP 服务器
