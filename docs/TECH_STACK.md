# Webook 技术栈说明

本文档详细介绍 Webook 项目所使用的核心技术、选型原因及其具体作用。

---

## 目录
1. [Web 框架 - Gin](#1-web-框架---gin)
2. [数据库 - MySQL + GORM](#2-数据库---mysql---gorm)
3. [缓存 - Redis](#3-缓存---redis)
4. [消息队列 - RabbitMQ](#4-消息队列---rabbitmq)
5. [认证 - JWT](#5-认证---jwt)
6. [密码加密 - bcrypt](#6-密码加密---bcrypt)
7. [依赖注入 - Wire](#7-依赖注入---wire)
8. [跨域处理 - CORS](#8-跨域处理---cors)
9. [正则表达式 - regexp2](#9-正则表达式---regexp2)
10. [单元测试 - testify + gomock](#10-单元测试---testify---gomock)
11. [技术依赖版本](#11-技术依赖版本)
12. [架构图](#12-架构图)
13. [总结](#13-总结)
---

## 1. Web 框架 - Gin

### 是什么？
[Gin](https://github.com/gin-gonic/gin) 是一个用 Go 语言编写的高性能 HTTP Web 框架。

### 为什么选择 Gin？
| 优势 | 说明 |
|------|------|
| **高性能** | 基于 httprouter，路由性能极高，比标准库快 40 倍 |
| **轻量级** | 核心代码精简，易于学习和使用 |
| **中间件支持** | 灵活的中间件机制，方便扩展功能 |
| **社区活跃** | Go Web 框架中 Star 最高，生态丰富 |

### 项目中的作用
- 处理 HTTP 请求路由（`/users`, `/auth/*` 等）
- 提供中间件机制（JWT 认证、CORS 配置）
- 请求参数绑定和响应输出

### 代码示例
```go
// internal/ioc/web.go
server := gin.Default()
server.Use(cors.New(...))  // 使用中间件
server.Use(middleware.NewJWTMiddlewareBuilder().Build())
```

---

## 2. 数据库 - MySQL + GORM

### 是什么？
- **MySQL**: 关系型数据库管理系统
- **GORM**: Go 语言中最流行的 ORM（对象关系映射）框架

### 为什么选择 MySQL？
| 优势 | 说明 |
|------|------|
| **成熟稳定** | 经过大规模生产验证，可靠性高 |
| **ACID 支持** | 完整的事务支持，保证数据一致性 |
| **生态丰富** | 工具链完善，运维经验丰富 |
| **性能优秀** | InnoDB 引擎支持高并发读写 |

### 为什么选择 GORM？
| 优势 | 说明 |
|------|------|
| **开发效率** | 自动生成 SQL，减少手写代码 |
| **类型安全** | 编译时检查，减少运行时错误 |
| **自动迁移** | 自动同步数据库表结构 |
| **链式调用** | 流畅的 API 设计 |

### 项目中的作用
- 存储用户账户信息（邮箱、密码哈希等）
- 自动创建和迁移数据库表结构
- 提供数据持久化能力

### 代码示例
```go
// internal/ioc/db.go
db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{})
db.AutoMigrate(&dao.User{})  // 自动迁移表结构
```

---

## 3. 缓存 - Redis

### 是什么？
[Redis](https://redis.io/) 是一个开源的内存数据结构存储系统，可用作数据库、缓存和消息中间件。

### 为什么使用 Redis？
| 优势 | 说明 |
|------|------|
| **高性能** | 内存存储，读写性能达 10 万+ QPS |
| **数据结构丰富** | 支持 String、Hash、List、Set、ZSet 等 |
| **原子操作** | 所有操作都是原子的，天然线程安全 |
| **过期机制** | 自动过期删除，适合缓存和 Token 黑名单 |

### 项目中的作用

#### 3.1 用户信息缓存
减少数据库查询压力，提升接口响应速度：

```go
// internal/adapters/outbound/persistence/redis/user_cache.go
func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
    key := fmt.Sprintf("user:info:%d", id)
    val, err := c.client.Get(ctx, key).Bytes()
    // ...
}
```

#### 3.2 Token 黑名单
用于用户退出登录时使 Refresh Token 失效：

```go
// internal/adapters/outbound/persistence/redis/token_blacklist.go
// 将退出的 Token Session ID 加入黑名单
blacklist.Add(ctx, ssid, expireTime)
// 检查 Token 是否在黑名单中
isBlacklisted, _ := blacklist.IsBlacklisted(ctx, ssid)
```

### 缓存策略
- **用户信息缓存**: 15 分钟过期（可配置）
- **Token 黑名单**: 与 Refresh Token 有效期一致（7 天）

---


## 4. 消息队列 - RabbitMQ

### 是什么？
[RabbitMQ](https://www.rabbitmq.com/) 是基于 AMQP 协议的消息队列系统，适合做异步解耦、削峰填谷、事件驱动。

### 为什么选择 RabbitMQ？
| 优势 | 说明 |
|------|------|
| **成熟稳定** | 社区活跃、生态完善、文档齐全 |
| **支持 ACK** | 至少一次投递，可靠性可控 |
| **易于理解** | 交换机/队列/路由键概念清晰，适合学习 |
| **便于扩展** | 可扩展死信队列、延迟队列等机制 |

### 在本项目中的作用
- 用于帖子互动统计（点赞/收藏/阅读）事件的异步处理
- 将高频写操作转为异步消息，降低数据库压力
- 配合 Redis 实现最终一致计数

### 设计要点
1. **生产者（Publisher）**：
   - 点赞/收藏/阅读在业务层创建事件并发布到 MQ
   - 确保业务主流程不被统计更新阻塞

2. **消费者（Consumer）**：
   - 消费事件并更新 Redis 计数
   - 使用 ACK 保证至少一次投递
   - 通过 Redis 去重 `event_id` 保证幂等

3. **拓扑结构**：
   - Exchange: `post.stats.exchange` (direct)
   - Queue: `post.stats.queue`
   - RoutingKey: `post.stats`

### 代码示例

**发布事件：**
```go
// internal/adapters/outbound/mq/stats_publisher.go
err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
    ContentType:  "application/json",
    DeliveryMode: amqp.Persistent,
    Body:         body,
})
```

**消费事件：**
```go
// internal/adapters/outbound/mq/stats_consumer.go
msg, ok := <-msgs
if !ok { return }
if err := c.handleMessage(ctx, msg); err != nil {
    _ = msg.Nack(false, true)
    return
}
_ = msg.Ack(false)
```

### 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| AMQP_URL | amqp://guest:guest@localhost:5672/ | RabbitMQ 地址 |
| MQ_EXCHANGE | post.stats.exchange | 交换机 |
| MQ_QUEUE | post.stats.queue | 队列 |
| MQ_ROUTING_KEY | post.stats | 路由键 |
| MQ_PREFETCH | 50 | 消费者预取数 |

### 常见问题

- **服务启动就 panic**：RabbitMQ 未启动或连接配置错误
- **统计不更新**：消费者未启动或队列有堆积消息
- **重复计数**：需要依赖 `event_id` 去重

---

## 5. 认证 - JWT

### 是什么？
JWT（JSON Web Token）是一种开放标准（RFC 7519），用于在各方之间安全传输 JSON 格式的信息。

### JWT 结构
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.    # Header（算法和类型）
eyJ1c2VySWQiOjEyMzQ1LCJleHAiOjE2...}.    # Payload（用户数据）
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c  # Signature（签名）
```

### 为什么选择 JWT？
| 优势 | Session 方案 | JWT 方案 |
|------|------------|---------|
| **无状态** | 服务端存储会话数据 | 客户端携带所有信息 |
| **可扩展** | 需要共享 Session 存储 | 任何服务器都可验证 |
| **跨域** | Cookie 受限 | Header 携带，不受限 |
| **移动端友好** | Cookie 支持较差 | 直接使用 Token |

### 项目中的作用

#### 4.1 双 Token 机制
| Token 类型 | 有效期 | 用途 |
|-----------|-------|------|
| **Access Token** | 30 分钟 | 访问 API 接口 |
| **Refresh Token** | 7 天 | 刷新 Access Token |

#### 4.2 安全增强措施
- **User-Agent 绑定**: Token 与设备绑定，防止盗用
- **黑名单机制**: 退出登录后使 Token 失效

### 代码示例
```go
// internal/adapters/inbound/http/middleware/jwt.go
type UserClaims struct {
    UserId    int64  `json:"userId"`      // 用户 ID
    UserAgent string `json:"userAgent"`   // 设备指纹
    jwt.RegisteredClaims                   // 标准字段（过期时间等）
}

// 生成 Token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString(jwtSecretKey)
```

---

## 6. 密码加密 - bcrypt

### 是什么？
bcrypt 是一种基于 Blowfish 加密算法的密码哈希函数，专为密码存储设计。

### 为什么选择 bcrypt？
| 特性 | 说明 |
|------|------|
| **自适应** | 可调整计算成本（cost），随硬件提升增加难度 |
| **内置盐值** | 自动生成并嵌入随机盐值，防彩虹表攻击 |
| **慢哈希** | 故意设计得较慢，增加暴力破解难度 |
| **单向函数** | 不可逆，无法从哈希值还原密码 |

### 为什么不用 MD5/SHA256？
| 算法 | 问题 |
|------|------|
| **MD5** | 速度快（每秒数十亿次），易被暴力破解；存在碰撞漏洞 |
| **SHA256** | 同样速度太快，不适合密码场景 |
| **bcrypt** | 故意慢，每秒仅数千次，暴力破解不可行 |

### 项目中的作用
```go
// internal/application/user.go
// 注册时加密密码
hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

// 登录时验证密码
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

---

## 7. 依赖注入 - Wire

### 是什么？
[Wire](https://github.com/google/wire) 是 Google 开发的编译时依赖注入工具。

### 为什么使用 Wire？
| 优势 | 说明 |
|------|------|
| **编译时检查** | 构建时发现依赖错误，而非运行时 |
| **代码生成** | 自动生成初始化代码，无反射开销 |
| **易于理解** | 生成的代码是普通 Go 代码，便于调试 |
| **无运行时依赖** | 最终二进制无需 Wire 库 |

### 项目中的分层架构

采用 **端口与适配器架构（六边形架构）** 设计，通过接口实现解耦：

```
┌─────────────────────────────────────────────────────────────┐
│                Inbound Adapters (adapters/inbound/)         │
│         http/UserHandler - HTTP 请求处理（Gin）               │
├─────────────────────────────────────────────────────────────┤
│                 Application 层 (application/)               │
│    user.go / post.go / auth.go - 业务逻辑/用例实现            │
│              ↓ 依赖 ports 接口（依赖倒置）                    │
├─────────────────────────────────────────────────────────────┤
│                      Ports 层 (ports/)                       │
│    input/  - 入站端口（Service 接口定义）                     │
│    output/ - 出站端口（Repo, Cache, Token 接口定义）          │
├─────────────────────────────────────────────────────────────┤
│               Outbound Adapters (adapters/outbound/)        │
│    persistence/mysql - MySQL 持久化实现                      │
│    persistence/redis - Redis 缓存实现                        │
│    jwt/              - JWT Token 服务实现                     │
│    repository/       - Repository 模式实现                   │
├─────────────────────────────────────────────────────────────┤
│                       IOC 层 (ioc/)                          │
│              依赖管理与组装（Wire Provider）                  │
└─────────────────────────────────────────────────────────────┘
```

**Ports 接口设计优势：**

| 优势 | 说明 |
|------|------|
| **依赖倒置** | 高层模块依赖抽象（接口），不依赖具体实现 |
| **可测试性** | 通过 Mock 接口实现纯单元测试 |
| **可替换性** | 如需更换 JWT 库或缓存方案，只需实现新适配器 |

### 代码示例
```go
// cmd/webook/wire.go
import (
    cache "webook/internal/adapters/outbound/persistence/redis"
    dao "webook/internal/adapters/outbound/persistence/mysql"
    service "webook/internal/application"
    web "webook/internal/adapters/inbound/http"
    "webook/internal/adapters/outbound/repository"
)

func InitWebServer(cfg *config.Config) *gin.Engine {
    wire.Build(
        // 基础设施
        ioc.NewDB,                              // MySQL
        ioc.NewRedis,                           // Redis
        ioc.NewLogger,                          // Logger
        ioc.NewJWTService,                      // JWT 实现
        ioc.NewTokenService,                    // TokenService 接口
        ioc.NewAccessTokenVerifier,             // Token 验证器

        // DAO 层 (Adapters/Outbound/Persistence/MySQL)
        dao.NewUserDAO,
        dao.NewPostDAO,
        dao.NewPublishedPostDAO,

        // Cache 层 (Adapters/Outbound/Persistence/Redis)
        ProvideUserCacheExpiration,
        cache.NewUserCache,
        cache.NewTokenBlacklist,
        cache.NewPostCache,

        // Repository 层 (Adapters/Outbound/Repository)
        repository.NewUserRepository,
        repository.NewCachedUserRepository,     // 装饰器模式
        repository.NewPostRepository,
        repository.NewPublishedPostRepository,
        repository.NewCachedPublishedPostRepository,

        // Application 层
        service.NewUserService,
        service.NewPostService,
        ProvideAccessExpireTime,
        ProvideRefreshExpireTime,
        service.NewAuthService,                 // 认证服务

        // Inbound Adapters (HTTP)
        web.NewUserHandler,
        web.NewPostHandler,
        ioc.NewGinEngine,
    )
    return nil
}
```

### 生成依赖注入代码

```powershell
.\script\dev\gen-wire.ps1
```

```bash
bash script/dev/gen-wire.sh
```

---

## 8. 跨域处理 - CORS

### 是什么？
CORS（Cross-Origin Resource Sharing）是一种机制，允许浏览器向不同源的服务器发起请求。

### 为什么需要 CORS？
浏览器的**同源策略**默认阻止跨域请求。当前端和后端部署在不同域名/端口时，需要配置 CORS：

| 场景 | 前端地址 | 后端地址 | 是否跨域 |
|------|---------|---------|---------|
| 开发环境 | `http://localhost:5173` | `http://localhost:8080` | ✅ 跨域 |
| 生产环境 | `https://www.example.com` | `https://api.example.com` | ✅ 跨域 |

### 项目中的配置
```go
// ioc/web.go
server.Use(cors.New(cors.Config{
    AllowOrigins:     cfg.CORS.AllowOrigins,  // 允许的源
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,  // 允许携带凭证
}))
```

---

## 9. 正则表达式 - regexp2

### 是什么？
[regexp2](https://github.com/dlclark/regexp2) 是一个功能丰富的正则表达式库，兼容 .NET/Perl 语法。

### 为什么不用标准库 regexp？
| 特性 | 标准库 regexp | regexp2 |
|------|--------------|---------|
| **回溯** | ❌ 不支持 | ✅ 支持 |
| **前瞻/后顾** | ❌ 不支持 | ✅ 支持 |
| **性能** | 快（RE2 算法） | 较慢但功能更全 |
| **复杂规则** | 受限 | 完整支持 |

### 项目中的作用
用于验证用户输入的格式（邮箱、密码强度）：

```go
// internal/adapters/inbound/http/user.go
emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
passwordRegex = `^.{6,16}$`

emailExp := regexp.MustCompile(emailRegex, regexp.None)
ok, _ := emailExp.MatchString(req.Email)
```

---

## 10. 单元测试 - testify + gomock

### 是什么？
- **testify**: 提供断言和 mock 工具的测试框架
- **gomock**: Google 出品的 mock 代码生成工具

### 为什么使用？
| 工具 | 作用 |
|------|------|
| **testify** | 丰富的断言方法（`assert.Equal`, `require.NoError` 等） |
| **gomock** | 自动生成接口的 mock 实现，方便单元测试 |

### 项目中的作用
```go
// internal/application/user_test.go
func TestUserService_Login(t *testing.T) {
    ctrl := gomock.NewController(t)
    mockRepo := mocks.NewMockUserRepository(ctrl)
    
    // 设置 mock 期望
    mockRepo.EXPECT().FindByEmail(gomock.Any(), "test@example.com").
        Return(domain.User{...}, nil)
    
    // 执行测试
    svc := application.NewUserService(mockRepo)
    user, err := svc.Login(ctx, "test@example.com", "password")
    
    // 断言
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
}
```

---

## 11. 技术依赖版本

| 技术 | 版本 | 包路径 |
|------|------|--------|
| Go | 1.25.4 | - |
| Gin | v1.11.0 | `github.com/gin-gonic/gin` |
| GORM | v1.31.1 | `gorm.io/gorm` |
| MySQL Driver | v1.6.0 | `gorm.io/driver/mysql` |
| Redis | v9.17.2 | `github.com/redis/go-redis/v9` |
| JWT | v5.3.0 | `github.com/golang-jwt/jwt/v5` |
| Wire | v0.7.0 | `github.com/google/wire` |
| bcrypt | v0.46.0 | `golang.org/x/crypto` |
| CORS | v1.7.6 | `github.com/gin-contrib/cors` |
| regexp2 | v1.11.5 | `github.com/dlclark/regexp2` |
| testify | v1.11.1 | `github.com/stretchr/testify` |
| gomock | v0.5.0 | `go.uber.org/mock` |

---

## 12. 架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                Inbound Adapters (adapters/inbound)               │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐ │
│  │   CORS      │  │   JWT       │  │     http/UserHandler      │ │
│  │ Middleware  │──│ Middleware  │──│ http/PostHandler         │ │
│  └─────────────┘  └─────────────┘  └───────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                   Application 层 (application/)                  │
│        user.go / post.go / auth.go (业务用例实现)                 │
└──────────────────────────────────────────────────────────────────┘
                        │ 依赖 ports 接口
                        ▼
┌──────────────────────────────────────────────────────────────────┐
│                       Ports 层 (ports/)                          │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ input/  (Service 接口)                                       │ │
│  │ output/ (Repository, Cache, Token 接口)                      │ │
│  └─────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
                        │ 实现
                        ▼
┌──────────────────────────────────────────────────────────────────┐
│               Outbound Adapters (adapters/outbound)              │
│  ┌─────────────────────────┐  ┌────────────────────────────────┐ │
│  │ persistence/mysql       │  │ persistence/redis              │ │
│  │ (UserDAO, PostDAO)      │  │ (UserCache, PostCache)         │ │
│  └─────────────────────────┘  └────────────────────────────────┘ │
│  ┌─────────────────────────┐  ┌────────────────────────────────┐ │
│  │ jwt/jwt.go              │  │ repository/                    │ │
│  │ (Token服务实现)          │  │ (Repository 模式实现)           │ │
│  └─────────────────────────┘  └────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
                        ┌──────────┴──────────┐
                        │   MySQL  /  Redis   │
                        └─────────────────────┘
```

---

## 13. 总结

本项目采用了 **端口与适配器架构（六边形架构 / Clean Architecture）** 设计，各层职责清晰：

| 层级 | 目录 | 职责 |
|------|------|------|
| 入站适配器 | `adapters/inbound/http` | HTTP 路由、中间件、参数验证 |
| **应用层** | `application/` | **核心业务逻辑与用例实现** |
| 端口层 | `ports/input & output` | 接口契约定义，实现依赖倒置 |
| **持久化适配器** | `adapters/outbound/persistence` | **MySQL/Redis 实现，Repo 模式实现** |
| JWT 适配器 | `adapters/outbound/jwt` | JWT Token 服务实现 |
| **IOC 层** | `ioc/` | **编译时依赖注入、组件组装** |

**设计亮点：**
- 通过 `ports` 包定义所有接口契约，实现 **依赖倒置原则 (DIP)**
- 严格遵循 **六边形架构** 目录规范，区分 Inbound/Outbound
- 使用 **装饰器模式** 实现缓存层（如 `CachedUserRepository`）
- `AuthService` 在应用层抽象认证逻辑，与 HTTP 适配器解耦
- `adapters/outbound/jwt/jwt.go` 实现 `ports/output.TokenService` 接口

---

