package ports

import (
	"context"
	"time"
	"webook/internal/domain"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

type PostCache interface {
	Get(ctx context.Context, id int64) (domain.Post, error)
	Set(ctx context.Context, p domain.Post) error
	Delete(ctx context.Context, id int64) error
}

type TokenBlacklist interface {
	Add(ctx context.Context, ssid string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, ssid string) (bool, error)
}
