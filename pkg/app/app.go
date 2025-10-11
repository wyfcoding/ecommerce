package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// App 是应用程序容器。
type App struct {
	opts   options
	ctx    context.Context
	cancel func()
}

// New 创建一个新的应用程序实例。
func New(opts ...Option) *App {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		opts:   o,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run 启动应用程序并等待信号以优雅地关闭。
func (a *App) Run() error {
	// 启动所有服务器
	for _, srv := range a.opts.servers {
		go func() {
			if err := srv.Start(a.ctx); err != nil {
				zap.S().Fatalf("Server failed to start: %v", err)
			}
		}()
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutting down application...")

	// 优雅地停止所有服务器
	if a.cancel != nil {
		a.cancel()
	}

	// 为关闭操作添加超时
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range a.opts.servers {
		if err := srv.Stop(shutdownCtx); err != nil {
			zap.S().Errorf("Server failed to stop: %v", err)
			return err
		}
	}

	// 执行所有清理函数
	for _, cleanup := range a.opts.cleanups {
		cleanup()
	}

	zap.S().Info("Application shut down gracefully.")
	return nil
}
