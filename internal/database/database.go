package database

import (
	"fmt"

	"stock/internal/config"
	"stock/internal/model"
	"stock/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	DB     *gorm.DB
	config *config.DatabaseConfig
	logger *utils.Logger
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *config.DatabaseConfig, log *utils.Logger) (*Database, error) {
	// 构建MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	// 配置GORM日志级别
	var gormLogLevel logger.LogLevel
	switch log.Level.String() {
	case "debug":
		gormLogLevel = logger.Info
	case "info":
		gormLogLevel = logger.Warn
	default:
		gormLogLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Info("Successfully connected to MySQL database")

	return &Database{
		DB:     db,
		config: cfg,
		logger: log,
	}, nil
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate() error {
	d.logger.Info("Starting database migration...")

	// 按依赖顺序迁移模型，避免外键约束问题
	models := []interface{}{
		&model.Stock{},              // 基础表，无外键依赖
		&model.DailyData{},          // 依赖Stock
		&model.WeeklyData{},         // 依赖Stock
		&model.MonthlyData{},        // 依赖Stock
		&model.YearlyData{},         // 依赖Stock
		&model.PerformanceReport{},  // 依赖Stock
		&model.ShareholderCount{},   // 依赖Stock
		&model.TechnicalIndicator{}, // 依赖Stock
		&model.Strategy{},           // 独立表
		&model.Portfolio{},          // 独立表
		&model.StrategyResult{},     // 依赖Strategy和Stock
		&model.PortfolioStock{},     // 依赖Portfolio和Stock
		&model.BacktestResult{},     // 依赖Strategy
	}

	for _, model := range models {
		if err := d.DB.AutoMigrate(model); err != nil {
			d.logger.Errorf("Failed to migrate model %T: %v", model, err)
			return fmt.Errorf("failed to migrate model %T: %v", model, err)
		}
	}

	d.logger.Info("Database migration completed successfully")
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取GORM数据库实例
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// HealthCheck 健康检查
func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats 获取数据库连接统计信息
func (d *Database) GetStats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
