package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/internal/model"
	"stock/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleCode string `json:"role_code"`
	jwt.RegisteredClaims
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	db        *gorm.DB
	jwtSecret []byte
	userRepo  *repository.UserRepository
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(db *gorm.DB, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		db:        db,
		jwtSecret: []byte(jwtSecret),
		userRepo:  repository.NewUserRepository(db),
	}
}

// RequireAuth JWT认证中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("JWT中间件被调用: %s %s\n", c.Request.Method, c.Request.URL.Path)

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		fmt.Printf("Authorization头: %s\n", authHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		// 解析JWT令牌
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌解析失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌验证失败",
			})
			c.Abort()
			return
		}

		// 检查令牌是否在黑名单中
		if m.isTokenBlacklisted(claims.ID) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌已失效",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.IsActive() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户已被禁用",
			})
			c.Abort()
			return
		}

		if user.IsLocked() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户已被锁定",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", user.ID)
		c.Set("user", user)
		c.Set("username", user.Username)

		// 获取角色信息
		roleCode := ""
		if user.RoleID != nil && *user.RoleID > 0 {
			roleRepo := repository.NewRoleRepository(m.db)
			if role, err := roleRepo.GetByID(*user.RoleID); err == nil {
				roleCode = role.RoleCode
			}
		}
		c.Set("roleCode", roleCode)

		// 调试日志
		fmt.Printf("JWT验证成功，用户ID: %d, 用户名: %s\n", user.ID, user.Username)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（用于某些可以匿名访问的接口）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil {
			c.Next()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.Next()
			return
		}

		if m.isTokenBlacklisted(claims.ID) {
			c.Next()
			return
		}

		user, err := m.userRepo.GetByIDWithRole(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		if !user.IsActive() || user.IsLocked() {
			c.Next()
			return
		}

		c.Set("userID", user.ID)
		c.Set("user", user)
		c.Set("username", user.Username)

		// 获取角色信息
		roleCode := ""
		if user.RoleID != nil && *user.RoleID > 0 {
			roleRepo := repository.NewRoleRepository(m.db)
			if role, err := roleRepo.GetByID(*user.RoleID); err == nil {
				roleCode = role.RoleCode
			}
		}
		c.Set("roleCode", roleCode)

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func (m *AuthMiddleware) GenerateToken(user *model.User) (string, error) {
	now := time.Now()

	// 获取角色信息
	roleCode := ""
	if user.RoleID != nil && *user.RoleID > 0 {
		roleRepo := repository.NewRoleRepository(m.db)
		if role, err := roleRepo.GetByID(*user.RoleID); err == nil {
			roleCode = role.RoleCode
		}
	}

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RoleCode: roleCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateTokenID(),
			Subject:   strconv.Itoa(int(user.ID)),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24小时过期
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "stock-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecret)
}

// RevokeToken 撤销令牌（加入黑名单）
func (m *AuthMiddleware) RevokeToken(tokenString, reason string) error {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return jwt.ErrInvalidKey
	}

	blacklistEntry := &model.JWTBlacklist{
		TokenID:      claims.ID,
		UserID:       claims.UserID,
		TokenHash:    generateTokenHash(tokenString),
		ExpiresAt:    claims.ExpiresAt.Time,
		RevokeReason: &reason,
	}

	return m.db.Create(blacklistEntry).Error
}

// isTokenBlacklisted 检查令牌是否在黑名单中
func (m *AuthMiddleware) isTokenBlacklisted(tokenID string) bool {
	var count int64
	m.db.Model(&model.JWTBlacklist{}).Where("token_id = ?", tokenID).Count(&count)
	return count > 0
}

// generateTokenID 生成令牌ID
func generateTokenID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// generateTokenHash 生成令牌哈希
func generateTokenHash(token string) string {
	// 这里可以使用更复杂的哈希算法
	return token[:32] // 简单截取前32位作为哈希
}
