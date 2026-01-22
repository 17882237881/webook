package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/ports"

	"github.com/google/uuid"
)

// AuthService encapsulates token and session behaviors.
type AuthService interface {
	GenerateTokenPair(ctx context.Context, userId int64, userAgent string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken, userAgent string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	tokens        ports.TokenService
	blacklist     ports.TokenBlacklist
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewAuthService(tokens ports.TokenService, blacklist ports.TokenBlacklist, accessExpire, refreshExpire time.Duration) AuthService {
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
