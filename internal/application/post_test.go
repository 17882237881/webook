package application

import (
	"context"
	"errors"
	"testing"
	repomocks "webook/internal/adapters/outbound/mocks"
	"webook/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostService_Save(t *testing.T) {
	tests := []struct {
		name    string
		post    domain.Post
		mock    func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository)
		wantId  int64
		wantErr error
	}{
		{
			name: "创建新帖子成功",
			post: domain.Post{
				Title:    "测试标题",
				Content:  "测试内容",
				AuthorId: 1,
			},
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
				return repo, pubRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "更新已有帖子成功",
			post: domain.Post{
				Id:       1,
				Title:    "更新后的标题",
				Content:  "更新后的内容",
				AuthorId: 1,
			},
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
				return repo, pubRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "创建帖子失败-数据库错误",
			post: domain.Post{
				Title:    "测试标题",
				Content:  "测试内容",
				AuthorId: 1,
			},
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("数据库错误"))
				return repo, pubRepo
			},
			wantId:  0,
			wantErr: errors.New("数据库错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, pubRepo := tt.mock(ctrl)
			svc := NewPostService(repo, pubRepo)

			id, err := svc.Save(context.Background(), tt.post)
			assert.Equal(t, tt.wantId, id)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostService_Publish(t *testing.T) {
	tests := []struct {
		name    string
		post    domain.Post
		mock    func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository)
		wantId  int64
		wantErr error
	}{
		{
			name: "发布帖子成功",
			post: domain.Post{
				Id:       1,
				Title:    "发布的标题",
				Content:  "发布的内容",
				AuthorId: 1,
			},
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					Sync(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
				return repo, pubRepo
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "发布帖子失败-同步错误",
			post: domain.Post{
				Id:       1,
				Title:    "发布的标题",
				Content:  "发布的内容",
				AuthorId: 1,
			},
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					Sync(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("同步失败"))
				return repo, pubRepo
			},
			wantId:  0,
			wantErr: errors.New("同步失败"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, pubRepo := tt.mock(ctrl)
			svc := NewPostService(repo, pubRepo)

			id, err := svc.Publish(context.Background(), tt.post)
			assert.Equal(t, tt.wantId, id)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostService_GetPublishedById(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		mock     func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository)
		wantPost domain.Post
		wantErr  error
	}{
		{
			name: "获取已发布帖子成功",
			id:   1,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				pubRepo.EXPECT().
					FindById(gomock.Any(), int64(1)).
					Return(domain.Post{
						Id:       1,
						Title:    "已发布的标题",
						Content:  "已发布的内容",
						AuthorId: 1,
						Status:   domain.PostStatusPublished,
					}, nil)
				return repo, pubRepo
			},
			wantPost: domain.Post{
				Id:       1,
				Title:    "已发布的标题",
				Content:  "已发布的内容",
				AuthorId: 1,
				Status:   domain.PostStatusPublished,
			},
			wantErr: nil,
		},
		{
			name: "帖子不存在",
			id:   999,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				pubRepo.EXPECT().
					FindById(gomock.Any(), int64(999)).
					Return(domain.Post{}, domain.ErrPostNotFound)
				return repo, pubRepo
			},
			wantPost: domain.Post{},
			wantErr:  domain.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, pubRepo := tt.mock(ctrl)
			svc := NewPostService(repo, pubRepo)

			post, err := svc.GetPublishedById(context.Background(), tt.id)
			assert.Equal(t, tt.wantPost, post)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostService_ListByAuthor(t *testing.T) {
	tests := []struct {
		name      string
		authorId  int64
		page      int
		pageSize  int
		mock      func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository)
		wantPosts []domain.Post
		wantTotal int64
		wantErr   error
	}{
		{
			name:     "获取作者帖子列表成功",
			authorId: 1,
			page:     1,
			pageSize: 10,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					FindByAuthor(gomock.Any(), int64(1), 0, 10).
					Return([]domain.Post{
						{Id: 1, Title: "帖子1", AuthorId: 1},
						{Id: 2, Title: "帖子2", AuthorId: 1},
					}, nil)
				repo.EXPECT().
					CountByAuthor(gomock.Any(), int64(1)).
					Return(int64(2), nil)
				return repo, pubRepo
			},
			wantPosts: []domain.Post{
				{Id: 1, Title: "帖子1", AuthorId: 1},
				{Id: 2, Title: "帖子2", AuthorId: 1},
			},
			wantTotal: 2,
			wantErr:   nil,
		},
		{
			name:     "获取第二页",
			authorId: 1,
			page:     2,
			pageSize: 10,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					FindByAuthor(gomock.Any(), int64(1), 10, 10). // offset = (2-1)*10 = 10
					Return([]domain.Post{}, nil)
				repo.EXPECT().
					CountByAuthor(gomock.Any(), int64(1)).
					Return(int64(12), nil)
				return repo, pubRepo
			},
			wantPosts: []domain.Post{},
			wantTotal: 12,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, pubRepo := tt.mock(ctrl)
			svc := NewPostService(repo, pubRepo)

			posts, total, err := svc.ListByAuthor(context.Background(), tt.authorId, tt.page, tt.pageSize)
			assert.Equal(t, tt.wantPosts, posts)
			assert.Equal(t, tt.wantTotal, total)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		authorId int64
		mock     func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository)
		wantErr  error
	}{
		{
			name:     "删除帖子成功",
			id:       1,
			authorId: 1,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					SyncStatus(gomock.Any(), int64(1), int64(1), domain.PostStatusPrivate).
					Return(nil)
				return repo, pubRepo
			},
			wantErr: nil,
		},
		{
			name:     "删除帖子失败-无权限",
			id:       1,
			authorId: 2,
			mock: func(ctrl *gomock.Controller) (*repomocks.MockPostRepository, *repomocks.MockPublishedPostRepository) {
				repo := repomocks.NewMockPostRepository(ctrl)
				pubRepo := repomocks.NewMockPublishedPostRepository(ctrl)
				repo.EXPECT().
					SyncStatus(gomock.Any(), int64(1), int64(2), domain.PostStatusPrivate).
					Return(domain.ErrPostNotAuthor)
				return repo, pubRepo
			},
			wantErr: domain.ErrPostNotAuthor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, pubRepo := tt.mock(ctrl)
			svc := NewPostService(repo, pubRepo)

			err := svc.Delete(context.Background(), tt.id, tt.authorId)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
