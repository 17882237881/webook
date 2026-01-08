package main

import (
	"time"
	"webook/internal/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	// middleware
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "authorization"},
		AllowCredentials: true, // 允许发送 cookie
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://localhost:3000"
		}, 
		MaxAge: 12 * time.Hour,
	}))


	u := web.NewUserHandler()
	u.RegisterRoutes(server) 
	server.Run(":8080")
} 
