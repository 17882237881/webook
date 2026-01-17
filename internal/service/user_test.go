package service

import (
	"context"
	"errors"
	"testing"
	"webook/internal/domain"
	repomocks "webook/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_SignUp(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		mock    func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantErr error
	}{
		{
			name: "注册成功",
			user: domain.User{
				Email:    "test@example.com",
				Password: "123456",
			},
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return repo
			},
			wantErr: nil,
		},
		{
			name: "邮箱重复",
			user: domain.User{
				Email:    "duplicate@example.com",
				Password: "123456",
			},
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(ErrDuplicateEmail)
				return repo
			},
			wantErr: ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			err := svc.SignUp(context.Background(), tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	// 预先生成一个加密密码用于测试
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		email    string
		password string
		mock     func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantUser domain.User
		wantErr  error
	}{
		{
			name:     "登录成功",
			email:    "test@example.com",
			password: "123456",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(hashedPassword),
					}, nil)
				return repo
			},
			wantUser: domain.User{
				Id:       1,
				Email:    "test@example.com",
				Password: string(hashedPassword),
			},
			wantErr: nil,
		},
		{
			name:     "用户不存在",
			email:    "notfound@example.com",
			password: "123456",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "notfound@example.com").
					Return(domain.User{}, errors.New("用户不存在"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name:     "密码错误",
			email:    "test@example.com",
			password: "wrongpassword",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "test@example.com").
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(hashedPassword),
					}, nil)
				return repo
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			user, err := svc.Login(context.Background(), tt.email, tt.password)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantUser.Id, user.Id)
				assert.Equal(t, tt.wantUser.Email, user.Email)
			}
		})
	}
}

func TestUserService_Profile(t *testing.T) {
	tests := []struct {
		name     string
		userId   int64
		mock     func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantUser domain.User
		wantErr  error
	}{
		{
			name:   "获取成功",
			userId: 1,
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:    1,
						Email: "test@example.com",
					}, nil)
				return repo
			},
			wantUser: domain.User{
				Id:    1,
				Email: "test@example.com",
			},
			wantErr: nil,
		},
		{
			name:   "用户不存在",
			userId: 999,
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(999)).
					Return(domain.User{}, errors.New("用户不存在"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  errors.New("用户不存在"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			user, err := svc.Profile(context.Background(), tt.userId)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser.Id, user.Id)
				assert.Equal(t, tt.wantUser.Email, user.Email)
			}
		})
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	// 预先生成加密密码
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)

	tests := []struct {
		name    string
		userId  int64
		oldPwd  string
		newPwd  string
		mock    func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantErr error
	}{
		{
			name:   "修改成功",
			userId: 1,
			oldPwd: "oldpass",
			newPwd: "newpass",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(oldHashedPassword),
					}, nil)
				repo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
				return repo
			},
			wantErr: nil,
		},
		{
			name:   "旧密码错误",
			userId: 1,
			oldPwd: "wrongoldpass",
			newPwd: "newpass",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(oldHashedPassword),
					}, nil)
				// 不应调用 Update
				return repo
			},
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name:   "用户不存在",
			userId: 999,
			oldPwd: "oldpass",
			newPwd: "newpass",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(999)).
					Return(domain.User{}, errors.New("用户不存在"))
				return repo
			},
			wantErr: errors.New("用户不存在"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			err := svc.UpdatePassword(context.Background(), tt.userId, tt.oldPwd, tt.newPwd)
			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.wantErr == ErrInvalidUserOrPassword {
					assert.Equal(t, ErrInvalidUserOrPassword, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
