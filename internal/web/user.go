package web

import "github.com/gin-gonic/gin"

// UserHandler 处理用户相关的请求
type UserHandler struct {
	

}



func(u *UserHandler) RegisterRoutes(server *gin.Engine) {

	ug := server.Group("/users")
	
	// 用户注册
	ug.POST("/signup",u.Signup)

	// 用户登录
	ug.POST("/login",u.Login)

	// 用户编辑
	ug.POST("/edit", u.Edit)

	// 用户详情
	ug.GET("/profile", u.Profile)
}


func(u *UserHandler) Signup(c *gin.Context) {
	
}

func(u *UserHandler) Login(c *gin.Context) {
	
}


func(u *UserHandler) Edit(c *gin.Context) {
	
}

func(u *UserHandler) Profile(c *gin.Context) {
	
	 
}
