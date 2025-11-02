package api

import (
	"net/http"
	"stock/internal/middleware"
	"stock/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, h *Handler) {
	// 创建权限中间件
	permissionMiddleware := NewPermissionMiddleware(h.db)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证模块（不需要权限验证）
		setupAuthRoutes(v1, h)

		// 用户管理模块（需要权限验证）
		setupUserRoutes(v1, h, permissionMiddleware)

		// 角色管理模块（需要权限验证）
		setupRoleRoutes(v1, h, permissionMiddleware)

		// 股票数据模块（需要权限验证）
		setupStockRoutes(v1, h, permissionMiddleware)

		// 策略管理模块（需要权限验证）
		setupStrategyRoutes(v1, h, permissionMiddleware)

		// 选股结果模块（需要权限验证）
		setupSelectionRoutes(v1, h, permissionMiddleware)

		// 数据采集模块（需要权限验证）
		setupCollectorRoutes(v1, h, permissionMiddleware)

		// 通知管理模块（需要权限验证）
		setupNotificationRoutes(v1, h, permissionMiddleware)

		// 投资组合模块（需要权限验证）
		setupPortfolioRoutes(v1, h, permissionMiddleware)

		// 报表分析模块（需要权限验证）
		setupReportRoutes(v1, h, permissionMiddleware)

		// 系统监控模块（需要权限验证）
		setupMonitoringRoutes(v1, h, permissionMiddleware)

		// 系统设置模块（需要权限验证）
		setupSettingsRoutes(v1, h, permissionMiddleware)
	}
}

// setupAuthRoutes 认证相关路由
func setupAuthRoutes(rg *gin.RouterGroup, h *Handler) {
	// 创建认证服务和处理器
	authService := service.NewAuthService(h.db, "your-jwt-secret-key")
	authHandler := NewAuthHandler(authService)

	auth := rg.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/profile", authHandler.GetProfile)
		auth.PUT("/profile", authHandler.UpdateProfile)
		auth.POST("/change-password", authHandler.ChangePassword)
		auth.GET("/login-logs", authHandler.GetLoginLogs)
	}

	// 将authHandler存储到全局变量或传递给其他函数
	setupPermissionRoutes(rg, h, authHandler)
}

// setupPermissionRoutes 权限相关路由
func setupPermissionRoutes(rg *gin.RouterGroup, h *Handler, authHandler *AuthHandler) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	// 权限管理
	permissions := rg.Group("/permissions")
	permissions.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		permissions.GET("", h.GetPermissions)
		permissions.POST("", h.CreatePermission)
		permissions.POST("/reset", h.ResetPermissions)
		permissions.GET("/:id", h.GetPermission)
		permissions.PUT("/:id", h.UpdatePermission)
		permissions.DELETE("/:id", h.DeletePermission)
		permissions.GET("/user/menu", authHandler.GetUserMenuPermissions) // 使用认证处理器
	}
}

// setupUserRoutes 用户管理路由
func setupUserRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	users := rg.Group("/users")
	users.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		users.GET("", pm.RequirePermission("system:users"), h.GetUsers)
		users.POST("", pm.RequirePermission("system:users"), h.CreateUser)
		users.GET("/:id", pm.RequirePermission("system:users"), h.GetUser)
		users.PUT("/:id", pm.RequirePermission("system:users"), h.UpdateUser)
		users.DELETE("/:id", pm.RequirePermission("system:users"), h.DeleteUser)
		users.GET("/:id/logs", pm.RequirePermission("system:users"), h.GetUserLogs)
	}
}

// setupRoleRoutes 角色管理路由
func setupRoleRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	roles := rg.Group("/roles")
	roles.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		roles.GET("", pm.RequirePermission("system:roles"), h.GetRoles)
		roles.POST("", pm.RequirePermission("system:roles"), h.CreateRole)
		roles.GET("/:id", pm.RequirePermission("system:roles"), h.GetRole)
		roles.PUT("/:id", pm.RequirePermission("system:roles"), h.UpdateRole)
		roles.DELETE("/:id", pm.RequirePermission("system:roles"), h.DeleteRole)
		roles.GET("/:id/permissions", pm.RequirePermission("system:roles"), h.GetRolePermissions)
		roles.PUT("/:id/permissions", pm.RequirePermission("system:roles"), h.UpdateRolePermissions)
	}
}

// setupStockRoutes 股票数据路由
func setupStockRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	stocks := rg.Group("/stocks")
	stocks.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		stocks.GET("", pm.RequirePermission("stocks:list"), h.GetStockList)
		stocks.GET("/:code", pm.RequirePermission("stocks:list"), h.GetStockDetail)
		stocks.GET("/:code/kline", pm.RequirePermission("stocks:list"), h.GetKLineData)
		stocks.GET("/:code/financial", pm.RequirePermission("stocks:list"), h.GetFinancialData)
		stocks.GET("/:code/realtime", pm.RequirePermission("stocks:realtime"), h.GetRealtimeData)
		stocks.POST("/:code/watch", pm.RequirePermission("stocks:watchlist"), h.AddToWatchlist)
		stocks.DELETE("/:code/watch", pm.RequirePermission("stocks:watchlist"), h.RemoveFromWatchlist)
	}
}

// setupStrategyRoutes 策略管理路由
func setupStrategyRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	strategies := rg.Group("/strategies")
	strategies.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		strategies.GET("", pm.RequirePermission("strategies:list"), h.GetStrategies)
		strategies.POST("", pm.RequirePermission("strategies:create"), h.CreateStrategy)
		strategies.GET("/:id", pm.RequirePermission("strategies:list"), h.GetStrategy)
		strategies.PUT("/:id", pm.RequirePermission("strategies:list"), h.UpdateStrategy)
		strategies.DELETE("/:id", pm.RequirePermission("strategies:list"), h.DeleteStrategy)
		strategies.POST("/:id/backtest", pm.RequirePermission("strategies:backtest"), h.RunBacktest)
		strategies.GET("/:id/results", pm.RequirePermission("strategies:list"), h.GetStrategyResults)
		strategies.POST("/:id/execute", pm.RequirePermission("strategies:list"), h.ExecuteStrategy)
	}
}

// setupSelectionRoutes 选股结果路由
func setupSelectionRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	selections := rg.Group("/selections")
	selections.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		selections.GET("", pm.RequirePermission("selections:list"), h.GetSelections)
		selections.GET("/today", pm.RequirePermission("selections:today"), h.GetTodaySelections)
		selections.GET("/history", pm.RequirePermission("selections:history"), h.GetSelectionHistory)
		selections.POST("", pm.RequirePermission("selections:list"), h.CreateSelection)
		selections.GET("/:id", pm.RequirePermission("selections:list"), h.GetSelection)
		selections.DELETE("/:id", pm.RequirePermission("selections:list"), h.DeleteSelection)
		selections.POST("/export", pm.RequirePermission("selections:list"), h.ExportSelections)
	}
}

// setupCollectorRoutes 数据采集路由
func setupCollectorRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	collectors := rg.Group("/collectors")
	collectors.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		collectors.GET("", pm.RequirePermission("collectors:manage"), h.GetCollectors)
		collectors.GET("/status", pm.RequirePermission("collectors:manage"), h.GetCollectorStatus)
		collectors.POST("/sync", pm.RequirePermission("collectors:sync"), h.SyncData)
		collectors.POST("/sync/:source", pm.RequirePermission("collectors:sync"), h.SyncDataFromSource)
		collectors.GET("/tasks", pm.RequirePermission("collectors:tasks"), h.GetCollectorTasks)
		collectors.POST("/tasks", pm.RequirePermission("collectors:tasks"), h.CreateCollectorTask)
		collectors.PUT("/tasks/:id", pm.RequirePermission("collectors:tasks"), h.UpdateCollectorTask)
		collectors.DELETE("/tasks/:id", pm.RequirePermission("collectors:tasks"), h.DeleteCollectorTask)
	}
}

// setupNotificationRoutes 通知管理路由
func setupNotificationRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	notifications := rg.Group("/notifications")
	notifications.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		// 机器人管理
		robots := notifications.Group("/robots")
		{
			robots.GET("", pm.RequirePermission("notifications:robots"), h.GetRobots)
			robots.POST("", pm.RequirePermission("notifications:robots"), h.CreateRobot)
			robots.GET("/:id", pm.RequirePermission("notifications:robots"), h.GetRobot)
			robots.PUT("/:id", pm.RequirePermission("notifications:robots"), h.UpdateRobot)
			robots.DELETE("/:id", pm.RequirePermission("notifications:robots"), h.DeleteRobot)
			robots.POST("/:id/test", pm.RequirePermission("notifications:robots"), h.TestRobot)
		}

		// 消息发送
		notifications.POST("/send", pm.RequirePermission("notifications:robots"), h.SendNotification)
		notifications.GET("/logs", pm.RequirePermission("notifications:logs"), h.GetNotificationLogs)
		notifications.GET("/templates", pm.RequirePermission("notifications:templates"), h.GetMessageTemplates)
		notifications.POST("/templates", pm.RequirePermission("notifications:templates"), h.CreateMessageTemplate)
	}
}

// setupPortfolioRoutes 投资组合路由
func setupPortfolioRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	portfolios := rg.Group("/portfolios")
	portfolios.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		portfolios.GET("", pm.RequirePermission("portfolios:manage"), h.GetPortfolios)
		portfolios.POST("", pm.RequirePermission("portfolios:manage"), h.CreatePortfolio)
		portfolios.GET("/:id", pm.RequirePermission("portfolios:manage"), h.GetPortfolio)
		portfolios.PUT("/:id", pm.RequirePermission("portfolios:manage"), h.UpdatePortfolio)
		portfolios.DELETE("/:id", pm.RequirePermission("portfolios:manage"), h.DeletePortfolio)
		portfolios.GET("/:id/performance", pm.RequirePermission("portfolios:performance"), h.GetPortfolioPerformance)
		portfolios.POST("/:id/stocks", pm.RequirePermission("portfolios:manage"), h.AddStockToPortfolio)
		portfolios.DELETE("/:id/stocks/:stock_id", pm.RequirePermission("portfolios:manage"), h.RemoveStockFromPortfolio)
	}
}

// setupReportRoutes 报表分析路由
func setupReportRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	reports := rg.Group("/reports")
	reports.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		reports.GET("/strategy-performance", pm.RequirePermission("reports:strategy"), h.GetStrategyPerformanceReport)
		reports.GET("/market-analysis", pm.RequirePermission("reports:market"), h.GetMarketAnalysisReport)
		reports.GET("/selection-success", pm.RequirePermission("reports:strategy"), h.GetSelectionSuccessReport)
		reports.GET("/risk-analysis", pm.RequirePermission("reports:risk"), h.GetRiskAnalysisReport)
		reports.POST("/custom", pm.RequirePermission("reports:strategy"), h.GenerateCustomReport)
		reports.GET("/dashboard", pm.RequirePermission("dashboard:view"), h.GetDashboardData)
	}
}

// setupMonitoringRoutes 系统监控路由
func setupMonitoringRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	monitoring := rg.Group("/monitoring")
	monitoring.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		monitoring.GET("/system", pm.RequirePermission("system:monitoring"), h.GetSystemStatus)
		monitoring.GET("/api-stats", pm.RequirePermission("system:monitoring"), h.GetAPIStats)
		monitoring.GET("/logs", pm.RequirePermission("system:monitoring"), h.GetSystemLogs)
		monitoring.GET("/performance", pm.RequirePermission("system:monitoring"), h.GetPerformanceMetrics)
		monitoring.GET("/errors", pm.RequirePermission("system:monitoring"), h.GetErrorLogs)
		monitoring.GET("/alerts", pm.RequirePermission("system:monitoring"), h.GetAlerts)
	}
}

// setupSettingsRoutes 系统设置路由
func setupSettingsRoutes(rg *gin.RouterGroup, h *Handler, pm *PermissionMiddleware) {
	// 创建JWT认证中间件
	authMiddleware := middleware.NewAuthMiddleware(h.db, "your-jwt-secret-key")

	settings := rg.Group("/settings")
	settings.Use(authMiddleware.RequireAuth()) // 添加JWT认证中间件
	{
		settings.GET("", pm.RequirePermission("system:settings"), h.GetSettings)
		settings.PUT("", pm.RequirePermission("system:settings"), h.UpdateSettings)
		settings.GET("/backup", pm.RequirePermission("system:settings"), h.BackupSettings)
		settings.POST("/restore", pm.RequirePermission("system:settings"), h.RestoreSettings)
	}
}
