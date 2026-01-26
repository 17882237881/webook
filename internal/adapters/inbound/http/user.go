package web

import (
	"net/http"
	"strconv"
	"webook/internal/domain"
	service "webook/internal/ports/input"
	"webook/internal/adapters/inbound/http/ginx"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user APIs.
type UserHandler struct {
	svc         service.UserService
	auth        service.AuthService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService, auth service.AuthService) *UserHandler {
	const (
		emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegex = `^.{6,16}$`
	)
	return &UserHandler{
		svc:         svc,
		auth:        auth,
		emailExp:    regexp.MustCompile(emailRegex, regexp.None),
		passwordExp: regexp.MustCompile(passwordRegex, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("", u.SignUp)
	ug.POST("/login", u.Login)
	ug.GET("/:id", u.Profile)
	ug.PUT("/:id/password", u.EditPassword)

	server.POST("/auth/refresh", u.RefreshToken)
	server.POST("/auth/logout", u.Logout)
}

// POST /users
func (u *UserHandler) SignUp(c *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid params")
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "regex error")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid email")
		return
	}

	if req.Password != req.ConfirmPassword {
		ginx.Error(c, ginx.CodeInvalidParams, "password mismatch")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "regex error")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid password")
		return
	}

	err = u.svc.SignUp(c.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == domain.ErrDuplicateEmail {
		ginx.Error(c, ginx.CodeDuplicateEmail, "duplicate email")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "signup failed")
		return
	}

	ginx.SuccessMsg(c, "signup success")
}

// POST /users/login
func (u *UserHandler) Login(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid params")
		return
	}

	user, err := u.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err == domain.ErrInvalidUserOrPassword {
		ginx.Error(c, ginx.CodeUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "login failed")
		return
	}

	accessToken, refreshToken, err := u.auth.GenerateTokenPair(c.Request.Context(), user.Id, c.GetHeader("User-Agent"))
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "token generate failed")
		return
	}

	ginx.Success(c, gin.H{
		"userId":       user.Id,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// POST /auth/refresh
func (u *UserHandler) RefreshToken(c *gin.Context) {
	type RefreshReq struct {
		RefreshToken string `json:"refreshToken"`
	}

	var req RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid params")
		return
	}

	accessToken, err := u.auth.RefreshAccessToken(c.Request.Context(), req.RefreshToken, c.GetHeader("User-Agent"))
	if err == domain.ErrUnauthorized {
		ginx.Error(c, ginx.CodeUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "refresh failed")
		return
	}

	ginx.Success(c, gin.H{
		"accessToken": accessToken,
	})
}

// POST /auth/logout
func (u *UserHandler) Logout(c *gin.Context) {
	type LogoutReq struct {
		RefreshToken string `json:"refreshToken"`
	}

	var req LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid params")
		return
	}

	if err := u.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		ginx.Error(c, ginx.CodeInternalError, "logout failed")
		return
	}

	ginx.SuccessMsg(c, "logout success")
}

// GET /users/:id
func (u *UserHandler) Profile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid id")
		return
	}

	currentUserId := c.GetInt64("userId")
	if currentUserId != id {
		ginx.ErrorWithStatus(c, http.StatusForbidden, ginx.CodeForbidden, "forbidden")
		return
	}

	user, err := u.svc.Profile(c.Request.Context(), id)
	if err != nil {
		ginx.Error(c, ginx.CodeNotFound, "user not found")
		return
	}

	ginx.Success(c, gin.H{
		"id":    user.Id,
		"email": user.Email,
	})
}

// PUT /users/:id/password
func (u *UserHandler) EditPassword(c *gin.Context) {
	type EditPasswordReq struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid id")
		return
	}

	currentUserId := c.GetInt64("userId")
	if currentUserId != id {
		ginx.ErrorWithStatus(c, http.StatusForbidden, ginx.CodeForbidden, "forbidden")
		return
	}

	var req EditPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid params")
		return
	}

	ok, err := u.passwordExp.MatchString(req.NewPassword)
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "regex error")
		return
	}
	if !ok {
		ginx.Error(c, ginx.CodeInvalidParams, "invalid password")
		return
	}

	err = u.svc.UpdatePassword(c.Request.Context(), id, req.OldPassword, req.NewPassword)
	if err == domain.ErrInvalidUserOrPassword {
		ginx.Error(c, ginx.CodeUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		ginx.Error(c, ginx.CodeInternalError, "update failed")
		return
	}

	ginx.SuccessMsg(c, "update success")
}
