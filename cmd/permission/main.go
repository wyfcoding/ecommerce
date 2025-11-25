package main

import (
	"github.com/wyfcoding/ecommerce/pkg/app"
	"github.com/wyfcoding/ecommerce/pkg/metrics"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log/slog"
)

const serviceName = "permission-service"

func main() {
	app.NewBuilder(serviceName).
		WithConfig(&struct{}{}).
		WithService(initService).
		WithGRPC(registerGRPC).
		WithGin(registerGin).
		Build().
		Run()
}

func initService(cfg interface{}, m *metrics.Metrics) (interface{}, func(), error) {
	return nil, func() {}, nil
}

func registerGRPC(s *grpc.Server, srv interface{}) {
	slog.Default().Info("gRPC server registered for permission service")
}

func registerGin(e *gin.Engine, srv interface{}) {
	slog.Default().Info("HTTP routes registered for permission service")
}
