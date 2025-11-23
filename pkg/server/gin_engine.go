package server

import (
	"ecommerce/pkg/middleware"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// NewDefaultGinEngine 创建一个新的 Gin 引擎，并应用默认及自定义的中间件。
func NewDefaultGinEngine(logger *slog.Logger, middlewares ...gin.HandlerFunc) *gin.Engine {
	engine := gin.New()

	// 应用默认中间件
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger(logger))

	// 应用外部传入的自定义中间件
	engine.Use(middlewares...)

	return engine
}
