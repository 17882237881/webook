package web

import (
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/service"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 处理用户相关的请求
type UserHandler struct {
	svc               service.UserService  // 依赖接口
	blacklist         cache.TokenBlacklist // Token 黑名单
	emailExp          *regexp.Regexp
	passwordExp       *regexp.Regexp
	jwtExpireTime     time.Duration // Access Token 有效期
	refreshExpireTime time.Duration // Refresh Token 有效期
}

// JWTExpireTime Access Token 过期时间类型（用于 Wire 依赖注入）
type JWTExpireTime time.Duration

// RefreshExpireTime Refresh Token 过期时间类型（用于 Wire 依赖注入）
type RefreshExpireTime time.Duration

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(svc service.UserService, blacklist cache.TokenBlacklist, jwtExpireTime JWTExpireTime, refreshExpireTime RefreshExpireTime) *UserHandler {
	const (
		emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegex = `^.{6,16}$`
	)
	return &UserHandler{
		svc:               svc,
		blacklist:         blacklist,
		emailExp:          regexp.MustCompile(emailRegex, regexp.None),
		passwordExp:       regexp.MustCompile(passwordRegex, regexp.None),
		jwtExpireTime:     time.Duration(jwtExpireTime),
		refreshExpireTime: time.Duration(refreshExpireTime),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	// RESTful API
	ug.POST("", u.SignUp)                   // 创建用户（注册）
	ug.POST("/login", u.Login)              // 登录（action 风格）
	ug.GET("/:id", u.Profile)               // 获取用户信息
	ug.PUT("/:id/password", u.EditPassword) // 修改密码

	// 认证相关
	server.POST("/auth/refresh", u.RefreshToken) // 刷新 Token
	server.POST("/auth/logout", u.Logout)        // 退出登录
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

	// 登录成功，生成 Token 对（Access Token + Refresh Token）
	// 生成唯一的 Session ID 用于退出登录时加入黑名单
	ssid := uuid.New().String()
	accessToken, refreshToken, err := middleware.GenerateTokenPair(
		user.Id,
		c.GetHeader("User-Agent"),
		ssid,
		u.jwtExpireTime,
		u.refreshExpireTime,
	)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "生成Token失败")
		return
	}

	ginx.Success(c, gin.H{
		"userId":       user.Id,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// RefreshToken 刷新 Access Token
// POST /auth/refresh
// 使用 Refresh Token 获取新的 Access Token
func (u *UserHandler) RefreshToken(c *gin.Context) {
	type RefreshReq struct {
		RefreshToken string `json:"refreshToken"`
	}

	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	// 解析并验证 Refresh Token
	claims, err := middleware.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		ginx.Error(c, ginx.CodeUnauthorized, "Refresh Token 无效或已过期")
		return
	}

	// 检查黑名单
	isBlacklisted, err := u.blacklist.IsBlacklisted(c.Request.Context(), claims.SSid)
	if err != nil || isBlacklisted {
		ginx.Error(c, ginx.CodeUnauthorized, "Token 已失效，请重新登录")
		return
	}

	// 生成新的 Access Token
	accessToken, err := middleware.GenerateAccessToken(
		claims.UserId,
		c.GetHeader("User-Agent"),
		u.jwtExpireTime,
	)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "生成Token失败")
		return
	}

	ginx.Success(c, gin.H{
		"accessToken": accessToken,
	})
}

// Logout 退出登录
// POST /auth/logout
// 将 Refresh Token 加入黑名单
func (u *UserHandler) Logout(c *gin.Context) {
	type LogoutReq struct {
		RefreshToken string `json:"refreshToken"`
	}

	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "参数错误")
		return
	}

	// 解析 Refresh Token 获取 SSid
	claims, err := middleware.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		// Token 无效也算退出成功
		ginx.SuccessMsg(c, "退出成功")
		return
	}

	// 将 SSid 加入黑名单，过期时间 = Refresh Token 剩余有效期
	_ = u.blacklist.Add(c.Request.Context(), claims.SSid, u.refreshExpireTime)

	ginx.SuccessMsg(c, "退出成功")
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
	id, err := strconv.ParseInt(idStr, 10, 64) // idStr 转换为 int64
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
