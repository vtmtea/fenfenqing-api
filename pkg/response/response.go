package response

import (
	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(200, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context) {
	Error(c, 404, "resource not found")
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context) {
	Error(c, 401, "unauthorized")
}
