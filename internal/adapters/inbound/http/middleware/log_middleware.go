package middleware

import (
	"bytes"
	"io"
	"time"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
)

type RequestLoggerBuilder struct {
	l           logger.Logger
	ignorePaths map[string]struct{}
}

func NewRequestLoggerBuilder(l logger.Logger) *RequestLoggerBuilder {
	return &RequestLoggerBuilder{
		l:           l,
		ignorePaths: make(map[string]struct{}),
	}
}

func (b *RequestLoggerBuilder) IgnorePath(path string) *RequestLoggerBuilder {
	b.ignorePaths[path] = struct{}{}
	return b
}

func (b *RequestLoggerBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		if _, ok := b.ignorePaths[path]; ok {
			ctx.Next()
			return
		}

		var body []byte
		if ctx.Request.Body != nil {
			body, _ = ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		ctx.Next()

		duration := time.Since(start)
		b.l.Info("HTTP Request",
			logger.String("method", ctx.Request.Method),
			logger.String("path", path),
			logger.Int("status", ctx.Writer.Status()),
			logger.Duration("duration", duration),
			// Truncate body if too long? For now keep it simple.
			logger.String("body", string(body)),
		)
	}
}
