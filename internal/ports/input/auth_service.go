package input

import "context"

// AuthService 认证业务接口
type AuthService interface {
	GenerateTokenPair(ctx context.Context, uid int64, userAgent string) (string, string, error)
	RefreshAccessToken(ctx context.Context, refreshToken, userAgent string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
}
