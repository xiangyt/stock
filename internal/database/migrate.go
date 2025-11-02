package database

import (
	"fmt"
	"stock/internal/model"
)

// MigrateUserSystem 迁移用户权限系统相关表
func (d *Database) MigrateUserSystem() error {
	d.logger.Info("开始迁移用户权限系统表...")

	// 按依赖顺序迁移用户权限相关模型
	userModels := []interface{}{
		&model.Role{},             // 角色表
		&model.Permission{},       // 权限表
		&model.RolePermission{},   // 角色权限关联表
		&model.User{},             // 用户表（更新外键）
		&model.UserLoginLog{},     // 用户登录日志表
		&model.UserOperationLog{}, // 用户操作日志表
		&model.JWTBlacklist{},     // JWT黑名单表
	}

	for _, model := range userModels {
		if err := d.DB.AutoMigrate(model); err != nil {
			d.logger.Errorf("迁移用户系统模型 %T 失败: %v", model, err)
			return fmt.Errorf("迁移用户系统模型 %T 失败: %v", model, err)
		}
		d.logger.Infof("成功迁移模型: %T", model)
	}

	d.logger.Info("用户权限系统表迁移完成")
	return nil
}

// InitializeDefaultData 初始化默认数据
func (d *Database) InitializeDefaultData() error {
	d.logger.Info("开始初始化默认数据...")

	// 检查是否已经初始化过
	var count int64
	if err := d.DB.Model(&model.Role{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查角色数据失败: %v", err)
	}

	if count > 0 {
		d.logger.Info("默认数据已存在，跳过初始化")
		return nil
	}

	// 创建默认角色
	roles := []model.Role{
		{
			RoleName:    "超级管理员",
			RoleCode:    "super_admin",
			Description: stringPtr("系统超级管理员，拥有所有权限"),
			IsSystem:    true,
			Status:      1,
			SortOrder:   1,
			CreatedBy:   "system",
			UpdatedBy:   "system",
		},
		{
			RoleName:    "管理员",
			RoleCode:    "admin",
			Description: stringPtr("系统管理员，拥有大部分管理权限"),
			IsSystem:    true,
			Status:      1,
			SortOrder:   2,
			CreatedBy:   "system",
			UpdatedBy:   "system",
		},
		{
			RoleName:    "普通用户",
			RoleCode:    "user",
			Description: stringPtr("普通用户，拥有基础功能权限"),
			IsSystem:    true,
			Status:      1,
			SortOrder:   3,
			CreatedBy:   "system",
			UpdatedBy:   "system",
		},
		{
			RoleName:    "访客",
			RoleCode:    "guest",
			Description: stringPtr("访客用户，只能查看基础信息"),
			IsSystem:    true,
			Status:      1,
			SortOrder:   4,
			CreatedBy:   "system",
			UpdatedBy:   "system",
		},
	}

	if err := d.DB.Create(&roles).Error; err != nil {
		return fmt.Errorf("创建默认角色失败: %v", err)
	}
	d.logger.Info("默认角色创建成功")

	// 创建默认权限
	permissions := []model.Permission{
		// 仪表板
		{PermissionName: "仪表板", PermissionCode: "dashboard:view", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/"), Description: stringPtr("仪表板访问权限"), SortOrder: 1, Status: 1},

		// 股票数据模块
		{PermissionName: "股票数据模块", PermissionCode: "stocks:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/stocks"), Description: stringPtr("股票数据模块权限"), SortOrder: 2, Status: 1},
		{PermissionName: "股票列表", PermissionCode: "stocks:list", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/stocks"), Description: stringPtr("股票列表权限"), SortOrder: 21, Status: 1},
		{PermissionName: "自选股", PermissionCode: "stocks:watchlist", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/stocks/watchlist"), Description: stringPtr("自选股管理权限"), SortOrder: 22, Status: 1},
		{PermissionName: "实时行情", PermissionCode: "stocks:realtime", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/stocks/realtime"), Description: stringPtr("实时行情权限"), SortOrder: 23, Status: 1},

		// 策略管理模块
		{PermissionName: "策略管理模块", PermissionCode: "strategies:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/strategies"), Description: stringPtr("策略管理模块权限"), SortOrder: 3, Status: 1},
		{PermissionName: "策略列表", PermissionCode: "strategies:list", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/strategies"), Description: stringPtr("策略列表权限"), SortOrder: 31, Status: 1},
		{PermissionName: "创建策略", PermissionCode: "strategies:create", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/strategies/create"), Description: stringPtr("创建策略权限"), SortOrder: 32, Status: 1},
		{PermissionName: "回测中心", PermissionCode: "strategies:backtest", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/strategies/backtest"), Description: stringPtr("回测中心权限"), SortOrder: 33, Status: 1},

		// 选股结果模块
		{PermissionName: "选股结果模块", PermissionCode: "selections:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/selections"), Description: stringPtr("选股结果模块权限"), SortOrder: 4, Status: 1},
		{PermissionName: "选股记录", PermissionCode: "selections:list", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/selections"), Description: stringPtr("选股记录权限"), SortOrder: 41, Status: 1},
		{PermissionName: "今日选股", PermissionCode: "selections:today", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/selections/today"), Description: stringPtr("今日选股权限"), SortOrder: 42, Status: 1},
		{PermissionName: "历史回顾", PermissionCode: "selections:history", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/selections/history"), Description: stringPtr("历史回顾权限"), SortOrder: 43, Status: 1},

		// 投资组合模块
		{PermissionName: "投资组合模块", PermissionCode: "portfolios:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/portfolios"), Description: stringPtr("投资组合模块权限"), SortOrder: 5, Status: 1},
		{PermissionName: "组合管理", PermissionCode: "portfolios:manage", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/portfolios"), Description: stringPtr("组合管理权限"), SortOrder: 51, Status: 1},
		{PermissionName: "业绩分析", PermissionCode: "portfolios:performance", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/portfolios/performance"), Description: stringPtr("业绩分析权限"), SortOrder: 52, Status: 1},

		// 数据采集模块
		{PermissionName: "数据采集模块", PermissionCode: "collectors:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/collectors"), Description: stringPtr("数据采集模块权限"), SortOrder: 6, Status: 1},
		{PermissionName: "采集器管理", PermissionCode: "collectors:manage", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/collectors"), Description: stringPtr("采集器管理权限"), SortOrder: 61, Status: 1},
		{PermissionName: "采集任务", PermissionCode: "collectors:tasks", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/collectors/tasks"), Description: stringPtr("采集任务权限"), SortOrder: 62, Status: 1},
		{PermissionName: "数据同步", PermissionCode: "collectors:sync", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/collectors/sync"), Description: stringPtr("数据同步权限"), SortOrder: 63, Status: 1},

		// 通知管理模块
		{PermissionName: "通知管理模块", PermissionCode: "notifications:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/notifications"), Description: stringPtr("通知管理模块权限"), SortOrder: 7, Status: 1},
		{PermissionName: "机器人配置", PermissionCode: "notifications:robots", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/notifications/robots"), Description: stringPtr("机器人配置权限"), SortOrder: 71, Status: 1},
		{PermissionName: "消息模板", PermissionCode: "notifications:templates", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/notifications/templates"), Description: stringPtr("消息模板权限"), SortOrder: 72, Status: 1},
		{PermissionName: "发送日志", PermissionCode: "notifications:logs", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/notifications/logs"), Description: stringPtr("发送日志权限"), SortOrder: 73, Status: 1},

		// 报表分析模块
		{PermissionName: "报表分析模块", PermissionCode: "reports:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/reports"), Description: stringPtr("报表分析模块权限"), SortOrder: 8, Status: 1},
		{PermissionName: "策略报告", PermissionCode: "reports:strategy", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/reports/strategy"), Description: stringPtr("策略报告权限"), SortOrder: 81, Status: 1},
		{PermissionName: "市场分析", PermissionCode: "reports:market", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/reports/market"), Description: stringPtr("市场分析权限"), SortOrder: 82, Status: 1},
		{PermissionName: "风险分析", PermissionCode: "reports:risk", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/reports/risk"), Description: stringPtr("风险分析权限"), SortOrder: 83, Status: 1},

		// 系统管理模块
		{PermissionName: "系统管理模块", PermissionCode: "system:module", ResourceType: stringPtr("module"), ResourcePath: stringPtr("/system"), Description: stringPtr("系统管理模块权限"), SortOrder: 9, Status: 1},
		{PermissionName: "用户管理", PermissionCode: "system:users", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/system/users"), Description: stringPtr("用户管理权限"), SortOrder: 91, Status: 1},
		{PermissionName: "角色管理", PermissionCode: "system:roles", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/system/roles"), Description: stringPtr("角色管理权限"), SortOrder: 92, Status: 1},
		{PermissionName: "系统监控", PermissionCode: "system:monitoring", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/system/monitoring"), Description: stringPtr("系统监控权限"), SortOrder: 93, Status: 1},
		{PermissionName: "系统设置", PermissionCode: "system:settings", ResourceType: stringPtr("menu"), ResourcePath: stringPtr("/system/settings"), Description: stringPtr("系统设置权限"), SortOrder: 94, Status: 1},
	}

	if err := d.DB.Create(&permissions).Error; err != nil {
		return fmt.Errorf("创建默认权限失败: %v", err)
	}
	d.logger.Info("默认权限创建成功")

	// 为超级管理员分配所有权限
	var allPermissions []model.Permission
	if err := d.DB.Find(&allPermissions).Error; err != nil {
		return fmt.Errorf("获取权限列表失败: %v", err)
	}

	var rolePermissions []model.RolePermission
	for _, permission := range allPermissions {
		rolePermissions = append(rolePermissions, model.RolePermission{
			RoleID:       1, // 超级管理员角色ID
			PermissionID: permission.ID,
			CreatedBy:    "system",
		})
	}

	if err := d.DB.Create(&rolePermissions).Error; err != nil {
		return fmt.Errorf("分配超级管理员权限失败: %v", err)
	}

	// 创建默认管理员用户
	adminUser := model.User{
		Username:     "admin",
		Email:        "admin@stock.com",
		PasswordHash: "$2a$10$N.zmdr9k7uOCQb376NoUnuTGJmKt6qPxgjjxJBPLjOLBbHv5c.gLa", // 密码: admin123
		RealName:     stringPtr("系统管理员"),
		Status:       1,
		RoleID:       uintPtr(1), // 超级管理员角色
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}

	if err := d.DB.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("创建默认管理员用户失败: %v", err)
	}

	d.logger.Info("默认数据初始化完成")
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}
