# Webook 帖子管理功能实现文档

## 目录

1. [功能概述](#功能概述)
2. [架构设计](#架构设计)
3. [核心功能实现](#核心功能实现)
4. [数据模型](#数据模型)
5. [API 接口文档](#api-接口文档)
6. [测试用例](#测试用例)

---

## 功能概述

本模块实现了完整的帖子管理系统，支持草稿与已发布帖子的双表存储设计：

| 功能 | 接口 | 说明 |
|------|------|------|
| 保存草稿 | `POST /posts` | 创建或更新草稿 |
| 发布帖子 | `POST /posts/publish` | 将帖子同步到已发布表 |
| 获取草稿 | `GET /posts/draft/:id` | 作者查看自己的草稿 |
| 获取已发布帖子 | `GET /posts/:id` | 读者查看已发布帖子 |
| 作者帖子列表 | `GET /posts/author` | 作者查看自己的所有帖子 |
| 公开帖子列表 | `GET /posts` | 读者浏览所有已发布帖子 |
| 删除帖子 | `DELETE /posts/:id` | 软删除，修改状态为私有 |

---

## 架构设计

采用 **端口与适配器架构（六边形架构）**，实现业务逻辑与基础设施解耦：

```
┌─────────────────────────────────────────────────────────────┐
│             Inbound Adapters (adapters/inbound/http/)       │
│  负责：HTTP 请求处理、参数校验、权限验证                         │
│  文件：internal/adapters/inbound/http/post.go                 │
└─────────────────────────────────────────────────────────────┘
                              ↓ 调用接口
┌─────────────────────────────────────────────────────────────┐
│                 Application 层 (application/)               │
│  post.go - 业务逻辑（分页计算、发布流程、权限校验）               │
└─────────────────────────────────────────────────────────────┘
                              ↓ 依赖 ports 接口
┌─────────────────────────────────────────────────────────────┐
│                      Ports 层 (ports/)                       │
│  input/  - PostService 接口定义                              │
│  output/ - PostRepository, PublishedPostRepository, Cache    │
└─────────────────────────────────────────────────────────────┘
                    ↓ 实现                    ↓ 实现
┌────────────────────────────┐    ┌────────────────────────────┐
│   Outbound Persistence     │    │      Outbound Redis        │
│  MySQL 实现                 │    │  Redis 缓存实现            │
│  adapters/outbound/        │    │  adapters/outbound/        │
│  persistence/mysql/        │    │  persistence/redis/        │
└────────────────────────────┘    └────────────────────────────┘
```

### 双表设计

为了实现草稿与已发布帖子的分离，采用双表存储：

| 表名 | 用途 | 特点 |
|------|------|------|
| `posts` | 草稿表 | 存储所有帖子（含草稿、已发布、已删除） |
| `published_posts` | 已发布表 | 仅存储已发布的帖子，供读者浏览 |

**优势：**
- 读者查询已发布帖子时无需过滤草稿
- 可针对已发布表单独优化（索引、缓存）
- 方便实现发布审核流程

---

## 核心功能实现

### 1. 保存草稿

**流程：**
```
用户提交 → 参数校验 → 获取用户ID → 判断新建/更新 → 持久化 → 返回ID
```

**Application 层实现：**
```go
// internal/application/post.go
func (s *postService) Save(ctx context.Context, p domain.Post) (int64, error) {
    if p.Id == 0 {
        return s.repo.Create(ctx, p)  // 新建
    }
    return p.Id, s.repo.Update(ctx, p)  // 更新
}
```

**设计特点：**
- 通过 `Id == 0` 区分新建和更新
- 所有保存操作只影响草稿表

---

### 2. 发布帖子

**流程：**
```
用户提交 → 参数校验 → 设置状态为已发布 → 同步到两张表 → 返回ID
```

**Persistence Adapter 实现（事务同步）：**
```go
// internal/adapters/outbound/persistence/mysql/post.go
func (d *PostDAO) Sync(ctx context.Context, p Post) (int64, error) {
    // 使用事务同步草稿表和已发布表
    // ...
}
```

**设计特点：**
- 使用数据库事务保证一致性
- 支持"新建并直接发布"场景
- 使用 Upsert 避免重复发布问题

---

### 3. 获取帖子列表

#### 3.1 作者帖子列表

只能获取自己的帖子：

```go
// GET /posts/author?page=1&pageSize=10
func (h *PostHandler) ListByAuthor(c *gin.Context) {
    authorId := c.GetInt64("userId")  // 从 JWT 中间件获取
    posts, err := h.svc.ListByAuthor(c.Request.Context(), authorId, page, pageSize)
    // ...
}
```

#### 3.2 公开帖子列表

从 `published_posts` 表读取，支持缓存：

```go
// internal/adapters/outbound/repository/post.go
func (r *cachedPublishedPostRepository) FindById(ctx context.Context, id int64) (domain.Post, error) {
    // 带有缓存的查询实现
    // ...
}
```

---

### 4. 删除帖子

采用**软删除**，修改状态为私有，同时从 `published_posts` 表移除：

```go
// internal/adapters/outbound/persistence/mysql/post.go
func (d *PostDAO) SyncStatus(ctx context.Context, id, authorId int64, status uint8) error {
    // 软删除同步逻辑
    // ...
}
```

---

## 数据模型

### Domain 层

```go
// internal/domain/post.go
type Post struct {
    Id       int64
    Title    string
    Content  string
    AuthorId int64
    Status   uint8  // 0: 草稿, 1: 已发布, 2: 私有
    Ctime    int64  // 创建时间戳
    Utime    int64  // 更新时间戳
}

// internal/domain/post.go
const (
    PostStatusDraft     uint8 = 0
    PostStatusPublished uint8 = 1
    PostStatusPrivate   uint8 = 2
)
```

// internal/adapters/outbound/persistence/mysql/user_model.go
// 数据库模型定义
type Post struct { ... }
type PublishedPost struct { ... }
```

---

// internal/ports/output/post_repository.go
type PostRepository interface { ... }

// internal/ports/output/cache.go
type PostCache interface { ... }
```

---

## API 接口文档

### POST /posts - 保存草稿

**请求头：**
```
Authorization: Bearer <token>
```

**请求体：**
```json
{
    "id": 0,           // 0 表示新建，非 0 表示更新
    "title": "帖子标题",
    "content": "帖子内容..."
}
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1
    }
}
```

---

### POST /posts/publish - 发布帖子

**请求头：**
```
Authorization: Bearer <token>
```

**请求体：**
```json
{
    "id": 0,           // 0 表示新建并发布
    "title": "帖子标题",
    "content": "帖子内容..."
}
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1
    }
}
```

---

### GET /posts/:id - 获取已发布帖子

**无需认证（公开接口）**

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1,
        "title": "帖子标题",
        "content": "帖子内容...",
        "authorId": 123,
        "ctime": 1705900800000,
        "utime": 1705900800000
    }
}
```

---

### GET /posts/draft/:id - 获取草稿

**请求头：**
```
Authorization: Bearer <token>
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "id": 1,
        "title": "帖子标题",
        "content": "帖子内容...",
        "status": 0,
        "ctime": 1705900800000,
        "utime": 1705900800000
    }
}
```

**权限校验：** 只能访问自己的草稿，否则返回 403

---

### GET /posts/author - 作者帖子列表

**请求头：**
```
Authorization: Bearer <token>
```

**查询参数：**
- `page`: 页码，默认 1
- `pageSize`: 每页数量，默认 10，最大 100

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "posts": [
            {
                "id": 1,
                "title": "帖子标题",
                "content": "帖子内容...",
                "authorId": 123,
                "status": 1,
                "ctime": 1705900800000,
                "utime": 1705900800000
            }
        ]
    }
}
```

---

### GET /posts - 公开帖子列表

**无需认证（公开接口）**

**查询参数：**
- `page`: 页码，默认 1
- `pageSize`: 每页数量，默认 10，最大 100

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "posts": [
            {
                "id": 1,
                "title": "帖子标题",
                "content": "帖子内容...",
                "authorId": 123,
                "status": 1,
                "ctime": 1705900800000,
                "utime": 1705900800000
            }
        ]
    }
}
```

---

### DELETE /posts/:id - 删除帖子

**请求头：**
```
Authorization: Bearer <token>
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "删除成功"
}
```

**权限校验：** 只能删除自己的帖子

---

## 测试用例

### Application 层单元测试

使用 `gomock` 生成 Repository Mock，实现纯单元测试：

```go
// internal/application/post_test.go
import (
    repomocks "webook/internal/adapters/outbound/mocks"
)
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
```

### 运行测试

```powershell
# 运行所有测试
go test ./internal/application/... -v

# 生成 Mock 文件
mockgen -source=internal/ports/output/post_repository.go -destination=internal/adapters/outbound/mocks/post_mock.go -package=repomocks
```

---

## 总结

帖子管理模块采用了以下设计模式和最佳实践：

| 设计点 | 实现方式 |
|--------|----------|
| **依赖倒置** | 应用层依赖 `ports` 接口，不依赖具体实现 |
| **双表存储** | 草稿表 + 已发布表，分离读写路径 |
| **装饰器模式** | `cachedPublishedPostRepository` 透明增加缓存 |
| **事务保证** | 发布/删除操作使用数据库事务 |
| **权限校验** | Web Adapter 层验证 authorId |
| **可测试性** | 通过接口 Mock 实现纯单元测试 |

**下一步可扩展：**
- 帖子搜索（Elasticsearch）
- 帖子点赞/收藏功能
- 帖子评论系统
- 帖子内容审核
- 帖子分类/标签
