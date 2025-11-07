package handler

import (
	"ecommerce/internal/user/service"
	"ecommerce/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, svc *service.UserService) {
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "healthy"})
	})
}
