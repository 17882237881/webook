package web

import (
	"net/http"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关的请求
type UserHandler struct {
	emailExp *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler() *UserHandler {
	const (
		emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegex = `^.{6,16}$`
	)
	return &UserHandler{
		emailExp:    regexp.MustCompile(emailRegex, regexp.None),
		passwordExp: regexp.MustCompile(passwordRegex, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {

	ug := server.Group("/users")

	// 用户注册
	ug.POST("/signup", u.Signup)

	// 用户登录
	ug.POST("/login", u.Login)

	// 用户编辑
	ug.POST("/edit", u.Edit)

	// 用户详情
	ug.GET("/profile", u.Profile)
}

func (u *UserHandler) Signup(c *gin.Context) {
	type SingUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SingUpReq
	if err := c.Bind(&req); err != nil { //Bind 方法会根据请求的 Content-Type 来解析你的数据到req里面
		return
	}
	
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "邮箱格式不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		c.String(http.StatusOK, "密码不一致")
		return
	}

	
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		c.String(http.StatusOK, "密码格式不正确")
		return
	}

	//省略数据库操作

	c.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(c *gin.Context) {

}

func (u *UserHandler) Edit(c *gin.Context) {

}

func (u *UserHandler) Profile(c *gin.Context) {

}
