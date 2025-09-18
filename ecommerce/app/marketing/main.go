package main

import (
	"log"
	"net"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartV1 "ecommerce/ecommerce/api/cart/v1"
	"ecommerce/ecommerce/app/cart/internal/biz"
	"ecommerce/ecommerce/app/cart/internal/data"
	"ecommerce/ecommerce/app/cart/internal/service"
)

const (
	cartServicePort    = ":9002"
	productServiceAddr = "localhost:9001"
	redisAddr          = "localhost:6379"
)

func main() {
	// --- 1. 初始化依赖 ---

	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 初始化 Product Service 的 gRPC 客户端连接
	conn, err := grpc.Dial(productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to product service: %v", err)
	}
	defer conn.Close()

	// --- 2. 依赖注入 (DI) ---
	productGreeter := data.NewProductGreeter(conn)
	cartRepo := data.NewCartRepo(rdb)
	cartUsecase := biz.NewCartUsecase(cartRepo, productGreeter)
	cartService := service.NewCartService(cartUsecase)

	// --- 3. 启动 gRPC 服务器 ---
	lis, err := net.Listen("tcp", cartServicePort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	cartV1.RegisterCartServer(s, cartService)

	log.Printf("gRPC server (cart-service) listening at %s", cartServicePort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
