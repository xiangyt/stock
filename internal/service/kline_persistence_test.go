package service

import (
	"os"
	"testing"
	"time"

	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库连接
func setupTestDB(t *testing.T) *gorm.DB {
	// 切换到项目根目录
	err := os.Chdir("../..")
	require.NoError(t, err, "切换目录失败")
	
	// 加载配置
	cfg, err := config.Load()
	require.NoError(t, err, "加载配置失败")

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化数据库
	dbManager, err := database.NewDatabase(&cfg.Database, logger)
	require.NoError(t, err, "连接数据库失败")

	db := dbManager.DB

	// 自动迁移测试需要的表结构
	err = db.AutoMigrate(
		&model.DailyData{},
		&model.WeeklyData{},
		&model.MonthlyData{},
		&model.YearlyData{},
	)
	require.NoError(t, err, "数据库迁移失败")

	return db
}

// TestKLinePersistenceService_DailyData 测试日K线数据持久化
func TestKLinePersistenceService_DailyData(t *testing.T) {
	db := setupTestDB(t)
	logger := utils.NewLogger(config.LogConfig{Level: "info"})
	service := NewKLinePersistenceService(db, logger)

	// 准备测试数据
	testData := model.DailyData{
		TsCode:    "600000.SH",
		TradeDate: 20240101,
		Open:      10.0,
		High:      11.0,
		Low:       9.5,
		Close:     10.5,
		Volume:    1000000,
		Amount:    10500000,
	}

	// 测试保存数据
	err := service.SaveDailyData(testData)
	assert.NoError(t, err, "保存日K线数据应该成功")

	// 测试批量保存
	batchData := []model.DailyData{
		{TsCode: "600000.SH", TradeDate: 20240102, Close: 10.8},
		{TsCode: "600000.SH", TradeDate: 20240103, Close: 11.2},
	}
	for _, data := range batchData {
		err = service.SaveDailyData(data)
		assert.NoError(t, err, "批量保存日K线数据应该成功")
	}

	// 测试时间范围查询
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dataList, err := service.GetDailyData("600000.SH", startDate, endDate, 10)
	assert.NoError(t, err, "时间范围查询应该成功")
	assert.Greater(t, len(dataList), 0, "应该查询到数据")
}

// TestKLinePersistenceService_WeeklyData 测试周K线数据持久化
func TestKLinePersistenceService_WeeklyData(t *testing.T) {
	db := setupTestDB(t)
	logger := utils.NewLogger(config.LogConfig{Level: "info"})
	service := NewKLinePersistenceService(db, logger)

	// 准备测试数据
	testData := model.WeeklyData{
		TsCode:    "600000.SH",
		TradeDate: 20240101,
		Open:      10.0,
		High:      12.0,
		Low:       9.0,
		Close:     11.5,
		Volume:    5000000,
		Amount:    57500000,
	}

	// 测试保存数据
	err := service.SaveWeeklyData(testData)
	assert.NoError(t, err, "保存周K线数据应该成功")

	// 测试批量保存
	batchData := []model.WeeklyData{
		{TsCode: "600000.SH", TradeDate: 20240108, Close: 11.8},
		{TsCode: "600000.SH", TradeDate: 20240115, Close: 12.2},
	}
	for _, data := range batchData {
		err = service.SaveWeeklyData(data)
		assert.NoError(t, err, "批量保存周K线数据应该成功")
	}

	// 测试时间范围查询
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dataList, err := service.GetWeeklyData("600000.SH", startDate, endDate, 10)
	assert.NoError(t, err, "时间范围查询应该成功")
	assert.Greater(t, len(dataList), 0, "应该查询到数据")
}

// TestKLinePersistenceService_MonthlyData 测试月K线数据持久化
func TestKLinePersistenceService_MonthlyData(t *testing.T) {
	db := setupTestDB(t)
	logger := utils.NewLogger(config.LogConfig{Level: "info"})
	service := NewKLinePersistenceService(db, logger)

	// 准备测试数据
	testData := model.MonthlyData{
		TsCode:    "600000.SH",
		TradeDate: 20240101,
		Open:      10.0,
		High:      15.0,
		Low:       8.5,
		Close:     14.2,
		Volume:    20000000,
		Amount:    284000000,
	}

	// 测试保存数据
	err := service.SaveMonthlyData(testData)
	assert.NoError(t, err, "保存月K线数据应该成功")

	// 测试批量保存
	batchData := []model.MonthlyData{
		{TsCode: "600000.SH", TradeDate: 20240201, Close: 14.8},
		{TsCode: "600000.SH", TradeDate: 20240301, Close: 15.2},
	}
	for _, data := range batchData {
		err = service.SaveMonthlyData(data)
		assert.NoError(t, err, "批量保存月K线数据应该成功")
	}

	// 测试时间范围查询
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dataList, err := service.GetMonthlyData("600000.SH", startDate, endDate, 10)
	assert.NoError(t, err, "时间范围查询应该成功")
	assert.Greater(t, len(dataList), 0, "应该查询到数据")
}

// TestKLinePersistenceService_YearlyData 测试年K线数据持久化
func TestKLinePersistenceService_YearlyData(t *testing.T) {
	db := setupTestDB(t)
	logger := utils.NewLogger(config.LogConfig{Level: "info"})
	service := NewKLinePersistenceService(db, logger)

	// 准备测试数据
	testData := model.YearlyData{
		TsCode:    "600000.SH",
		TradeDate: 20240101,
		Open:      10.0,
		High:      20.0,
		Low:       8.0,
		Close:     18.5,
		Volume:    240000000,
		Amount:    4440000000,
	}

	// 测试保存数据
	err := service.SaveYearlyData(testData)
	assert.NoError(t, err, "保存年K线数据应该成功")

	// 测试批量保存
	batchData := []model.YearlyData{
		{TsCode: "600000.SH", TradeDate: 20250101, Close: 19.8},
		{TsCode: "600000.SH", TradeDate: 20260101, Close: 21.2},
	}
	for _, data := range batchData {
		err = service.SaveYearlyData(data)
		assert.NoError(t, err, "批量保存年K线数据应该成功")
	}

	// 测试时间范围查询
	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)

	dataList, err := service.GetYearlyData("600000.SH", startDate, endDate, 10)
	assert.NoError(t, err, "时间范围查询应该成功")
	assert.Greater(t, len(dataList), 0, "应该查询到数据")
}

// TestKLinePersistenceService_Integration 集成测试
func TestKLinePersistenceService_Integration(t *testing.T) {
	db := setupTestDB(t)
	logger := utils.NewLogger(config.LogConfig{Level: "info"})
	service := NewKLinePersistenceService(db, logger)

	// 准备测试数据
	tsCode := "600000.SH"

	// 保存各类型数据
	dailyData := model.DailyData{
		TsCode: tsCode, TradeDate: 20240101, Close: 10.0,
	}
	weeklyData := model.WeeklyData{
		TsCode: tsCode, TradeDate: 20240101, Close: 10.0,
	}
	monthlyData := model.MonthlyData{
		TsCode: tsCode, TradeDate: 20240101, Close: 10.0,
	}
	yearlyData := model.YearlyData{
		TsCode: tsCode, TradeDate: 20240101, Close: 10.0,
	}

	// 测试保存所有类型的数据
	assert.NoError(t, service.SaveDailyData(dailyData))
	assert.NoError(t, service.SaveWeeklyData(weeklyData))
	assert.NoError(t, service.SaveMonthlyData(monthlyData))
	assert.NoError(t, service.SaveYearlyData(yearlyData))

	// 验证数据已保存
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dailyList, err := service.GetDailyData(tsCode, startDate, endDate, 10)
	assert.NoError(t, err)
	assert.Len(t, dailyList, 1)

	weeklyList, err := service.GetWeeklyData(tsCode, startDate, endDate, 10)
	assert.NoError(t, err)
	assert.Len(t, weeklyList, 1)

	monthlyList, err := service.GetMonthlyData(tsCode, startDate, endDate, 10)
	assert.NoError(t, err)
	assert.Len(t, monthlyList, 1)

	yearlyList, err := service.GetYearlyData(tsCode, startDate, endDate, 10)
	assert.NoError(t, err)
	assert.Len(t, yearlyList, 1)

	t.Log("✅ K线数据持久化集成测试完成")
}
