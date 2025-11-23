package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	server *http.Server
	addr   string
	logger *slog.Logger
}

func NewGinServer(engine *gin.Engine, addr string, logger *slog.Logger) *GinServer {
	return &GinServer{
		server: &http.Server{
			Addr:    addr,
			Handler: engine,
		},
		addr:   addr,
		logger: logger,
	}
}

func (s *GinServer) Start(ctx context.Context) error {
	s.logger.Info("Starting Gin server", "addr", s.addr)

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

func (s *GinServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Gin server")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
