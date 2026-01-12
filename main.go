package main

import (
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db := initDB(cfg)

	// 依赖注入：DAO → Repository → Service → Handler
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO)
	userSvc := service.NewUserService(userRepo)
	u := web.NewUserHandler(userSvc)

	server := gin.Default()

	// Session 配置
	store := cookie.NewStore([]byte(cfg.Session.Secret))
	server.Use(sessions.Sessions(cfg.Session.Name, store))

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

	// 登录校验中间件 - RESTful 路径白名单
	server.Use(middleware.NewLoginMiddlewareBuilder().
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
