package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"ecommerce/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件，捕获panic并记录日志
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic堆栈信息
				logger.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)

				// 返回500错误
				response.ErrorWithStatus(c, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
