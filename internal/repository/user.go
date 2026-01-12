package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
}

// userRepository 用户仓储实现
type userRepository struct {
	dao *dao.UserDAO
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(dao *dao.UserDAO) UserRepository {
	return &userRepository{
		dao: dao,
	}
}

func (r *userRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *userRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *userRepository) Update(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Password: u.Password,
	})
}

// toDomain 将 DAO 实体转换为领域对象
func (r *userRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
