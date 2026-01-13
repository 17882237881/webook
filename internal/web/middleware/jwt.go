package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================================
// JWT (JSON Web Token) 认证模块
// ============================================================================
// JWT 由三部分组成：Header.Payload.Signature
//   - Header:    包含算法类型（如 HS256）和 Token 类型（JWT）
//   - Payload:   包含用户数据（Claims），如 userId、过期时间等
//   - Signature: 使用密钥对 Header 和 Payload 进行签名，防止篡改
//
// JWT 使用步骤：
//   步骤1: 初始化密钥（程序启动时调用 InitJWT）
//   步骤2: 生成 Token（用户登录成功后调用 GenerateToken）
//   步骤3: 携带 Token（前端在请求 Header 中携带 Authorization: Bearer <token>）
//   步骤4: 验证 Token（中间件验证每个请求的 Token）
// ============================================================================

// jwtSecretKey JWT 签名密钥
// 用于生成和验证 Token 的签名，必须保密
// 如果密钥泄露，任何人都可以伪造有效的 Token
var jwtSecretKey []byte

// ============================================================================
// 步骤1: 初始化密钥
// ============================================================================

// InitJWT 初始化 JWT 密钥
// 必须在程序启动时调用，设置用于签名和验证的密钥
// 参数 secretKey: 密钥字符串，建议从环境变量或配置中心读取，不要硬编码
// 使用示例: middleware.InitJWT(os.Getenv("JWT_SECRET"))
func InitJWT(secretKey string) {
	jwtSecretKey = []byte(secretKey)
}

// ============================================================================
// JWT Claims 定义（对应 JWT 的 Payload 部分）
// ============================================================================

// UserClaims 自定义 JWT Claims（JWT 中存储的数据，即 Payload 部分）
// Claims 是 JWT 中用于存储用户信息的载荷
type UserClaims struct {
	// UserId 自定义字段：存储用户 ID
	// 这是我们业务需要的数据，会被编码到 Token 中
	UserId int64 `json:"userId"`

	// RegisteredClaims JWT 标准字段（内嵌）
	// 包含以下标准字段：
	//   - ExpiresAt: 过期时间（exp）- Token 何时失效
	//   - IssuedAt:  签发时间（iat）- Token 何时创建
	//   - NotBefore: 生效时间（nbf）- Token 何时开始有效
	//   - Issuer:    签发者（iss）  - 谁签发的 Token
	//   - Subject:   主题（sub）    - Token 的主题
	//   - ID:        唯一标识（jti）- Token 的唯一 ID
	jwt.RegisteredClaims
}

// ============================================================================
// 步骤2: 生成 Token（用户登录成功后调用）
// ============================================================================

// GenerateToken 生成 JWT Token
// 在用户登录验证成功后调用此函数，生成一个包含用户信息的 Token
//
// 参数:
//   - userId:     用户 ID，将被编码到 Token 中
//   - expireTime: Token 有效期，如 24*time.Hour 表示 24 小时后过期
//
// 返回:
//   - string: 生成的 Token 字符串，格式为 "xxxxx.yyyyy.zzzzz"
//   - error:  如果签名失败则返回错误
//
// 使用示例:
//
//	token, err := middleware.GenerateToken(user.Id, 24*time.Hour)
//	c.Header("x-jwt-token", token) // 返回给前端
func GenerateToken(userId int64, expireTime time.Duration) (string, error) {
	// 第一步：构造 Payload（Claims）
	// 创建包含用户信息和标准字段的 Claims 结构
	claims := UserClaims{
		UserId: userId, // 存入用户 ID
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: 设置过期时间
			// 当前时间 + expireTime = Token 失效时间
			// 验证时会检查此字段，过期的 Token 会被拒绝
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),

			// IssuedAt: 设置签发时间
			// 记录 Token 的创建时间，可用于审计或刷新策略
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// 第二步：创建 Token 对象
	// jwt.NewWithClaims 会自动生成 Header 部分：{"alg":"HS256","typ":"JWT"}
	// jwt.SigningMethodHS256 指定使用 HMAC-SHA256 算法进行签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 第三步：签名生成最终的 Token 字符串
	// SignedString 方法会：
	//   1. 将 Header 和 Payload 分别进行 Base64Url 编码
	//   2. 用密钥对 "Header.Payload" 进行 HMAC-SHA256 签名
	//   3. 将签名结果 Base64Url 编码
	//   4. 拼接成最终格式：Header.Payload.Signature
	return token.SignedString(jwtSecretKey)
}

// ============================================================================
// 步骤4: 验证 Token（中间件）
// ============================================================================

// JWTMiddlewareBuilder JWT 认证中间件构建器
// 使用 Builder 模式，支持链式调用配置，然后构建中间件
// 使用示例:
//
//	middleware := NewJWTMiddlewareBuilder().
//	    IgnorePaths("/users/login", "/users/signup").
//	    Build()
type JWTMiddlewareBuilder struct {
	// paths 白名单路径列表
	// 这些路径不需要 Token 验证，如登录、注册接口
	paths []string
}

// NewJWTMiddlewareBuilder 创建一个新的 JWT 中间件构建器
func NewJWTMiddlewareBuilder() *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{}
}

// IgnorePaths 设置不需要登录校验的路径（白名单）
// 登录、注册等接口不需要 Token，应该加入白名单
// 支持可变参数和链式调用
//
// 使用示例:
//
//	builder.IgnorePaths("/users/login", "/users/signup")
func (j *JWTMiddlewareBuilder) IgnorePaths(paths ...string) *JWTMiddlewareBuilder {
	j.paths = append(j.paths, paths...)
	return j // 返回自身，支持链式调用
}

// Build 构建并返回 Gin 中间件函数
// 这个中间件会在每个请求到达处理函数之前执行，验证 Token 的有效性
func (j *JWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ========================================
		// 第一步：白名单检查
		// ========================================
		// 检查当前请求路径是否在白名单中
		// 白名单中的路径（如登录、注册）不需要 Token 验证
		for _, path := range j.paths {
			if c.Request.URL.Path == path {
				c.Next() // 放行，继续执行后续处理
				return
			}
		}

		// ========================================
		// 第二步：获取 Authorization Header
		// ========================================
		// 前端需要在请求头中携带 Token
		// 格式：Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有提供 Token，返回 401 未授权
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// ========================================
		// 第三步：解析 Token 格式
		// ========================================
		// Token 格式必须是: "Bearer <token>"
		// Bearer 是一种 Token 类型标识，OAuth 2.0 规范推荐使用
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 格式不正确，返回 401 未授权
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := parts[1] // 提取实际的 Token 字符串

		// ========================================
		// 第四步：解析并验证 Token
		// ========================================
		// jwt.ParseWithClaims 会执行以下操作：
		//   1. 将 Token 字符串按 "." 分割成 Header、Payload、Signature
		//   2. Base64 解码 Header 和 Payload
		//   3. 从 Header 中读取签名算法
		//   4. 使用回调函数返回的密钥重新计算签名
		//   5. 比对计算的签名和 Token 中的签名是否一致
		//   6. 检查 Token 是否过期（ExpiresAt）
		//   7. 将 Payload 解析到 claims 结构体中
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// KeyFunc: 返回用于验证签名的密钥
			// 这里直接返回之前初始化的密钥
			// 高级用法：可以根据 token.Header 中的信息动态选择密钥
			return jwtSecretKey, nil
		})

		// ========================================
		// 第五步：检查验证结果
		// ========================================
		// err != nil: 解析或验证过程中出错（格式错误、签名不匹配等）
		// !token.Valid: Token 无效（已过期、未生效等）
		if err != nil || !token.Valid {
			// Token 无效，返回 401 未授权
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// ========================================
		// 第六步：将用户信息存入 Context
		// ========================================
		// 验证通过后，将 claims 中的 userId 存入 Gin 的 Context
		// 后续的处理函数可以通过 c.Get("userId") 或 c.GetInt64("userId") 获取
		// 这样就实现了"无状态"认证：无需查询数据库或 Session，直接从 Token 中获取用户信息
		c.Set("userId", claims.UserId)

		// 继续执行后续的处理函数
		c.Next()
	}
}
