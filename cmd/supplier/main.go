package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	supplierv1 "github.com/wyfcoding/ecommerce/goapi/supplier/v1"
	"github.com/wyfcoding/ecommerce/internal/supplier/application"
	"github.com/wyfcoding/ecommerce/internal/supplier/domain"
	"github.com/wyfcoding/ecommerce/internal/supplier/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/supplier/interfaces"
	"github.com/wyfcoding/pkg/config"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置 (简化)
	cfg := &config.ServerConfig{}
	cfg.Name = "supplier-service"
	cfg.GRPC.Addr = ":9005"

	// DB 连接
	dsn := "root:root@tcp(127.0.0.1:3306)/ecommerce_supplier?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	_ = db.AutoMigrate(&domain.Supplier{}, &domain.ProductSupply{})

	// 依赖注入
	repo := infrastructure.NewSupplierRepository(db)
	app := application.NewSupplierApplicationService(repo)
	handler := interfaces.NewSupplierHandler(app, repo)

	// gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	supplierv1.RegisterSupplierServiceServer(s, handler)

	fmt.Printf("server listening at %v\n", lis.Addr())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 优雅关停
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.GracefulStop()
}
