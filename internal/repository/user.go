package repository

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/ports"
	"webook/internal/repository/dao"

	"gorm.io/gorm"
)

// NewUserRepository builds a DAO-backed user repository.
func NewUserRepository(dao *dao.UserDAO) ports.UserRepository {
	return &userRepository{dao: dao}
}

// NewCachedUserRepository wraps a repository with cache behavior.
func NewCachedUserRepository(repo ports.UserRepository, cache ports.UserCache) ports.UserRepository {
	return &cachedUserRepository{repo: repo, cache: cache}
}

type userRepository struct {
	dao *dao.UserDAO
}

type cachedUserRepository struct {
	repo  ports.UserRepository
	cache ports.UserCache
}

func (r *userRepository) Create(ctx context.Context, u domain.User) error {
	err := r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
	if err == dao.ErrDuplicateEmail {
		return domain.ErrDuplicateEmail
	}
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(u), nil
}

func (r *userRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(u), nil
}

func (r *userRepository) Update(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Password: u.Password,
	})
}

func (r *cachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.repo.Create(ctx, u)
}

func (r *cachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.repo.FindByEmail(ctx, email)
}

func (r *cachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, nil
	}

	u, err = r.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	go func() {
		_ = r.cache.Set(ctx, u)
	}()
	return u, nil
}

func (r *cachedUserRepository) Update(ctx context.Context, u domain.User) error {
	err := r.repo.Update(ctx, u)
	if err != nil {
		return err
	}
	return r.cache.Delete(ctx, u.Id)
}

func toDomainUser(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
