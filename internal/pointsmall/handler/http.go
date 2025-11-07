package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"ecommerce/internal/pointsmall/service"
	"ecommerce/pkg/response"
)

type PointsMallHandler struct {
	service service.PointsMallService
	logger  *zap.Logger
}

func NewPointsMallHandler(service service.PointsMallService, logger *zap.Logger) *PointsMallHandler {
	return &PointsMallHandler{service: service, logger: logger}
}

func (h *PointsMallHandler) RegisterRoutes(r *gin.RouterGroup) {
	pm := r.Group("/points-mall")
	{
		pm.GET("/health", func(c *gin.Context) { 
			response.Success(c, gin.H{"status": "ok"}) 
		})
		// TODO: 添加具体路由
	}
}
