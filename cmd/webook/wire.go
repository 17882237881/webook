//go:build wireinject

package main

import (
	"webook/config"
	"webook/internal/ioc"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// InitWebServer 初始化 Web 服务器，使用 Wire 进行依赖注入
func InitWebServer(cfg *config.Config) *gin.Engine {
	wire.Build(
		// 基础设施层
		ioc.NewDB,
		ioc.NewRedis,

		// DAO 层
		dao.NewUserDAO,

		// Cache 层
		ProvideUserCacheExpiration,
		cache.NewUserCache,
		cache.NewTokenBlacklist,

		// Repository 层
		repository.NewUserRepository,

		// Service 层
		service.NewUserService,

		// Handler 层
		ProvideJWTExpireTime,
		ProvideRefreshExpireTime,
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

// ProvideJWTExpireTime 提供 Access Token 过期时间
func ProvideJWTExpireTime(cfg *config.Config) web.JWTExpireTime {
	return web.JWTExpireTime(cfg.JWT.ExpireTime)
}

// ProvideRefreshExpireTime 提供 Refresh Token 过期时间
func ProvideRefreshExpireTime(cfg *config.Config) web.RefreshExpireTime {
	return web.RefreshExpireTime(cfg.JWT.RefreshExpireTime)
}
