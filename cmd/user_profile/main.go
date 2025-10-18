package main

import (
	"fmt"
	"time"

	v1 "ecommerce/api/user_profile/v1"
	"ecommerce/internal/user_profile/handler"
	"ecommerce/internal/user_profile/model"
	"ecommerce/internal/user_profile/repository"
	"ecommerce/internal/user_profile/service"
	"ecommerce/pkg/app"
	configpkg "ecommerce/pkg/config"
	mysqlpkg "ecommerce/pkg/database/mysql"
	redisPkg "ecommerce/pkg/database/redis"
	"ecommerce/pkg/metrics"
	"ecommerce/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm/logger"
)

// Config is the service-specific configuration structure.
type Config struct {
	configpkg.Config
	Data struct {
		configpkg.DataConfig
		Database struct {
			configpkg.DatabaseConfig
			LogLevel      logger.LogLevel `toml:"log_level"`
			SlowThreshold time.Duration     `toml:"slow_threshold"`
		} `toml:"database"`
		Redis redisPkg.Config `toml:"redis"`
	} `toml:"data"`
}

func main() {
	app.NewBuilder("user_profile").
		WithConfig(&Config{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		WithMetrics("9096").
		Build().
		Run()
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	v1.RegisterUserProfileServiceServer(s, srv.(v1.UserProfileServiceServer))
}

func registerGin(e *gin.Engine, srv interface{}) {
	userProfileHandler := handler.NewUserProfileHandler(srv.(*service.UserProfileService))
	// e.g., e.GET("/v1/profiles/:id", userProfileHandler.GetProfile)
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	config := cfg.(*Config)

	db, cleanupDB, err := mysqlpkg.NewGORMDB(&config.Data.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(&model.UserProfile{}, &model.UserBehaviorEvent{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	redisClient, cleanupRedis, err := redisPkg.NewRedisClient(&config.Data.Redis)
	if err != nil {
		cleanupDB()
		return nil, nil, fmt.Errorf("failed to new redis client: %w", err)
	}

	userProfileRepo := repository.NewUserProfileRepo(db, redisClient)
	userProfileService := service.NewUserProfileService(userProfileRepo)

	cleanup := func() {
		cleanupRedis()
		cleanupDB()
	}

	return userProfileService, cleanup, nil
}
