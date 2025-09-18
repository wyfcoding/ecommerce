package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartV1 "ecommerce/ecommerce/api/cart/v1"
	orderV1 "ecommerce/ecommerce/api/order/v1"
	productV1 "ecommerce/ecommerce/api/product/v1"
	userV1 "ecommerce/ecommerce/api/user/v1"
	"ecommerce/ecommerce/gateway/internal/middleware"
)

const (
	gatewayAddr        = ":8080"
	userServiceAddr    = "localhost:9000"
	productServiceAddr = "localhost:9001"
	cartServiceAddr    = "localhost:9002"
	orderServiceAddr   = "localhost:9003"
)

// handleAlipayNotify 是一个特殊的 HTTP Handler，用于处理支付宝的异步通知
func handleAlipayNotify(orderSvcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("failed to parse form: %v", err)
			return
		}

		// 1. 将 Form 数据转换为 map
		notifyData := make(map[string]string)
		for key, values := range r.Form {
			notifyData[key] = values[0]
		}

		// 2. 创建到 order-service 的 gRPC 连接
		conn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("failed to connect to order service: %v", err)
			return
		}
		defer conn.Close()

		// 3. 调用 gRPC 接口
		client := orderV1.NewOrderClient(conn)
		_, err = client.ProcessPaymentNotification(context.Background(), &orderV1.ProcessPaymentNotificationRequest{
			NotificationData: notifyData,
		})

		// 4. 根据支付宝的要求，返回 "success" 或 "failure"
		if err != nil {
			log.Printf("failed to process notification: %v", err)
			w.Write([]byte("failure"))
		} else {
			w.Write([]byte("success"))
		}
	}
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 gRPC-Gateway 的 Mux
	mux := runtime.NewServeMux()

	// --- 注册 User Service ---
	optsUser := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := userV1.RegisterUserHandlerFromEndpoint(ctx, mux, userServiceAddr, optsUser)
	if err != nil {
		log.Fatalf("failed to register user service: %v", err)
	}

	// --- 注册 Product Service ---
	optsProduct := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = productV1.RegisterProductHandlerFromEndpoint(ctx, mux, productServiceAddr, optsProduct)
	if err != nil {
		log.Fatalf("failed to register product service: %v", err)
	}

	log.Printf("API Gateway listening on %s", gatewayAddr)
	log.Printf("Proxying requests to user-service at %s and product-service at %s", userServiceAddr, productServiceAddr)

	// --- 注册 Cart Service ---
	optsCart := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = cartV1.RegisterCartHandlerFromEndpoint(ctx, mux, cartServiceAddr, optsCart)
	if err != nil {
		log.Fatalf("failed to register cart service: %v", err)
	}

	log.Printf("Proxying requests to cart-service at %s", cartServiceAddr)

	// --- 注册 Order Service ---
	optsOrder := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = orderV1.RegisterOrderHandlerFromEndpoint(ctx, mux, orderServiceAddr, optsOrder)
	if err != nil {
		log.Fatalf("failed to register order service: %v", err)
	}

	log.Printf("Proxying requests to order-service at %s", orderServiceAddr)

	// --- 定义服务器和中间件 ---
	server := &http.Server{
		Addr: gatewayAddr,
	}

	// 定义不需要认证的路径
	unprotectedPaths := []string{
		"/api.user.v1.User/RegisterByPassword", // 注册接口
		"/api.user.v1.User/LoginByPassword",    // 登录接口
		"/v1/user/register",                    // HTTP 注解对应的路径
		"/v1/user/login",
		"/v1/categories",    // 商品分类通常不需要登录
		"/v1/products/spu/", // 商品详情通常不需要登录
	}

	JWTSecret := []byte("your-very-secret-key")
	// 应用 JWT 认证中间件
	// 注意：JWTSecret 应该从统一的配置中心加载，这里为了演示，暂时借用 user/biz 的定义
	jwtAuthHandler := middleware.JWTAuthMiddleware(JWTSecret, unprotectedPaths)(mux)

	server.Handler = jwtAuthHandler

	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to serve: %v", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux) // gRPC-Gateway 的 handler

	// 注册支付宝回调的专用 handler
	httpMux.Handle("/v1/payment/alipay/notify", handleAlipayNotify(orderServiceAddr))

	// ...
	server := &http.Server{
		Addr:    gatewayAddr,
		Handler: jwtAuthHandler(httpMux), // 将新的 httpMux 包装进中间件
	}
}
