package main

import (
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db := initDB(cfg)

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})

	// 初始化 JWT
	middleware.InitJWT(cfg.JWT.SecretKey)

	// 依赖注入：DAO → Cache → Repository → Service → Handler
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(rdb, cfg.Cache.UserExpiration)
	userRepo := repository.NewUserRepository(userDAO, userCache)
	userSvc := service.NewUserService(userRepo)
	u := web.NewUserHandler(userSvc, cfg.JWT.ExpireTime)

	server := gin.Default()

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
		IgnorePaths("/users", "/users/login"). // POST /users 注册, POST /users/login 登录
		Build())

	u.RegisterRoutes(server)
	server.Run(cfg.Server.Port)
}

func initDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 自动迁移数据库表结构
	err = db.AutoMigrate(&dao.User{})
	if err != nil {
		panic(err)
	}
	return db
}
