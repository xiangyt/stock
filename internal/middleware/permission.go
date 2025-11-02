package middleware

import (
	"net/http"
	"stock/internal/model"
	"stock/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionMiddleware 权限验证中间件
type PermissionMiddleware struct {
	db       *gorm.DB
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{
		db:       db,
		userRepo: repository.NewUserRepository(db),
		roleRepo: repository.NewRoleRepository(db),
	}
}

// RequirePermission 权限验证中间件
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户信息获取失败",
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

// RequireAnyPermission 检查用户是否拥有任意一个权限
func (m *PermissionMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户信息获取失败",
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

// hasPermission 检查用户是否拥有指定权限
func (m *PermissionMiddleware) hasPermission(user *model.User, permissionCode string) bool {
	// 检查用户是否有角色
	if user.RoleID == nil || *user.RoleID == 0 {
		return false
	}

	// 获取角色信息
	role, err := m.roleRepo.GetByID(*user.RoleID)
	if err != nil || role.Status != 1 {
		return false
	}

	// 如果是管理员，拥有所有权限
	if role.RoleCode == "admin" || role.RoleCode == "super_admin" {
		return true
	}

	// 获取角色权限
	permissions, err := m.roleRepo.GetRolePermissions(*user.RoleID)
	if err != nil {
		return false
	}

	// 检查用户角色权限
	for _, permission := range permissions {
		if permission.PermissionCode == permissionCode && permission.Status == 1 {
			return true
		}
	}

	return false
}

// GetUserPermissions 获取用户权限列表
func (m *PermissionMiddleware) GetUserPermissions(userID uint) ([]string, error) {
	user, err := m.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	var permissions []string
	if user.RoleID != nil {
		// 通过角色仓储获取角色权限
		rolePermissions, err := m.roleRepo.GetRolePermissions(*user.RoleID)
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

// GetUserMenuPermissions 获取用户菜单权限
func (m *PermissionMiddleware) GetUserMenuPermissions(userID uint) ([]model.Permission, error) {
	user, err := m.userRepo.GetByID(userID)
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

// RoutePermissionMap 路由权限映射
var RoutePermissionMap = map[string]string{
	// 股票数据
	"/stocks":           "stocks:list",
	"/stocks/watchlist": "stocks:watchlist",
	"/stocks/realtime":  "stocks:realtime",

	// 策略管理
	"/strategies":          "strategies:list",
	"/strategies/create":   "strategies:create",
	"/strategies/backtest": "strategies:backtest",

	// 选股结果
	"/selections":         "selections:list",
	"/selections/today":   "selections:today",
	"/selections/history": "selections:history",

	// 投资组合
	"/portfolios":             "portfolios:manage",
	"/portfolios/performance": "portfolios:performance",

	// 数据采集
	"/collectors":       "collectors:manage",
	"/collectors/tasks": "collectors:tasks",
	"/collectors/sync":  "collectors:sync",

	// 通知管理
	"/notifications/robots":    "notifications:robots",
	"/notifications/templates": "notifications:templates",
	"/notifications/logs":      "notifications:logs",

	// 报表分析
	"/reports/strategy": "reports:strategy",
	"/reports/market":   "reports:market",
	"/reports/risk":     "reports:risk",

	// 系统管理
	"/system/users":      "system:users",
	"/system/roles":      "system:roles",
	"/system/monitoring": "system:monitoring",
	"/system/settings":   "system:settings",
}

// RequireRoutePermission 根据路由自动检查权限
func (m *PermissionMiddleware) RequireRoutePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 查找对应的权限编码
		permissionCode, exists := RoutePermissionMap[path]
		if !exists {
			// 如果路由没有配置权限要求，直接通过
			c.Next()
			return
		}

		// 获取用户信息
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证用户",
			})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户信息获取失败",
			})
			c.Abort()
			return
		}

		// 检查权限
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
