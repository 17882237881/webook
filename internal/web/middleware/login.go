package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string // 不需要登录校验的路径
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

// IgnorePaths 设置不需要登录校验的路径
func (l *LoginMiddlewareBuilder) IgnorePaths(paths ...string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, paths...)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查当前路径是否在白名单中
		for _, path := range l.paths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 从 session 获取用户 ID
		sess := sessions.Default(c)
		userId := sess.Get("userId")
		if userId == nil {
			// 未登录，返回 401
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 将 userId 存入 Context，供后续 Handler 使用
		// session 存储的是 int64，需要类型断言
		c.Set("userId", userId.(int64))

		// 已登录，继续处理请求
		c.Next()
	}
}
