package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	"webook/internal/ports"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	accessKey  []byte
	refreshKey []byte
}

func NewJWTService(accessKey string, refreshKey string) *JWTService {
	if refreshKey == "" {
		refreshKey = accessKey + "_refresh"
	}
	return &JWTService{
		accessKey:  []byte(accessKey),
		refreshKey: []byte(refreshKey),
	}
}

type UserClaims struct {
	UserId    int64  `json:"userId"`
	UserAgent string `json:"userAgent"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserId int64  `json:"userId"`
	SSid   string `json:"ssid"`
	jwt.RegisteredClaims
}

func hashUserAgent(userAgent string) string {
	hash := sha256.Sum256([]byte(userAgent))
	return hex.EncodeToString(hash[:])
}

func (s *JWTService) GenerateTokenPair(userId int64, userAgent, ssid string, accessExpire, refreshExpire time.Duration) (accessToken, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessToken(userId, userAgent, accessExpire)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = s.GenerateRefreshToken(userId, ssid, refreshExpire)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *JWTService) GenerateAccessToken(userId int64, userAgent string, expireTime time.Duration) (string, error) {
	claims := UserClaims{
		UserId:    userId,
		UserAgent: hashUserAgent(userAgent),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessKey)
}

func (s *JWTService) GenerateRefreshToken(userId int64, ssid string, expireTime time.Duration) (string, error) {
	claims := RefreshClaims{
		UserId: userId,
		SSid:   ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshKey)
}

func (s *JWTService) ParseRefreshToken(tokenString string) (ports.RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return s.refreshKey, nil
	})
	if err != nil || !token.Valid {
		return ports.RefreshClaims{}, err
	}
	return ports.RefreshClaims{
		UserId: claims.UserId,
		SSid:   claims.SSid,
	}, nil
}

func (s *JWTService) Verify(tokenString, userAgent string) (int64, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return s.accessKey, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	if claims.UserAgent != "" && claims.UserAgent != hashUserAgent(userAgent) {
		return 0, errors.New("user agent mismatch")
	}
	return claims.UserId, nil
}
