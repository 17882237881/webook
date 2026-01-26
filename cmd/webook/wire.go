//go:build wireinject

package main

import (
	"time"
	"webook/config"
	web "webook/internal/adapters/inbound/http"
	mq "webook/internal/adapters/outbound/mq"
	dao "webook/internal/adapters/outbound/persistence/mysql"
	cache "webook/internal/adapters/outbound/persistence/redis"
	"webook/internal/adapters/outbound/repository"
	"webook/internal/application"
	"webook/internal/ioc"

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
		ioc.NewRabbitMQConn,
		ioc.NewRabbitMQProducerChannel,
		ioc.NewPostStatsPublisher,

		dao.NewUserDAO,
		dao.NewPostDAO,
		dao.NewPublishedPostDAO,
		dao.NewPostStatsDAO,
		dao.NewPostLikeDAO,
		dao.NewPostCollectDAO,

		ProvideUserCacheExpiration,
		cache.NewUserCache,
		cache.NewTokenBlacklist,
		cache.NewPostCache,
		cache.NewPostStatsCache,

		repository.NewUserRepository,
		repository.NewCachedUserRepository,
		repository.NewPostRepository,
		repository.NewPublishedPostRepository,
		repository.NewCachedPublishedPostRepository,
		repository.NewPostStatsRepository,
		repository.NewPostLikeRepository,
		repository.NewPostCollectRepository,

		application.NewUserService,
		application.NewPostService,
		application.NewPostInteractionService,
		ProvideAccessExpireTime,
		ProvideRefreshExpireTime,
		application.NewAuthService,

		web.NewUserHandler,
		web.NewPostHandler,
		ioc.NewGinEngine,
	)
	return nil
}

func InitPostStatsWorker(cfg *config.Config) *application.PostStatsWorker {
	wire.Build(
		ioc.NewDB,
		ioc.NewRedis,
		ioc.NewLogger,
		ioc.NewRabbitMQConn,
		ioc.NewRabbitMQConsumerChannel,
		ioc.NewPostStatsConsumer,

		dao.NewPostStatsDAO,
		cache.NewPostStatsCache,
		repository.NewPostStatsRepository,

		application.NewPostStatsFlusher,
		application.NewPostStatsWorker,

		wire.Bind(new(application.RabbitMQStatsConsumerWrapper), new(*mq.RabbitMQStatsConsumer)),
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
