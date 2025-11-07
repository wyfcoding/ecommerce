package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"ecommerce/internal/report/service"
	"ecommerce/pkg/response"
)

type ReportHandler struct {
	service service.ReportService
	logger  *zap.Logger
}

func NewReportHandler(service service.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{service: service, logger: logger}
}

func (h *ReportHandler) RegisterRoutes(r *gin.RouterGroup) {
	report := r.Group("/reports")
	{
		report.GET("/health", func(c *gin.Context) { 
			response.Success(c, gin.H{"status": "ok"}) 
		})
		// TODO: 添加具体路由
	}
}
