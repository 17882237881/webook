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

采用 **分层架构**，实现关注点分离：

```
┌─────────────────────────────────────────────────────────────┐
│                      Handler 层 (web/)                       │
│  负责：HTTP 请求处理、参数校验、响应格式化                      │
│  文件：internal/web/user.go                                   │
└─────────────────────────────────────────────────────────────┘
                              ↓ 调用接口
┌─────────────────────────────────────────────────────────────┐
│                    Service 层 (service/)                     │
│  负责：业务逻辑、密码加密、业务错误定义                         │
│  文件：internal/service/user.go                               │
└─────────────────────────────────────────────────────────────┘
                              ↓ 调用接口
┌─────────────────────────────────────────────────────────────┐
│                  Repository 层 (repository/)                 │
│  负责：数据持久化、领域对象转换                                │
│  文件：internal/repository/user.go                            │
└─────────────────────────────────────────────────────────────┘
                              ↓ 调用
┌─────────────────────────────────────────────────────────────┐
│                       DAO 层 (dao/)                          │
│  负责：数据库操作、SQL 执行                                    │
│  文件：internal/repository/dao/user.go                        │
└─────────────────────────────────────────────────────────────┘
```

### 为什么这样设计？

| 优势 | 说明 |
|------|------|
| **可测试性** | 每层都依赖接口，便于 Mock 测试 |
| **可维护性** | 修改一层不影响其他层 |
| **可扩展性** | 如需换数据库，只改 DAO 层 |

### 依赖注入

在 `main.go` 中手动组装依赖：

```go
// DAO → Repository → Service → Handler
userDAO := dao.NewUserDAO(db)
userRepo := repository.NewUserRepository(userDAO)
userSvc := service.NewUserService(userRepo)
u := web.NewUserHandler(userSvc, cfg.JWT.ExpireTime)
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
// internal/service/user.go
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
// internal/repository/dao/user.go
func (d *UserDAO) Insert(ctx context.Context, u User) error {
    err := d.db.WithContext(ctx).Create(&u).Error
    if me, ok := err.(*mysql.MySQLError); ok {
        if me.Number == 1062 { // MySQL 唯一索引冲突错误码
            return ErrDuplicateEmail
        }
    }
    return err
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
封装统一响应工具 `pkg/ginx`：

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
internal/
├── domain/              # 领域对象
│   └── user.go          # User 结构体
├── web/                 # HTTP 处理层
│   ├── user.go          # UserHandler
│   └── middleware/      # 中间件
│       ├── jwt.go       # JWT 认证
│       └── login.go     # Session 认证（备选）
├── service/             # 业务逻辑层
│   └── user.go          # UserService
└── repository/          # 数据持久化层
    ├── user.go          # UserRepository
    └── dao/             # 数据访问对象
        └── user.go      # UserDAO
```

---

## 总结

本模块实现了一个安全、可扩展的用户认证系统：

1. **分层架构**：Handler → Service → Repository → DAO
2. **安全设计**：bcrypt 密码加密、JWT Token、User-Agent 绑定
3. **统一规范**：RESTful API、统一响应格式、错误码体系
4. **问题驱动**：每个设计决策都有其背景和原因

下一步可扩展：
- 短信/邮箱验证码登录
- OAuth 第三方登录
- Token 刷新机制
- 登录日志审计
