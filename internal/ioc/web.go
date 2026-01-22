package ioc

import (
	"webook/config"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewGinEngine 创建并配置 Gin 引擎
func NewGinEngine(cfg *config.Config, userHandler *web.UserHandler, postHandler *web.PostHandler, l logger.Logger) *gin.Engine {
	server := gin.Default()

	// 初始化 JWT
	middleware.InitJWT(cfg.JWT.SecretKey)

	// 请求日志中间件（最先添加，记录所有请求）
	server.Use(middleware.NewRequestLoggerBuilder(l).
		IgnorePath("/health"). // 健康检查不记录
		Build())

	// CORS 中间件配置
	server.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// 开发环境允许 localhost
			return true
		},
		MaxAge: cfg.CORS.MaxAge,
	}))

	// JWT 登录校验中间件 - RESTful 路径白名单
	server.Use(middleware.NewJWTMiddlewareBuilder().
		IgnorePaths("/users", "/users/login", "/auth/refresh", "/auth/logout"). // 注册、登录、刷新、退出
		Build())

	// 注册路由
	userHandler.RegisterRoutes(server)
	postHandler.RegisterRoutes(server)

	return server
}
