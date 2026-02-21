package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/wyfcoding/ecommerce/internal/iot/application"
	iotmysql "github.com/wyfcoding/ecommerce/internal/iot/infrastructure/persistence/mysql"
	"github.com/wyfcoding/pkg/response"
	"github.com/wyfcoding/pkg/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := envOrDefault("IOT_DSN", "root:password@tcp(127.0.0.1:3306)/ecommerce_iot?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&iotmysql.DeviceModel{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	repo := iotmysql.NewIoTRepository(db)
	service := application.NewIoTService(repo)

	engine := server.NewDefaultGinEngine(gin.Recovery())
	v1 := engine.Group("/api/v1/iot")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, gin.H{"status": "ok"})
		})

		v1.POST("/devices/register", func(c *gin.Context) {
			var req struct {
				Name       string `json:"name"`
				DeviceType string `json:"device_type"`
				OwnerID    string `json:"owner_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", err.Error())
				return
			}
			if req.Name == "" || req.DeviceType == "" || req.OwnerID == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid request", "name/device_type/owner_id are required")
				return
			}
			device, err := service.RegisterDevice(c.Request.Context(), req.Name, req.DeviceType, req.OwnerID)
			if err != nil {
				response.Error(c, err)
				return
			}
			response.Success(c, device)
		})

		v1.POST("/devices/:id/telemetry", func(c *gin.Context) {
			deviceID := c.Param("id")
			if deviceID == "" {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid id", "device id is required")
				return
			}
			var payload map[string]interface{}
			if err := c.ShouldBindJSON(&payload); err != nil {
				response.ErrorWithStatus(c, http.StatusBadRequest, "invalid payload", err.Error())
				return
			}
			if err := service.ReportTelemetry(c.Request.Context(), deviceID, payload); err != nil {
				response.Error(c, err)
				return
			}
			response.Success(c, gin.H{"device_id": deviceID, "reported": true})
		})

		v1.GET("/devices/:id", func(c *gin.Context) {
			deviceID := c.Param("id")
			device, err := repo.FindDeviceByID(c.Request.Context(), deviceID)
			if err != nil {
				response.Error(c, err)
				return
			}
			response.Success(c, device)
		})
	}

	addr := envOrDefault("IOT_HTTP_ADDR", ":9204")
	srv := server.NewGinServer(engine, addr, logger)
	go func() {
		if err := srv.Start(context.Background()); err != nil {
			slog.Error("server exit", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = srv.Stop(context.Background())
	slog.Info("service iot gracefully stopped")
}

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
