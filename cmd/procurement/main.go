package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	procurementv1 "github.com/wyfcoding/ecommerce/goapi/procurement/v1"
	"github.com/wyfcoding/ecommerce/internal/procurement/application"
	"github.com/wyfcoding/ecommerce/internal/procurement/domain"
	"github.com/wyfcoding/ecommerce/internal/procurement/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/procurement/interfaces"
	"github.com/wyfcoding/pkg/config"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置 (简化)
	cfg := &config.ServerConfig{}
	cfg.Name = "procurement-service"
	cfg.GRPC.Addr = ":9004"

	// DB 连接
	dsn := "root:root@tcp(127.0.0.1:3306)/ecommerce_procurement?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	_ = db.AutoMigrate(&domain.PurchaseRequest{}, &domain.PurchaseRequestItem{}, &domain.PurchaseOrder{}, &domain.PurchaseOrderItem{})

	// 依赖注入
	repo := infrastructure.NewProcurementRepository(db)
	cmd := application.NewProcurementCommandService(repo)
	handler := interfaces.NewProcurementHandler(cmd, repo)

	// gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	procurementv1.RegisterProcurementServiceServer(s, handler)

	fmt.Printf("%s listening at %v\n", cfg.Name, lis.Addr())

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
