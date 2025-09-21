package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	v1 "ecommerce/api/product/v1"
	"ecommerce/app/product/internal/biz"
	"ecommerce/app/product/internal/data"
	"ecommerce/app/product/internal/service"
	"ecommerce/pkg/snowflake"
)

func main() {
	// --- 1. 初始化依赖 ---

	// 初始化 Snowflake
	// 注意：在真实项目中，startTime 和 machineID 应该从配置文件中读取
	if err := snowflake.Init("2025-01-01", 2); err != nil { // 使用不同的 machineID
		log.Fatalf("failed to init snowflake: %v", err)
	}

	// 初始化数据库连接
	// DSN (Data Source Name) 应该从配置文件中读取
	dsn := "root:your_password@tcp(127.0.0.1:3306)/genesis_product?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// --- 2. 依赖注入 (DI) - 串联三层架构 ---
	productRepo := data.NewProductRepo(db)
	productUsecase := biz.NewProductUsecase(productRepo)
	productService := service.NewProductService(productUsecase)

	// --- 3. 启动 gRPC 服务器 ---
	// 端口应从配置中读取
	port := ":9001"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	v1.RegisterProductServer(s, productService)

	log.Printf("gRPC server (product-service) listening at %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
