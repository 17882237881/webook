package middleware

import (
	"time"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLoggerBuilder HTTP 请求日志中间件构建器
type RequestLoggerBuilder struct {
	logger       logger.Logger
	allowedPaths map[string]bool // 不记录的路径
}

// NewRequestLoggerBuilder 创建请求日志中间件构建器
func NewRequestLoggerBuilder(l logger.Logger) *RequestLoggerBuilder {
	return &RequestLoggerBuilder{
		logger:       l,
		allowedPaths: make(map[string]bool),
	}
}

// IgnorePath 忽略指定路径的日志
func (b *RequestLoggerBuilder) IgnorePath(path string) *RequestLoggerBuilder {
	b.allowedPaths[path] = true
	return b
}

// Build 构建中间件
func (b *RequestLoggerBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否忽略此路径
		if b.allowedPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// 生成请求 ID
		traceId := uuid.New().String()
		c.Set("traceId", traceId)

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []logger.Field{
			logger.String("traceId", traceId),
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.Float64("latency_ms", float64(latency.Milliseconds())),
			logger.String("ip", clientIP),
			logger.String("userAgent", userAgent),
		}

		if query != "" {
			fields = append(fields, logger.String("query", query))
		}

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			b.logger.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			b.logger.Warn("HTTP Request", fields...)
		} else {
			b.logger.Info("HTTP Request", fields...)
		}
	}
}
