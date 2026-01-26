package application

import (
	"context"
	"time"
	"webook/internal/domain"
	input "webook/internal/ports/input"
	output "webook/internal/ports/output"

	"github.com/google/uuid"
)

type authService struct {
	tokens        output.TokenService
	blacklist     output.TokenBlacklist
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewAuthService(tokens output.TokenService, blacklist output.TokenBlacklist, accessExpire, refreshExpire time.Duration) input.AuthService {
	return &authService{
		tokens:        tokens,
		blacklist:     blacklist,
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}
}

func (a *authService) GenerateTokenPair(ctx context.Context, userId int64, userAgent string) (string, string, error) {
	ssid := uuid.NewString()
	return a.tokens.GenerateTokenPair(userId, userAgent, ssid, a.accessExpire, a.refreshExpire)
}

func (a *authService) RefreshAccessToken(ctx context.Context, refreshToken, userAgent string) (string, error) {
	claims, err := a.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	isBlacklisted, err := a.blacklist.IsBlacklisted(ctx, claims.SSid)
	if err != nil || isBlacklisted {
		return "", domain.ErrUnauthorized
	}

	return a.tokens.GenerateAccessToken(claims.UserId, userAgent, a.accessExpire)
}

func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := a.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	return a.blacklist.Add(ctx, claims.SSid, a.refreshExpire)
}
