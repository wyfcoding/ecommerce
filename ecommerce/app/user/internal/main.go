package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	v1 "ecommerce/ecommerce/api/user/v1"
	"ecommerce/ecommerce/app/user/internal/biz"
	"ecommerce/ecommerce/app/user/internal/data"
	"ecommerce/ecommerce/app/user/internal/service"
	"ecommerce/ecommerce/pkg/snowflake"
)

func main() {
	// --- 1. 初始化依赖 ---

	// 初始化 Snowflake
	// 注意：startTime 和 machineID 应该从配置文件中读取
	if err := snowflake.Init("2025-01-01", 1); err != nil {
		log.Fatalf("failed to init snowflake: %v", err)
	}

	// 初始化数据库连接
	// DSN (Data Source Name) 应该从配置文件中读取
	dsn := "root:your_password@tcp(127.0.0.1:3306)/genesis_user?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// --- 2. 依赖注入 (DI) ---
	// 这是我们分层架构的核心体现
	userRepo := data.NewUserRepo(db)
	userUsecase := biz.NewUserUsecase(userRepo)
	userService := service.NewUserService(userUsecase)

	// --- 3. 启动 gRPC 服务器 ---
	lis, err := net.Listen("tcp", ":9000") // 端口应从配置中读取
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	v1.RegisterUserServer(s, userService)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
