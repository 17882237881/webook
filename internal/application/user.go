package application

import (
	"context"
	"webook/internal/domain"
	input "webook/internal/ports/input"
	output "webook/internal/ports/output"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo output.UserRepository
}

func NewUserService(repo output.UserRepository) input.UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
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
		return domain.User{}, domain.ErrInvalidUserOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, domain.ErrInvalidUserOrPassword
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

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPwd))
	if err != nil {
		return domain.ErrInvalidUserOrPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Update(ctx, u)
}
