package input

import (
	"context"
	"webook/internal/domain"
)

// UserService 用户业务接口
type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	UpdatePassword(ctx context.Context, id int64, old, new string) error
}
