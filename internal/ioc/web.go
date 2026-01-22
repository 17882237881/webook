package ioc

import (
	"webook/config"
	"webook/internal/ports"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewGinEngine(cfg *config.Config, userHandler *web.UserHandler, postHandler *web.PostHandler, verifier ports.AccessTokenVerifier, l logger.Logger) *gin.Engine {
	server := gin.Default()

	server.Use(middleware.NewRequestLoggerBuilder(l).
		IgnorePath("/health").
		Build())

	server.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: cfg.CORS.MaxAge,
	}))

	server.Use(middleware.NewJWTMiddlewareBuilder(verifier).
		IgnorePaths("/users", "/users/login", "/auth/refresh", "/auth/logout").
		Build())

	userHandler.RegisterRoutes(server)
	postHandler.RegisterRoutes(server)

	return server
}
