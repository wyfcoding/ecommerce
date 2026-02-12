package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	i18nv1 "github.com/wyfcoding/ecommerce/go-api/i18n/v1"
	"github.com/wyfcoding/ecommerce/internal/i18n/application"
	"github.com/wyfcoding/ecommerce/internal/i18n/domain"
	"github.com/wyfcoding/ecommerce/internal/i18n/infrastructure"
	"github.com/wyfcoding/ecommerce/internal/i18n/interfaces"
	"github.com/wyfcoding/pkg/config"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := &config.ServerConfig{}
	cfg.Name = "i18n-service"
	cfg.GRPC.Addr = ":9006"

	// DB 连接
	dsn := "root:root@tcp(127.0.0.1:3306)/ecommerce_i18n?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	_ = db.AutoMigrate(&domain.Language{}, &domain.Translation{})

	// 依赖注入
	repo := infrastructure.NewI18nRepository(db)
	app := application.NewI18nService(repo)
	handler := interfaces.NewI18nHandler(app, repo)

	// 初始化默认数据
	go func() {
		time.Sleep(3 * time.Second)
		_ = app.InitDefaults(context.Background())
	}()

	// gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	i18nv1.RegisterI18NServiceServer(s, handler)

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
