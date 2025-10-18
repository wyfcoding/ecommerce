package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewDefaultGinEngine 创建一个新的 Gin 引擎，并应用默认及自定义的中间件。
func NewDefaultGinEngine(middleware ...gin.HandlerFunc) *gin.Engine {
	engine := gin.New()

	// 应用默认中间件
	engine.Use(gin.Recovery())
	engine.Use(ZapLogger())

	// 应用外部传入的自定义中间件
	engine.Use(middleware...)

	return engine
}

// ZapLogger 是一个用于 Zap 日志记录的 Gin 中间件。
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				zap.S().Error(e)
			}
		} else {
			zap.S().Infow("Request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"ip", c.ClientIP(),
				"user-agent", c.Request.UserAgent(),
				"latency", latency,
			)
		}
	}
}