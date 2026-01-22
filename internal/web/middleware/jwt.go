package middleware

import (
	"net/http"
	"strings"
	"webook/internal/ports"

	"github.com/gin-gonic/gin"
)

// JWTMiddlewareBuilder builds an auth middleware using a token verifier.
type JWTMiddlewareBuilder struct {
	paths    []string
	verifier ports.AccessTokenVerifier
}

func NewJWTMiddlewareBuilder(verifier ports.AccessTokenVerifier) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{verifier: verifier}
}

func (j *JWTMiddlewareBuilder) IgnorePaths(paths ...string) *JWTMiddlewareBuilder {
	j.paths = append(j.paths, paths...)
	return j
}

func (j *JWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, path := range j.paths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userId, err := j.verifier.Verify(parts[1], c.GetHeader("User-Agent"))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userId", userId)
		c.Next()
	}
}
