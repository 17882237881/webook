package web

import (
	"net/http"
	"webook/internal/domain"
	"webook/internal/service"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关的请求
// 将预编译的正则表达式存储在结构体字段中，而不是在每次请求时编译
// 好处：正则表达式只在 NewUserHandler() 时编译一次，后续所有请求都复用已编译的对象
// 避免了每次请求都重新编译正则表达式的性能开销
type UserHandler struct {
	svc         *service.UserService // 业务逻辑层
	emailExp    *regexp.Regexp       // 预编译的邮箱格式正则表达式
	passwordExp *regexp.Regexp       // 预编译的密码格式正则表达式
}

// NewUserHandler 创建 UserHandler 实例
// 在这里完成正则表达式的编译，确保只编译一次
func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegex = `^.{6,16}$`
	)
	return &UserHandler{
		svc:         svc,
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

	err = u.svc.SignUp(c.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.String(http.StatusOK, "注册失败")
		return
	}

	c.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		c.String(http.StatusOK, "邮箱或密码不正确")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "系统错误")
		return
	}

	c.String(http.StatusOK, "登录成功")
	_ = user // 后续可用于设置 session
}

func (u *UserHandler) Edit(c *gin.Context) {

}

func (u *UserHandler) Profile(c *gin.Context) {

}
