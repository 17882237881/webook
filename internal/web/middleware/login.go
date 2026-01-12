package middleware

import (
	"net/http"
	"time"

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
				c.Next() // 白名单路径，直接放行（继续执行下一个中间件）
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

		// Session 续期逻辑（滑动过期）
		// 如果距离上次更新超过 10 分钟，就刷新 Session
		const refreshInterval = 10 * 60 * 1000 // 10分钟（毫秒）
		now := time.Now().UnixMilli()
		updateTime := sess.Get("updateTime")

		if updateTime == nil || now-updateTime.(int64) > refreshInterval {
			sess.Set("updateTime", now)
			sess.Options(sessions.Options{
				MaxAge: 30 * 60, // 续期30分钟
			})
			sess.Save()
		}

		// 将 userId 存入 Context，供后续 Handler 使用
		// session 存储的是 int64，需要类型断言
		c.Set("userId", userId.(int64))

		// 已登录，继续下一个中间件
		c.Next()
	}
}
