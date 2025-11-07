package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"ecommerce/internal/scheduler/service"
	"ecommerce/pkg/response"
)

type SchedulerHandler struct {
	service service.SchedulerService
	logger  *zap.Logger
}

func NewSchedulerHandler(service service.SchedulerService, logger *zap.Logger) *SchedulerHandler {
	return &SchedulerHandler{service: service, logger: logger}
}

func (h *SchedulerHandler) RegisterRoutes(r *gin.RouterGroup) {
	scheduler := r.Group("/scheduler")
	{
		scheduler.GET("/health", func(c *gin.Context) { 
			response.Success(c, gin.H{"status": "ok"}) 
		})
		// TODO: 添加具体路由
	}
}
