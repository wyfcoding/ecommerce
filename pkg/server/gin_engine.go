package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewDefaultGinEngine creates a new Gin engine with default middleware.
func NewDefaultGinEngine() *gin.Engine {
	engine := gin.New()

	// Add recovery middleware
	engine.Use(gin.Recovery())

	// Add logging middleware
	engine.Use(ZapLogger())

	return engine
}

// ZapLogger is a Gin middleware for logging requests using Zap.
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			// Log errors
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
