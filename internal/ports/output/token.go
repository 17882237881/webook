package output

import "time"

type RefreshClaims struct {
	UserId int64
	SSid   string
}

type TokenService interface {
	GenerateTokenPair(userId int64, userAgent, ssid string, accessExpire, refreshExpire time.Duration) (accessToken, refreshToken string, err error)
	GenerateAccessToken(userId int64, userAgent string, expireTime time.Duration) (string, error)
	ParseRefreshToken(tokenString string) (RefreshClaims, error)
}

type AccessTokenVerifier interface {
	Verify(tokenString, userAgent string) (userId int64, err error)
}
