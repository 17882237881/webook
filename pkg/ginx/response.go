package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"`           // 业务状态码：0 成功，非0 失败
	Msg  string `json:"msg"`            // 提示信息
	Data any    `json:"data,omitempty"` // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// SuccessMsg 成功响应（仅消息）
func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  msg,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// ErrorWithStatus 带 HTTP 状态码的错误响应
func ErrorWithStatus(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
	})
}

// 常用业务错误码
const (
	CodeSuccess        = 0
	CodeInvalidParams  = 400001
	CodeUnauthorized   = 401001
	CodeForbidden      = 403001
	CodeNotFound       = 404001
	CodeDuplicateEmail = 409001
	CodeInternalError  = 500001
)
