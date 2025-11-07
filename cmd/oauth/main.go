package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/oauth/handler"
	"ecommerce/internal/oauth/model"
	"ecommerce/internal/oauth/repository"
	"ecommerce/internal/oauth/service"
	userRepo "ecommerce/internal/user/repository"
	"ecommerce/pkg/config"
	"ecommerce/pkg/database"
	"ecommerce/pkg/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	logger := logging.NewLogger(cfg.Log.Level, cfg.Log.Filename)
	defer logger.Sync()

	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}

	// OAuth配置
	oauthConfigs := map[model.OAuthProvider]*model.OAuthConfig{
		model.OAuthProviderWechat: {
			Provider:    model.OAuthProviderWechat,
			AppID:       cfg.OAuth.Wechat.AppID,
			AppSecret:   cfg.OAuth.Wechat.AppSecret,
			RedirectURL: cfg.OAuth.Wechat.RedirectURL,
			Scope:       "snsapi_userinfo",
			AuthURL:     "https://open.weixin.qq.com/connect/oauth2/authorize",
			TokenURL:    "https://api.weixin.qq.com/sns/oauth2/access_token",
			UserInfoURL: "https://api.weixin.qq.com/sns/userinfo",
		},
		model.OAuthProviderQQ: {
			Provider:    model.OAuthProviderQQ,
			AppID:       cfg.OAuth.QQ.AppID,
			AppSecret:   cfg.OAuth.QQ.AppSecret,
			RedirectURL: cfg.OAuth.QQ.RedirectURL,
			Scope:       "get_user_info",
			AuthURL:     "https://graph.qq.com/oauth2.0/authorize",
			TokenURL:    "https://graph.qq.com/oauth2.0/token",
			UserInfoURL: "https://graph.qq.com/user/get_user_info",
		},
	}

	repo := repository.NewOAuthRepo(db)
	userRepository := userRepo.NewUserRepo(db)
	svc := service.NewOAuthService(repo, userRepository, oauthConfigs, logger)
	h := handler.NewOAuthHandler(svc, logger)

	r := gin.Default()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{Addr: ":8010", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	logger.Info("OAuth服务启动成功", zap.Int("port", 8010))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}
