package web

import (
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关的请求
type UserHandler struct {
	svc           service.UserService // 依赖接口
	emailExp      *regexp.Regexp
	passwordExp   *regexp.Regexp
	jwtExpireTime time.Duration
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(svc service.UserService, jwtExpireTime time.Duration) *UserHandler {
	const (
		emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegex = `^.{6,16}$`
	)
	return &UserHandler{
		svc:           svc,
		emailExp:      regexp.MustCompile(emailRegex, regexp.None),
		passwordExp:   regexp.MustCompile(passwordRegex, regexp.None),
		jwtExpireTime: jwtExpireTime,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	// RESTful API
	ug.POST("", u.SignUp)                   // 创建用户（注册）
	ug.POST("/login", u.Login)              // 登录（action 风格）
	ug.GET("/:id", u.Profile)               // 获取用户信息
	ug.PUT("/:id/password", u.EditPassword) // 修改密码
}

// SignUp 用户注册
// POST /users
func (u *UserHandler) SignUp(c *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "系统错误")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "邮箱格式不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		ginx.Error(c, ginx.CodeInvalidParams, "两次密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "系统错误")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "密码长度应为6-16位")
		return
	}

	err = u.svc.SignUp(c.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrDuplicateEmail {
		ginx.Error(c, ginx.CodeDuplicateEmail, "邮箱已被注册")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "注册失败")
		return
	}

	ginx.SuccessMsg(c, "注册成功")
}

// Login 用户登录
// POST /users/login
func (u *UserHandler) Login(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	user, err := u.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ginx.Error(c, ginx.CodeUnauthorized, "邮箱或密码不正确")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "系统错误")
		return
	}

	// 登录成功，生成 JWT Token
	token, err := middleware.GenerateToken(user.Id, u.jwtExpireTime)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "生成Token失败")
		return
	}

	ginx.Success(c, gin.H{
		"userId": user.Id,
		"token":  token,
	})
}

// Profile 获取用户信息
// GET /users/:id
func (u *UserHandler) Profile(c *gin.Context) {
	// 从 URL 参数获取 id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的用户ID")
		return
	}

	// 从 Context 获取当前登录用户（中间件已设置）
	currentUserId := c.GetInt64("userId")
	if currentUserId != id {
		ginx.ErrorWithStatus(c, http.StatusForbidden, ginx.CodeForbidden, "无权访问")
		return
	}

	user, err := u.svc.Profile(c.Request.Context(), id)
	if err != nil {
		ginx.Error(c, ginx.CodeNotFound, "用户不存在")
		return
	}

	ginx.Success(c, gin.H{
		"id":    user.Id,
		"email": user.Email,
	})
}

// EditPassword 修改密码
// PUT /users/:id/password
func (u *UserHandler) EditPassword(c *gin.Context) {
	type EditPasswordReq struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "无效的用户ID")
		return
	}

	currentUserId := c.GetInt64("userId")
	if currentUserId != id {
		ginx.ErrorWithStatus(c, http.StatusForbidden, ginx.CodeForbidden, "无权操作")
		return
	}

	var req EditPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	ok, err := u.passwordExp.MatchString(req.NewPassword)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "系统错误")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "新密码长度应为6-16位")
		return
	}

	err = u.svc.UpdatePassword(c.Request.Context(), id, req.OldPassword, req.NewPassword)
	if err == service.ErrInvalidUserOrPassword {
		ginx.Error(c, ginx.CodeUnauthorized, "原密码不正确")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "修改失败")
		return
	}

	ginx.SuccessMsg(c, "密码修改成功")
}
