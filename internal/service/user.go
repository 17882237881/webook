package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/repository/dao"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserOrPassword = errors.New("邮箱或密码不正确")
	ErrDuplicateEmail        = dao.ErrDuplicateEmail
)

// UserService 用户服务接口
type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	UpdatePassword(ctx context.Context, id int64, oldPwd, newPwd string) error
}

// userService 用户服务实现
type userService struct {
	repo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 使用 bcrypt 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 使用 bcrypt 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *userService) UpdatePassword(ctx context.Context, id int64, oldPwd, newPwd string) error {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return err
	}
	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPwd))
	if err != nil {
		return ErrInvalidUserOrPassword
	}
	// 加密新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Update(ctx, u)
}
