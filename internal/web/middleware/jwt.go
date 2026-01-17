package middleware

import (
	"crypto/sha256"
	"encoding/hex"
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

// jwtSecretKey Access Token 签名密钥
// 用于生成和验证短期 Access Token 的签名
var jwtSecretKey []byte

// refreshSecretKey Refresh Token 签名密钥
// 用于生成和验证长期 Refresh Token 的签名
// 使用不同的密钥增强安全性，即使 Access Token 密钥泄露，Refresh Token 仍然安全
var refreshSecretKey []byte

// ============================================================================
// 步骤1: 初始化密钥
// ============================================================================

// InitJWT 初始化 JWT 密钥
// 必须在程序启动时调用，设置用于签名和验证的密钥
// 参数:
//   - accessKey: Access Token 密钥（短期 Token）
//   - refreshKey: Refresh Token 密钥（长期 Token），可选，为空则使用 accessKey + "_refresh"
//
// 使用示例: middleware.InitJWT("your-access-secret", "your-refresh-secret")
func InitJWT(accessKey string, refreshKey ...string) {
	jwtSecretKey = []byte(accessKey)
	if len(refreshKey) > 0 && refreshKey[0] != "" {
		refreshSecretKey = []byte(refreshKey[0])
	} else {
		refreshSecretKey = []byte(accessKey + "_refresh")
	}
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

	// UserAgent 用户代理哈希值（安全增强）
	// 将 Token 绑定到特定的浏览器/设备，防止 Token 被盗用到其他设备
	// 存储的是 User-Agent 的 SHA256 哈希值，而非原始值
	UserAgent string `json:"userAgent"`

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

// hashUserAgent 计算 User-Agent 的 SHA256 哈希值
// 使用哈希而非原始值，可以：
//  1. 减少 Token 体积（User-Agent 可能很长）
//  2. 避免在 Token 中暴露用户的浏览器信息
func hashUserAgent(userAgent string) string {
	hash := sha256.Sum256([]byte(userAgent))
	return hex.EncodeToString(hash[:])
}

// GenerateToken 生成 JWT Token
// 在用户登录验证成功后调用此函数，生成一个包含用户信息的 Token
//
// 参数:
//   - userId:     用户 ID，将被编码到 Token 中
//   - userAgent:  用户的 User-Agent，用于绑定设备（安全增强）
//   - expireTime: Token 有效期，如 24*time.Hour 表示 24 小时后过期
//
// 返回:
//   - string: 生成的 Token 字符串，格式为 "xxxxx.yyyyy.zzzzz"
//   - error:  如果签名失败则返回错误
//
// 使用示例:
//
//	token, err := middleware.GenerateToken(user.Id, c.GetHeader("User-Agent"), 24*time.Hour)
//	c.Header("x-jwt-token", token) // 返回给前端
func GenerateToken(userId int64, userAgent string, expireTime time.Duration) (string, error) {
	// 第一步：构造 Payload（Claims）
	// 创建包含用户信息和标准字段的 Claims 结构
	claims := UserClaims{
		UserId:    userId,                   // 存入用户 ID
		UserAgent: hashUserAgent(userAgent), // 存入 User-Agent 哈希值（安全绑定）
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
// 长短 Token 机制（Access Token + Refresh Token）
// ============================================================================
// Access Token:  短期有效（如 30 分钟），用于 API 访问
// Refresh Token: 长期有效（如 7 天），用于刷新 Access Token
//
// 使用流程：
//   1. 用户登录 → 返回 Access Token + Refresh Token
//   2. 用 Access Token 访问 API
//   3. Access Token 过期 → 用 Refresh Token 获取新的 Access Token
//   4. Refresh Token 过期 → 用户需要重新登录
// ============================================================================

// RefreshClaims Refresh Token 的 Claims
// 相比 Access Token，Refresh Token 只需要存储 userId，不需要 UserAgent
// 因为 Refresh Token 只用于换取新的 Access Token，不用于 API 访问
type RefreshClaims struct {
	UserId int64  `json:"userId"`
	SSid   string `json:"ssid"` // Session ID，用于黑名单标识
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 Access Token（短期 Token）
// 封装 GenerateToken，用于长短 Token 机制
func GenerateAccessToken(userId int64, userAgent string, expireTime time.Duration) (string, error) {
	return GenerateToken(userId, userAgent, expireTime)
}

// GenerateRefreshToken 生成 Refresh Token（长期 Token）
// Refresh Token 使用独立的密钥签名，有效期较长（如 7 天）
//
// 参数:
//   - userId: 用户 ID
//   - ssid: Session ID，用于标识此次登录会话，退出时用于加入黑名单
//   - expireTime: 有效期，建议 7 天或更长
//
// 返回:
//   - string: Refresh Token 字符串
//   - error: 签名错误
func GenerateRefreshToken(userId int64, ssid string, expireTime time.Duration) (string, error) {
	claims := RefreshClaims{
		UserId: userId,
		SSid:   ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecretKey)
}

// GenerateTokenPair 生成 Token 对（Access Token + Refresh Token）
// 登录成功后调用此函数，一次性生成两个 Token
//
// 参数:
//   - userId: 用户 ID
//   - userAgent: 用户 User-Agent（用于 Access Token 设备绑定）
//   - ssid: Session ID，用于标识此次登录会话
//   - accessExpire: Access Token 有效期（建议 30 分钟）
//   - refreshExpire: Refresh Token 有效期（建议 7 天）
//
// 返回:
//   - accessToken: 短期 Token，用于 API 访问
//   - refreshToken: 长期 Token，用于刷新 Access Token
//   - error: 生成错误
func GenerateTokenPair(userId int64, userAgent, ssid string, accessExpire, refreshExpire time.Duration) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userId, userAgent, accessExpire)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = GenerateRefreshToken(userId, ssid, refreshExpire)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ParseRefreshToken 解析并验证 Refresh Token
// 用于刷新接口，验证 Refresh Token 是否有效
//
// 参数:
//   - tokenString: Refresh Token 字符串
//
// 返回:
//   - *RefreshClaims: 解析出的 Claims（包含 userId）
//   - error: 解析或验证错误（过期、签名无效等）
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return refreshSecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
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
		// 第六步：验证 User-Agent（安全增强）
		// ========================================
		// 检查当前请求的 User-Agent 是否与 Token 中绑定的一致
		// 如果不一致，说明 Token 可能被盗用到了其他设备
		currentUA := hashUserAgent(c.GetHeader("User-Agent"))
		if claims.UserAgent != "" && claims.UserAgent != currentUA {
			// User-Agent 不匹配，可能是 Token 被盗用
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// ========================================
		// 第七步：将用户信息存入 Context
		// ========================================
		// 验证通过后，将 claims 中的 userId 存入 Gin 的 Context
		// 后续的处理函数可以通过 c.Get("userId") 或 c.GetInt64("userId") 获取
		// 这样就实现了"无状态"认证：无需查询数据库或 Session，直接从 Token 中获取用户信息
		c.Set("userId", claims.UserId)

		// 继续执行后续的处理函数
		c.Next()
	}
}
