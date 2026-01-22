//go:build wireinject

package main

import (
	"time"
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

func InitWebServer(cfg *config.Config) *gin.Engine {
	wire.Build(
		ioc.NewDB,
		ioc.NewRedis,
		ioc.NewLogger,
		ioc.NewJWTService,
		ioc.NewTokenService,
		ioc.NewAccessTokenVerifier,

		dao.NewUserDAO,
		dao.NewPostDAO,
		dao.NewPublishedPostDAO,

		ProvideUserCacheExpiration,
		cache.NewUserCache,
		cache.NewTokenBlacklist,
		cache.NewPostCache,

		repository.NewUserRepository,
		repository.NewCachedUserRepository,
		repository.NewPostRepository,
		repository.NewPublishedPostRepository,
		repository.NewCachedPublishedPostRepository,

		service.NewUserService,
		service.NewPostService,
		ProvideAccessExpireTime,
		ProvideRefreshExpireTime,
		service.NewAuthService,

		web.NewUserHandler,
		web.NewPostHandler,
		ioc.NewGinEngine,
	)
	return nil
}

func ProvideUserCacheExpiration(cfg *config.Config) cache.UserCacheExpiration {
	return cache.UserCacheExpiration(cfg.Cache.UserExpiration)
}

func ProvideAccessExpireTime(cfg *config.Config) time.Duration {
	return cfg.JWT.ExpireTime
}

func ProvideRefreshExpireTime(cfg *config.Config) time.Duration {
	return cfg.JWT.RefreshExpireTime
}
