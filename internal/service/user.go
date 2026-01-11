package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"
)

var ErrInvalidUserOrPassword = errors.New("邮箱或密码不正确")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	return svc.repo.Creat(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if u.Password != password {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}
