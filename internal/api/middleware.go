package api

import (
	"net/http"
	"strings"

	"stock/internal/model"
	"stock/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionMiddleware 权限验证中间件
type PermissionMiddleware struct {
	db       *gorm.DB
	roleRepo *repository.RoleRepository
}

// NewPermissionMiddleware 创建权限中间件实例
func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{
		db:       db,
		roleRepo: repository.NewRoleRepository(db),
	}
}

// RequirePermission 权限验证中间件
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取当前用户（这里需要根据实际的认证方式获取）
		// 这里假设用户信息已经存储在上下文中
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权访问",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		userRepo := repository.NewUserRepository(m.db)
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户是否拥有权限
		if !m.hasPermission(user, permissionCode) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission 检查用户是否拥有指定权限
func (m *PermissionMiddleware) hasPermission(user *model.User, permissionCode string) bool {
	if user.RoleID == nil {
		return false
	}

	// 获取用户角色
	roleRepo := repository.NewRoleRepository(m.db)
	role, err := roleRepo.GetByID(*user.RoleID)
	if err != nil {
		return false
	}

	// 如果是超级管理员，拥有所有权限
	if role.RoleCode == "super_admin" {
		return true
	}

	// 获取角色权限
	permissions, err := roleRepo.GetRolePermissions(*user.RoleID)
	if err != nil {
		return false
	}

	// 检查是否拥有指定权限
	for _, permission := range permissions {
		if permission.PermissionCode == permissionCode && permission.Status == model.PermissionStatusActive {
			return true
		}
	}

	return false
}

// RequireAnyPermission 检查用户是否拥有任意一个权限
func (m *PermissionMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权访问",
			})
			c.Abort()
			return
		}

		userRepo := repository.NewUserRepository(m.db)
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户是否拥有任意一个权限
		for _, permissionCode := range permissionCodes {
			if m.hasPermission(user, permissionCode) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		c.Abort()
	}
}

// GetUserPermissions 获取用户的所有权限
func (m *PermissionMiddleware) GetUserPermissions(userID uint) ([]string, error) {
	userRepo := repository.NewUserRepository(m.db)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	var permissions []string
	if user.RoleID != nil {
		roleRepo := repository.NewRoleRepository(m.db)
		rolePermissions, err := roleRepo.GetRolePermissions(*user.RoleID)
		if err != nil {
			return nil, err
		}

		for _, permission := range rolePermissions {
			if permission.Status == model.PermissionStatusActive {
				permissions = append(permissions, permission.PermissionCode)
			}
		}
	}

	return permissions, nil
}

// GetUserMenuPermissions 获取用户的菜单权限
func (m *PermissionMiddleware) GetUserMenuPermissions(userID uint) ([]model.Permission, error) {
	userRepo := repository.NewUserRepository(m.db)
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	var menuPermissions []model.Permission
	if user.RoleID != nil && *user.RoleID > 0 {
		// 获取角色权限
		permissions, err := m.roleRepo.GetRolePermissions(*user.RoleID)
		if err != nil {
			return nil, err
		}

		for _, permission := range permissions {
			// 只返回菜单类型的权限
			if permission.Status == model.PermissionStatusActive &&
				permission.ResourceType != nil &&
				(*permission.ResourceType == "menu" || *permission.ResourceType == "module") {
				menuPermissions = append(menuPermissions, permission)
			}
		}
	}

	return menuPermissions, nil
}

// RoutePermissionMap 路由权限映射表
var RoutePermissionMap = map[string]string{
	// 仪表板
	"/": "dashboard:view",

	// 股票数据模块
	"/stocks":           "stocks:list",
	"/stocks/watchlist": "stocks:watchlist",
	"/stocks/realtime":  "stocks:realtime",
	"/stock/:code":      "stocks:detail",

	// 策略管理模块
	"/strategies":          "strategies:list",
	"/strategies/create":   "strategies:create",
	"/strategies/backtest": "strategies:backtest",

	// 选股结果模块
	"/selections":         "selections:list",
	"/selections/today":   "selections:today",
	"/selections/history": "selections:history",

	// 投资组合模块
	"/portfolios":             "portfolios:manage",
	"/portfolios/performance": "portfolios:performance",

	// 数据采集模块
	"/collectors":       "collectors:manage",
	"/collectors/tasks": "collectors:tasks",
	"/collectors/sync":  "collectors:sync",

	// 通知管理模块
	"/notifications/robots":    "notifications:robots",
	"/notifications/templates": "notifications:templates",
	"/notifications/logs":      "notifications:logs",

	// 报表分析模块
	"/reports/strategy": "reports:strategy",
	"/reports/market":   "reports:market",
	"/reports/risk":     "reports:risk",

	// 系统管理模块
	"/system/users":      "system:users",
	"/system/roles":      "system:roles",
	"/system/monitoring": "system:monitoring",
	"/system/settings":   "system:settings",
}

// AutoPermissionCheck 自动权限检查中间件
func (m *PermissionMiddleware) AutoPermissionCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径
		path := c.Request.URL.Path

		// 移除API前缀
		path = strings.TrimPrefix(path, "/api/v1")

		// 查找对应的权限编码
		permissionCode, exists := RoutePermissionMap[path]
		if !exists {
			// 如果没有配置权限，默认允许访问
			c.Next()
			return
		}

		// 检查权限
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权访问",
			})
			c.Abort()
			return
		}

		userRepo := repository.NewUserRepository(m.db)
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		if !m.hasPermission(user, permissionCode) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
