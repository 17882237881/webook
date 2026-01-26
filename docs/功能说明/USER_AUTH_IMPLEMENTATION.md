# Webook 用户登录注册功能实现文档

## 目录

1. [功能概述](#功能概述)
2. [架构设计](#架构设计)
3. [核心功能实现](#核心功能实现)
4. [遇到的问题与解决方案](#遇到的问题与解决方案)
5. [安全增强](#安全增强)
6. [API 接口文档](#api-接口文档)

---

## 功能概述

本模块实现了完整的用户认证系统，包括：

| 功能 | 接口 | 说明 |
|------|------|------|
| 用户注册 | `POST /users` | 邮箱 + 密码注册 |
# Webook 用户登录注册功能实现文档

## 目录

1. [功能概述](#功能概述)
2. [架构设计](#架构设计)
3. [核心功能实现](#核心功能实现)
4. [遇到的问题与解决方案](#遇到的问题与解决方案)
5. [安全增强](#安全增强)
6. [API 接口文档](#api-接口文档)

---

## 功能概述

本模块实现了完整的用户认证系统，包括：

| 功能 | 接口 | 说明 |
|------|------|------|
| 用户注册 | `POST /users` | 邮箱 + 密码注册 |
| 用户登录 | `POST /users/login` | 返回 JWT Token |
| 获取用户信息 | `GET /users/:id` | 需要登录 |
| 修改密码 | `PUT /users/:id/password` | 需要登录 |

---

## 架构设计

采用 **端口与适配器架构（六边形架构）**，通过接口实现关注点分离：

```
┌─────────────────────────────────────────────────────────────┐
│             Inbound Adapters (adapters/inbound/http/)       │
│  负责：HTTP 请求处理、参数校验、响应格式化                      │
│  文件：internal/adapters/inbound/http/user.go                 │
└─────────────────────────────────────────────────────────────┘
                              ↓ 调用接口
┌─────────────────────────────────────────────────────────────┐
│                 Application 层 (application/)               │
│  user.go - 用户业务逻辑（密码加密、登录验证）                  │
│  auth.go - 认证服务（Token 生成/刷新/退出）                    │
└─────────────────────────────────────────────────────────────┘
                              ↓ 依赖 ports 接口
┌─────────────────────────────────────────────────────────────┐
│                      Ports 层 (ports/)                       │
│  input/  - 业务接口 (UserService, AuthService)               │
│  output/ - 基础设施接口 (UserRepo, TokenService, Cache)       │
└─────────────────────────────────────────────────────────────┘
                    ↓ 实现                    ↓ 实现
┌────────────────────────────┐    ┌────────────────────────────┐
│   Outbound Persistence     │    │       JWT Adapter          │
│  MySQL/Redis 实现           │    │  JWT 实现 (TokenService)    │
│  adapters/outbound/        │    │  adapters/outbound/        │
│  persistence/              │    │  jwt/jwt.go                │
└────────────────────────────┘    └────────────────────────────┘
```

### 为什么这样设计？

| 优势 | 说明 |
|------|------|
| **依赖倒置** | 应用层依赖 ports 接口，不依赖具体实现 |
| **可测试性** | 通过 Mock 接口实现纯单元测试 |
| **可维护性** | 修改一层不影响其他层 |
| **可替换性** | 如需换 JWT 库，只需实现新的 `TokenService` 适配器 |

### 依赖注入 (Wire)

使用 [Google Wire](https://github.com/google/wire) 进行编译时依赖注入，自动生成依赖组装代码。

**依赖链：**

```
config.Load → ioc.NewDB → dao.NewUserDAO → repository.NewUserRepository → service.NewUserService → web.NewUserHandler → ioc.NewGinEngine
                  ↓                                      ↑
            ioc.NewRedis → cache.NewUserCache ──────────┘
```

**Wire 配置 (`cmd/webook/wire.go`):**

```go
//go:build wireinject

func InitWebServer(cfg *config.Config) *gin.Engine {
    wire.Build(
        // 基础设施
        ioc.NewDB,                    // 数据库
        ioc.NewRedis,                 // Redis
        ioc.NewLogger,                // 日志
        ioc.NewJWTService,            // JWT 实现
        ioc.NewTokenService,          // TokenService 接口
        ioc.NewAccessTokenVerifier,   // Token 验证器

        // DAO 层 (Adapters/Outbound/Persistence/MySQL)
        dao.NewUserDAO,

        // Cache 层 (Adapters/Outbound/Persistence/Redis)
        ProvideUserCacheExpiration,
        cache.NewUserCache,
        cache.NewTokenBlacklist,

        // Repository 层 (Adapters/Outbound/Repository)
        repository.NewUserRepository,
        repository.NewCachedUserRepository,  // 装饰器模式

        // Application 层
        service.NewUserService,
        ProvideAccessExpireTime,
        ProvideRefreshExpireTime,
        service.NewAuthService,       // 认证服务

        // Inbound Adapters (HTTP)
        web.NewUserHandler,
        ioc.NewGinEngine,
    )
    return nil
}
```

**Wire 命令：**

```powershell
# 安装 Wire
go install github.com/google/wire/cmd/wire@v0.7.0

# 生成依赖注入代码
cd d:\go\webook
# 推荐：统一入口
.\\script\\dev\\gen-wire.ps1

# 或（Linux/macOS）
bash script/dev/gen-wire.sh

# 也可以直接用 wire
wire ./cmd/webook
```

生成的 `wire_gen.go` 包含真正的依赖组装代码，`main.go` 直接调用：

```go
func main() {
    cfg := config.Load()
    server := InitWebServer(cfg)
    server.Run(cfg.Server.Port)
}
```

---

## 核心功能实现

### 1. 用户注册

**流程图：**

```
用户提交 → 参数校验 → 邮箱格式验证 → 密码一致性检查 → 密码加密 → 存入数据库
                ↓
            密码强度验证
```

**密码加密：** 使用 `bcrypt` 算法

```go
// internal/application/user.go
hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
u.Password = string(hash)
```

**为什么用 bcrypt？**
- 自动加盐，防止彩虹表攻击
- 可调节计算成本，抵抗暴力破解
- 业界标准，Go 官方推荐

### 2. 用户登录

**流程图：**

```
用户提交 → 查询用户 → 验证密码 → 生成 JWT Token → 返回 Token
              ↓
         用户不存在/密码错误 → 返回统一错误（防信息泄露）
```

**统一错误处理：** 无论是用户不存在还是密码错误，都返回相同的错误信息：

```go
var ErrInvalidUserOrPassword = errors.New("邮箱或密码不正确")
```

**为什么？** 防止攻击者通过错误信息探测有效邮箱

### 3. JWT 认证

**Token 结构：**

```go
type UserClaims struct {
    UserId    int64  `json:"userId"`    // 用户 ID
    UserAgent string `json:"userAgent"` // User-Agent 哈希（安全增强）
    jwt.RegisteredClaims                // 标准字段（过期时间等）
}
```

**Token 验证流程（中间件）：**

```
请求到达 → 检查白名单 → 提取 Token → 验证签名 → 检查过期 → 验证 User-Agent → 放行
```

### 3.1 长短 Token 机制

使用 **Access Token + Refresh Token** 双 Token 机制，平衡安全性和用户体验。

| Token | 有效期 | 用途 | 存储建议 |
|-------|--------|------|----------|
| Access Token | 30 分钟 | API 访问认证 | 内存/sessionStorage |
| Refresh Token | 7 天 | 刷新 Access Token | HttpOnly Cookie |

**工作流程：**

```
1. 用户登录 → 返回 Access Token + Refresh Token
2. 用 Access Token 访问 API
3. Access Token 过期（401）→ 用 Refresh Token 调用 /auth/refresh
4. 获取新的 Access Token → 继续访问
5. Refresh Token 过期 → 重新登录
```

**Refresh Token 刷新原理：**

```go
// POST /auth/refresh
func (u *UserHandler) RefreshToken(c *gin.Context) {
    // 1. 解析 Refresh Token，提取 userId
    claims, err := middleware.ParseRefreshToken(req.RefreshToken)
    
    // 2. 用 userId 生成新的 Access Token
    accessToken, err := middleware.GenerateAccessToken(claims.UserId, userAgent, expireTime)
    
    // 3. 返回新 Token
    ginx.Success(c, gin.H{"accessToken": accessToken})
}
```

**关键设计：**
- **独立密钥**：Access Token 和 Refresh Token 使用不同的签名密钥
- **最小化载荷**：Refresh Token 只存储 userId 和 SSid
- **白名单**：`/auth/refresh` 和 `/auth/logout` 接口无需 Access Token 认证

### 3.2 退出登录

使用 **Redis 黑名单** 机制使 Refresh Token 失效。

**API 接口：**

```
POST /auth/logout
请求体：{ "refreshToken": "xxx" }
响应：{ "msg": "退出成功" }
```

**工作流程：**

```
1. 用户调用 /auth/logout，携带 refreshToken
2. 服务端解析 Token，提取 SSid
3. 将 SSid 加入 Redis 黑名单（TTL = Refresh Token 剩余有效期）
4. 后续使用该 Refresh Token 刷新时被拒绝
```

**实现代码（基于 AuthService 抽象）：**

```go
// internal/application/auth.go
func (a *authService) Logout(ctx context.Context, refreshToken string) error {
    // 1. 解析 Refresh Token 获取 SSid
    claims, err := a.tokens.ParseRefreshToken(refreshToken)
    // ...
}

// internal/adapters/inbound/http/user.go
func (u *UserHandler) Logout(c *gin.Context) {
    // ...
    if err := u.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
        ginx.Error(c, ginx.CodeInternalError, "logout failed")
        return
    }
    ginx.SuccessMsg(c, "logout success")
}
```

**黑名单 Key 设计：**
```
token:blacklist:{ssid}
```

---

### 4. Redis 缓存层

**缓存策略：Cache-Aside 模式**

```
查询用户 → 检查缓存 → 命中 → 直接返回
              ↓ 未命中
          查询数据库 → 回写缓存 → 返回
```

**缓存实现：**

```go
// internal/adapters/outbound/persistence/redis/user_cache.go
type UserCache interface {
    Get(ctx context.Context, id int64) (domain.User, error)
    Set(ctx context.Context, u domain.User) error
    Delete(ctx context.Context, id int64) error
}
```

**缓存 Key 设计：**

```go
func (c *RedisUserCache) key(id int64) string {
    return fmt.Sprintf("user:info:%d", id)
}
```

**缓存一致性保证：**

| 操作 | 缓存处理 |
|------|----------|
| 查询用户 | 先查缓存，未命中查 DB 并异步回写 |
| 修改密码 | 更新 DB 后删除缓存 |
| 注册用户 | 不预热缓存（首次登录时缓存） |

**配置项（config/config.go）：**

```go
type CacheConfig struct {
    UserExpiration time.Duration // 用户缓存过期时间，默认 15 分钟
}
```

---

## 遇到的问题与解决方案

### 问题 1：正则表达式库选型

**问题描述：**  
Go 标准库 `regexp` 不支持复杂正则（如零宽断言），无法实现完整的密码强度验证。

**解决方案：**  
使用第三方库 `github.com/dlclark/regexp2`，支持完整的正则语法。

```go
import regexp "github.com/dlclark/regexp2"

emailExp := regexp.MustCompile(emailRegex, regexp.None)
ok, err := emailExp.MatchString(email)
```

---

### 问题 2：邮箱重复注册检测

**问题描述：**  
如何高效检测邮箱是否已被注册？

**解决方案：**  
利用数据库唯一索引约束，在 DAO 层捕获错误：

```go
// internal/adapters/outbound/persistence/mysql/user.go
func (d *UserDAO) Insert(ctx context.Context, u User) error {
    // ...
}
```

**优势：**
- 无需先查询再插入（避免竞态条件）
- 利用数据库保证数据一致性
- 性能更好

---

### 问题 3：密码明文传输风险

**问题描述：**  
密码在服务端以明文接收，存在泄露风险。

**解决方案：**
1. **传输层加密**：必须使用 HTTPS
2. **存储加密**：使用 bcrypt 加密后存储
3. **日志脱敏**：永不记录密码相关日志

---

### 问题 4：Token 被盗用

**问题描述：**  
JWT Token 被盗后，攻击者可在任何设备使用。

**解决方案：**  
将 User-Agent 绑定到 Token 中，验证时检查一致性：

```go
// 生成 Token 时绑定 User-Agent
func GenerateToken(userId int64, userAgent string, expireTime time.Duration) (string, error) {
    claims := UserClaims{
        UserId:    userId,
        UserAgent: hashUserAgent(userAgent), // 存储哈希值
        // ...
    }
}

// 验证时检查 User-Agent
currentUA := hashUserAgent(c.GetHeader("User-Agent"))
if claims.UserAgent != currentUA {
    c.AbortWithStatus(http.StatusUnauthorized)
}
```

---

### 问题 5：统一响应格式

**问题描述：**  
各接口返回格式不统一，前端处理困难。

**解决方案：**  
封装统一响应工具 `internal/adapters/inbound/http/ginx`：

```go
// 成功响应
type Response struct {
    Code int         `json:"code"`
    Msg  string      `json:"msg"`
    Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func Error(c *gin.Context, code int, msg string) {
    c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}
```

---

## 安全增强

| 安全措施 | 实现方式 |
|----------|---------|
| 密码加密存储 | bcrypt 算法 |
| Token 设备绑定 | User-Agent 哈希验证 |
| 统一错误信息 | 防止信息泄露 |
| 权限验证 | 只能访问/修改自己的资源 |
| 接口白名单 | 登录/注册无需 Token |
| Redis 缓存 | 减少 DB 压力，提升性能 |

---

## API 接口文档

### POST /users - 用户注册

**请求体：**
```json
{
    "email": "user@example.com",
    "password": "123456",
    "confirmPassword": "123456"
}
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "注册成功"
}
```

**错误响应：**
```json
{
    "code": 400002,
    "msg": "邮箱已被注册"
}
```

---

### POST /users/login - 用户登录

**请求体：**
```json
{
    "email": "user@example.com",
    "password": "123456"
}
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "userId": 1,
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```

---

### GET /users/:id - 获取用户信息

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
        "email": "user@example.com"
    }
}
```

---

### PUT /users/:id/password - 修改密码

**请求头：**
```
Authorization: Bearer <token>
```

**请求体：**
```json
{
    "oldPassword": "123456",
    "newPassword": "654321"
}
```

**成功响应：**
```json
{
    "code": 0,
    "msg": "密码修改成功"
}
```

---

## 项目文件结构

```
webook/
├── cmd/webook/
│   ├── wire.go          # Wire 注入器定义
│   └── main.go          # 应用入口
├── internal/
│   ├── domain/          # 🔵 核心领域层
│   ├── application/     # 🟢 应用层 (业务逻辑)
│   ├── ports/           # 🟡 端口层 (接口定义)
│   │   ├── input/       #   入站端口
│   │   └── output/      #   出站端口
│   ├── adapters/        # 🔴 适配器层
│   │   ├── inbound/     #   入站 (HTTP)
│   │   └── outbound/    │   出站 (MySQL/Redis/JWT)
│   └── ioc/             # ⚙️ IOC 容器
└── pkg/                 # 公共公共库
```

---

## 总结

本模块实现了一个安全、可扩展的用户认证系统：

1. **端口与适配器架构**：Web Adapter → Application → Ports(接口) → Persistence/JWT Adapters
2. **AuthService 抽象**：Token 生成/刷新/退出逻辑与 Web Adapter 解耦
3. **Wire 依赖注入**：编译时代码生成，零运行时开销
4. **安全设计**：bcrypt 密码加密、JWT Token、User-Agent 绑定
5. **性能优化**：Redis 缓存用户信息，装饰器模式实现缓存层
6. **统一规范**：RESTful API、统一响应格式、错误码体系

下一步可扩展：
- 短信/邮箱验证码登录
- OAuth 第三方登录
- 登录日志审计
- 登录接口缓存优化（email → userId 索引）
