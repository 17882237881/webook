package middleware

import (
	"net/http"
	"strings"
	ports "webook/internal/ports/output"

	"github.com/gin-gonic/gin"
)

type JWTMiddlewareBuilder struct {
	verifier    ports.AccessTokenVerifier
	ignorePaths map[string]struct{}
}

func NewJWTMiddlewareBuilder(verifier ports.AccessTokenVerifier) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{
		verifier:    verifier,
		ignorePaths: make(map[string]struct{}),
	}
}

func (b *JWTMiddlewareBuilder) IgnorePaths(paths ...string) *JWTMiddlewareBuilder {
	for _, path := range paths {
		b.ignorePaths[path] = struct{}{}
	}
	return b
}

func (b *JWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if _, ok := b.ignorePaths[path]; ok {
			ctx.Next()
			return
		}

		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(authCode, " ")
		if len(segs) != 2 || segs[0] != "Bearer" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := segs[1]
		userId, err := b.verifier.Verify(tokenStr, ctx.Request.UserAgent())
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("userId", userId)
		ctx.Next()
	}
}
