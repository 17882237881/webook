package ioc

import (
	"webook/config"
	auth "webook/internal/adapters/outbound/jwt"
	ports "webook/internal/ports/output"
)

func NewJWTService(cfg *config.Config) *auth.JWTService {
	return auth.NewJWTService(cfg.JWT.SecretKey, "")
}

func NewTokenService(svc *auth.JWTService) ports.TokenService {
	return svc
}

func NewAccessTokenVerifier(svc *auth.JWTService) ports.AccessTokenVerifier {
	return svc
}
