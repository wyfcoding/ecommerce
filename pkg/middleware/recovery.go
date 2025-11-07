package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 恢复中间件，捕获panic并记录日志
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic堆栈信息
				zap.S().Errorf("panic recovered: %v\n%s", err, debug.Stack())

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("internal server error: %v", err)})
				c.Abort()
			}
		}()
		c.Next()
	}
}
