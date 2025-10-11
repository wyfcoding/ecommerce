package main

import (
	"fmt"
	"time"

	"ecommerce/api/user/v1"
	"ecommerce/internal/user/biz"
	"ecommerce/internal/user/data"
	"ecommerce/internal/user/handler"
	"ecommerce/internal/user/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/redis"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      gormlogger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("user").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9090").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterUserServer(s, srv.(v1.UserServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	userHandler := handler.NewUserHandler(srv.(*service.UserService))
	// Define routes here, e.g.:
	e.GET("/v1/users/:id", userHandler.GetUser)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	dataInstance, cleanupData, err := data.NewData(config.Data.Database.DSN, zap.L(), config.Data.Database.LogLevel, config.Data.Database.SlowThreshold)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to new data: %w", err)
	}

	_, cleanupRedis, err := redis.NewRedisClient(&config.Data.Redis)
	if err != nil {
		cleanupData()
		return nil, nil, fmt.Errorf("failed to new redis client: %w", err)
	}

	userRepo := data.NewUserRepo(dataInstance)
	addressRepo := data.NewAddressRepo(dataInstance)

	userUsecase := biz.NewUserUsecase(userRepo)
	addressUsecase := biz.NewAddressUsecase(addressRepo)

	userService := service.NewUserService(userUsecase, addressUsecase, config.JWT.Secret, config.JWT.Issuer, config.JWT.Expire)

	cleanup := func() {
		cleanupRedis()
		cleanupData()
	}

	return userService, cleanup, nil
}