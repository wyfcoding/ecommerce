package aftersaleshandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/internal/aftersales/service"
	"ecommerce/pkg/logging"
)

// StartGinServer 启动 Gin HTTP 服务器
func StartGinServer(aftersalesService *service.AftersalesService, addr string, port int) (*http.Server, chan error) {
	errChan := make(chan error, 1)
	r := gin.New()
	r.Use(logging.GinLogger(zap.L()), gin.Recovery()) // 使用项目的 GinLogger

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- 在此处注册服务特定的 HTTP 路由 ---
	// 示例：
	// aftersalesHandler := handler.NewAftersalesHandler(aftersalesService)
	// r.GET("/api/v1/aftersales/requests/:id", aftersalesHandler.GetAftersalesRequest)
	// r.POST("/api/v1/aftersales/requests", aftersalesHandler.CreateAftersalesRequest)
	//
	// --- 应用级降级 ---
	// 降级逻辑高度依赖于具体的应用场景。
	// 例如，如果下游服务不可用，您可以：
	// - 返回缓存数据。
	// - 返回默认响应。
	// - 重定向到静态错误页面。
	// - 使用功能的简化版本。
	// 这通常涉及检查依赖项的健康/状态
	// 或在您的处理程序中实现回退逻辑。
	// --------------------------------------------------

	httpEndpoint := fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    httpEndpoint,
		Handler: r,
	}

	zap.S().Infof("Gin HTTP server listening at %s", httpEndpoint)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve Gin HTTP: %w", err)
		}
		close(errChan)
	}()
	return server, errChan
}
