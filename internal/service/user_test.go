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
			name: "signup success",
			user: domain.User{
				Email:    "test@example.com",
				Password: "password123",
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
			name: "signup failed - email exists",
			user: domain.User{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("email already exists"))
				return repo
			},
			wantErr: errors.New("email already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			err := svc.SignUp(context.Background(), tt.user)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		email    string
		password string
		mock     func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantUser domain.User
		wantErr  error
	}{
		{
			name:     "login success",
			email:    "test@example.com",
			password: "password123",
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
			name:     "login failed - user not found",
			email:    "notexist@example.com",
			password: "password123",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "notexist@example.com").
					Return(domain.User{}, errors.New("user not found"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  domain.ErrInvalidUserOrPassword,
		},
		{
			name:     "login failed - wrong password",
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
			wantErr:  domain.ErrInvalidUserOrPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tt.mock(ctrl)
			svc := NewUserService(repo)

			user, err := svc.Login(context.Background(), tt.email, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
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
			name:   "get profile success",
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
			name:   "user not found",
			userId: 999,
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(999)).
					Return(domain.User{}, errors.New("user not found"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  errors.New("user not found"),
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
				assert.Equal(t, tt.wantUser, user)
			}
		})
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name    string
		userId  int64
		oldPwd  string
		newPwd  string
		mock    func(ctrl *gomock.Controller) *repomocks.MockUserRepository
		wantErr error
	}{
		{
			name:   "update password success",
			userId: 1,
			oldPwd: "oldpassword",
			newPwd: "newpassword123",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(hashedPassword),
					}, nil)
				repo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
				return repo
			},
			wantErr: nil,
		},
		{
			name:   "update password failed - wrong old password",
			userId: 1,
			oldPwd: "wrongpassword",
			newPwd: "newpassword123",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.User{
						Id:       1,
						Email:    "test@example.com",
						Password: string(hashedPassword),
					}, nil)
				return repo
			},
			wantErr: domain.ErrInvalidUserOrPassword,
		},
		{
			name:   "update password failed - user not found",
			userId: 999,
			oldPwd: "oldpassword",
			newPwd: "newpassword123",
			mock: func(ctrl *gomock.Controller) *repomocks.MockUserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindById(gomock.Any(), int64(999)).
					Return(domain.User{}, errors.New("user not found"))
				return repo
			},
			wantErr: errors.New("user not found"),
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
