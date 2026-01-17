//go:build wireinject

package main

import (
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// InitWebServer 初始化 Web 服务器，使用 Wire 进行依赖注入
func InitWebServer() *gin.Engine {
	wire.Build(
		// 配置
		config.Load,

		// 基础设施层
		ioc.NewDB,
		ioc.NewRedis,

		// DAO 层
		dao.NewUserDAO,

		// Cache 层
		ProvideUserCacheExpiration,
		cache.NewUserCache,

		// Repository 层
		repository.NewUserRepository,

		// Service 层
		service.NewUserService,

		// Handler 层
		ProvideJWTExpireTime,
		web.NewUserHandler,

		// Web 层
		ioc.NewGinEngine,
	)
	return nil
}

// ProvideUserCacheExpiration 提供用户缓存过期时间
func ProvideUserCacheExpiration(cfg *config.Config) cache.UserCacheExpiration {
	return cache.UserCacheExpiration(cfg.Cache.UserExpiration)
}

// ProvideJWTExpireTime 提供 JWT 过期时间
func ProvideJWTExpireTime(cfg *config.Config) web.JWTExpireTime {
	return web.JWTExpireTime(cfg.JWT.ExpireTime)
}
