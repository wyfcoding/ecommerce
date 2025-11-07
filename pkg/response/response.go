package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  msg,
		"data": data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": 500,
		"msg":  err.Error(),
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code": 400,
		"msg":  msg,
	})
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code": 401,
		"msg":  msg,
	})
}

// NotFound 404错误
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{
		"code": 404,
		"msg":  msg,
	})
}

// InternalError 500错误
func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": 500,
		"msg":  msg,
	})
}
