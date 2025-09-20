package main

import (
	// ...
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// ...
	"ecommerce/ecommerce/app/admin/internal/biz"
	"ecommerce/ecommerce/app/admin/internal/data"
	"ecommerce/ecommerce/app/admin/internal/service"
)

func main() {
	// ... 数据库等依赖初始化 ...

	// 初始化到 Order Service 的 gRPC 客户端连接
	orderSvcConn, err := grpc.Dial(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to order service: %v", err)
	}
	defer orderSvcConn.Close()

	// 初始化到 Product Service 的 gRPC 客户端连接
	productSvcConn, err := grpc.Dial(productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to product service: %v", err)
	}
	defer productSvcConn.Close()
	// JWT 密钥应该从配置中读取
	jwtSecret := "a_different_secret_for_admin"

	// 依赖注入
	authRepo := data.NewAuthRepo(db)
	authUsecase := biz.NewAuthUsecase(authRepo, jwtSecret)
	adminService := service.NewAdminService(authUsecase)

	// 新增商品管理相关的依赖注入
	productGreeter := data.NewProductGreeter(productSvcConn)
	productUsecase := biz.NewProductUsecase(productGreeter)

	// 更新 AdminService 的构造函数调用
	adminService := service.NewAdminService(authUsecase, productUsecase)

	// 创建权限拦截器
	authInterceptor := service.NewAuthInterceptor(authUsecase, jwtSecret)

	// 创建 gRPC 服务器并应用拦截器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Auth),
	)

	// 注册服务
	v1.RegisterAdminServer(s, adminService)

	// ... 启动服务器 ...
}
