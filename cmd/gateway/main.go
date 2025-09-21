package main

import (
	"context"
	"ecommerce/pkg/logging"
	"fmt"
	"net/http"
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

	cartV1 "ecommerce/api/cart/v1"
	orderV1 "ecommerce/api/order/v1"
	productV1 "ecommerce/api/product/v1"
	userV1 "ecommerce/api/user/v1"
	adminV1 "ecommerce/api/admin/v1"
	marketingV1 "ecommerce/api/marketing/v1"
	"ecommerce/internal/gateway/internal/middleware"
)

// Config 结构体用于映射 gateway.toml 配置文件
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string        `toml:"addr"`
			Port    int           `toml:"port"`
			Timeout time.Duration `toml:"timeout"`
		} `toml:"http"`
	} `toml:"server"`
	JWT struct {
		Secret string `toml:"secret"`
	} `toml:"jwt"`
	Services map[string]struct {
		Addr string `toml:"addr"`
	} `toml:"services"`
	Log struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
		Output string `toml:"output"`
	} `toml:"log"`
}

func main() {
	// 1. 加载配置
	// 注意：实际部署时，配置文件路径可能需要调整
	config, err := loadConfig("./configs/gateway.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := logging.NewLogger(config.Log.Level, config.Log.Format, config.Log.Output)
	zap.ReplaceGlobals(logger)

	// 3. 初始化 Gin 引擎
	r := gin.New()
	r.Use(gin.Recovery())            // 使用 Gin 的 Recovery 中间件防止 panic
	r.Use(logging.GinLogger(logger)) // 使用我们自己的日志中间件

	// 4. 创建 gRPC-Gateway Mux
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. 注册所有后端的 gRPC 服务到 Gateway
	// 从配置文件动态获取服务地址，而不是硬编码
	registerServiceHandlers(ctx, gwmux, config, opts)

	// 6. 设置路由和中间件
	// 创建一个需要 JWT 认证的路由组
	v1 := r.Group("/v1")
	v1.Use(middleware.JWTAuthMiddleware(config.JWT.Secret))
	{
		// 在这个路由组下的所有请求都需要通过 JWT 认证
		// 例如：获取用户信息、创建订单、添加到购物车等
		// 注意：我们不需要在这里定义具体的路由，gRPC-Gateway 会处理
		// 我们只需要将所有 /v1/* 的请求都代理给 gRPC-Gateway 即可
	}

	// 注册豁免认证的路由
	// 登录和注册接口不需要认证，所以我们在应用认证中间件的路由组之外单独注册
	r.POST("/v1/user/register", gin.WrapH(gwmux))
	r.POST("/v1/user/login", gin.WrapH(gwmux))
	// 商品列表和详情页通常也不需要认证
	r.GET("/v1/products", gin.WrapH(gwmux))
	r.GET("/v1/products/:id", gin.WrapH(gwmux))
	r.GET("/v1/categories", gin.WrapH(gwmux))

	// 将所有其他请求（包括受保护的 /v1 组）都代理到 gRPC-Gateway
	r.Any("/*any", gin.WrapH(gwmux))

	// 7. 启动 HTTP 服务器
	httpServer, httpErrChan := startHTTPServer(config.Server.HTTP.Addr, config.Server.HTTP.Port, r)
	if httpServer == nil {
		zap.S().Fatalf("failed to start HTTP server: %v", <-httpErrChan)
	}

	// 8. 实现优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		zap.S().Info("Shutting down API Gateway...")
	case err := <-httpErrChan:
		zap.S().Errorf("HTTP server error: %v", err)
		zap.S().Info("Shutting down API Gateway due to error...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zap.S().Errorf("API Gateway shutdown error: %v", err)
	}
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer(addr string, port int, handler http.Handler) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: handler,
	}

	zap.S().Infof("API Gateway listening on %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}

// loadConfig 从 TOML 文件加载配置
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}


	return &config, nil
}

// registerServiceHandlers 动态注册所有 gRPC 服务处理器
func registerServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux, config *Config, opts []grpc.DialOption) {
	// 注册用户服务
	if userService, ok := config.Services["user"]; ok {
		if err := userV1.RegisterUserHandlerFromEndpoint(ctx, gwmux, userService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register user service: %v", err)
		}
		zap.S().Infof("Registered user service at %s", userService.Addr)
	}

	// 注册商品服务
	if productService, ok := config.Services["product"]; ok {
		if err := productV1.RegisterProductHandlerFromEndpoint(ctx, gwmux, productService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register product service: %v", err)
		}
		zap.S().Infof("Registered product service at %s", productService.Addr)
	}

	// 注册购物车服务
	if cartService, ok := config.Services["cart"]; ok {
		if err := cartV1.RegisterCartHandlerFromEndpoint(ctx, gwmux, cartService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register cart service: %v", err)
		}
		zap.S().Infof("Registered cart service at %s", cartService.Addr)
	}

	// 注册订单服务
	if orderService, ok := config.Services["order"]; ok {
		if err := orderV1.RegisterOrderHandlerFromEndpoint(ctx, gwmux, orderService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register order service: %v", err)
		}
		zap.S().Infof("Registered order service at %s", orderService.Addr)
	}

	// 注册管理服务
	if adminService, ok := config.Services["admin"]; ok {
		if err := adminV1.RegisterAdminHandlerFromEndpoint(ctx, gwmux, adminService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register admin service: %v", err)
		}
		zap.S().Infof("Registered admin service at %s", adminService.Addr)
	}

	// 注册营销服务
	if marketingService, ok := config.Services["marketing"]; ok {
		if err := marketingV1.RegisterMarketingHandlerFromEndpoint(ctx, gwmux, marketingService.Addr, opts); err != nil {
			zap.S().Fatalf("failed to register marketing service: %v", err)
		}
		zap.S().Infof("Registered marketing service at %s", marketingService.Addr)
	}
}
