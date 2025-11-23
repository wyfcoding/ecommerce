package app

import (
	"context"
	"ecommerce/pkg/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App 是应用程序容器。
type App struct {
	name   string
	logger *slog.Logger
	opts   options
	ctx    context.Context
	cancel func()
}

// New 创建一个新的应用程序实例。
func New(name string, logger *slog.Logger, opts ...Option) *App {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		name:   name,
		logger: logger,
		opts:   o,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run 启动应用程序并等待信号以优雅地关闭。
func (a *App) Run() error {
	// 启动所有服务器
	for _, srv := range a.opts.servers {
		go func(s server.Server) {
			if err := s.Start(a.ctx); err != nil {
				a.logger.Error("Server failed to start", "error", err)
				// Optionally, signal shutdown if a critical server fails to start
				a.cancel()
			}
		}(srv)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.logger.Info("Shutting down application", "name", a.name)

	// 优雅地停止所有服务器
	if a.cancel != nil {
		a.cancel()
	}

	// 为关闭操作添加超时
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range a.opts.servers {
		if err := srv.Stop(shutdownCtx); err != nil {
			a.logger.Error("Server failed to stop", "error", err)
			return err
		}
	}

	// 执行所有清理函数
	for _, cleanup := range a.opts.cleanups {
		cleanup()
	}

	a.logger.Info("Application shut down gracefully.")
	return nil
}
