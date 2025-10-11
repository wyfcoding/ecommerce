package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinServer 是一个 Gin HTTP 服务器实现。
type GinServer struct {
	srv *http.Server
}

// NewGinServer 创建一个新的 Gin HTTP 服务器。
func NewGinServer(engine *gin.Engine, addr string) *GinServer {
	return &GinServer{
		srv: &http.Server{
			Addr:    addr,
			Handler: engine,
		},
	}
}

// Start 启动 Gin HTTP 服务器。
func (s *GinServer) Start(ctx context.Context) error {
	zap.S().Infof("HTTP server listening at %s", s.srv.Addr) // 日志信息保持英文
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop 停止 Gin HTTP 服务器。
func (s *GinServer) Stop(ctx context.Context) error {
	zap.S().Info("Stopping HTTP server...") // 日志信息保持英文
	return s.srv.Shutdown(ctx)
}
