package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey []byte

// InitJWT 初始化 JWT 密钥
func InitJWT(secretKey string) {
	jwtSecretKey = []byte(secretKey)
}

// UserClaims 自定义 JWT Claims
type UserClaims struct {
	UserId int64 `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userId int64, expireTime time.Duration) (string, error) {
	claims := UserClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// JWTMiddlewareBuilder JWT 认证中间件构建器
type JWTMiddlewareBuilder struct {
	paths []string // 不需要登录校验的路径
}

func NewJWTMiddlewareBuilder() *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{}
}

// IgnorePaths 设置不需要登录校验的路径
func (j *JWTMiddlewareBuilder) IgnorePaths(paths ...string) *JWTMiddlewareBuilder {
	j.paths = append(j.paths, paths...)
	return j
}

func (j *JWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查当前路径是否在白名单中
		for _, path := range j.paths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Token 格式: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 解析并验证 Token
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 将 userId 存入 Context
		c.Set("userId", claims.UserId)
		c.Next()
	}
}
