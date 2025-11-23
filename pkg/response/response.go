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

// SuccessWithStatus 带状态码的成功响应
func SuccessWithStatus(c *gin.Context, status int, msg string, data interface{}) {
	c.JSON(status, gin.H{
		"code": 0,
		"msg":  msg,
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

// ErrorWithStatus 带状态码的错误响应
func ErrorWithStatus(c *gin.Context, status int, msg string, detail string) {
	c.JSON(status, gin.H{
		"code":   status,
		"msg":    msg,
		"detail": detail,
	})
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"code": code,
		"msg":  msg,
	})
}

// SuccessWithPagination 分页成功响应
func SuccessWithPagination(c *gin.Context, data interface{}, total int64, page, size int32) {
	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "success",
		"data":  data,
		"total": total,
		"page":  page,
		"size":  size,
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
