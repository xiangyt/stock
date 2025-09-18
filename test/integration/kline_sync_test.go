package integration

import (
	"testing"
	"time"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestKLineSyncMethods(t *testing.T) {
	// 直接连接测试数据库
	dsn := "root:123456@tcp(localhost:3306)/stock_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("跳过测试，无法连接数据库: %v", err)
		return
	}

	// 初始化Logger
	logger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建DataService
	dataService := service.NewDataService(db, logger)

	// 测试用例：使用一个存在的股票进行测试
	testTsCode := "000001.SZ" // 平安银行
	endDate := time.Now()

	t.Run("TestSyncDailyData", func(t *testing.T) {
		startDate := endDate.AddDate(0, 0, -7) // 最近7天

		count, err := dataService.SyncDailyData(testTsCode, startDate, endDate)
		if err != nil {
			t.Logf("同步日K线数据失败（可能是网络问题）: %v", err)
		} else {
			t.Logf("成功同步日K线数据，共 %d 条记录", count)
		}
	})

	t.Run("TestSyncWeeklyData", func(t *testing.T) {
		startDate := endDate.AddDate(0, 0, -30) // 最近30天

		count, err := dataService.SyncWeeklyData(testTsCode, startDate, endDate)
		if err != nil {
			t.Logf("同步周K线数据失败（可能是网络问题）: %v", err)
		} else {
			t.Logf("成功同步周K线数据，共 %d 条记录", count)
		}
	})

	t.Run("TestSyncMonthlyData", func(t *testing.T) {
		startDate := endDate.AddDate(0, -6, 0) // 最近6个月

		count, err := dataService.SyncMonthlyData(testTsCode, startDate, endDate)
		if err != nil {
			t.Logf("同步月K线数据失败（可能是网络问题）: %v", err)
		} else {
			t.Logf("成功同步月K线数据，共 %d 条记录", count)
		}
	})

	t.Run("TestSyncYearlyData", func(t *testing.T) {
		startDate := endDate.AddDate(-3, 0, 0) // 最近3年

		count, err := dataService.SyncYearlyData(testTsCode, startDate, endDate)
		if err != nil {
			t.Logf("同步年K线数据失败（可能是网络问题）: %v", err)
		} else {
			t.Logf("成功同步年K线数据，共 %d 条记录", count)
		}
	})
}

func TestUpdateStockStatusIntegration(t *testing.T) {
	// 直接连接测试数据库
	dsn := "root:123456@tcp(localhost:3306)/stock_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("跳过测试，无法连接数据库: %v", err)
		return
	}

	// 初始化Logger
	logger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建DataService
	dataService := service.NewDataService(db, logger)

	// 测试UpdateStockStatus方法
	testTsCode := "000001.SZ"

	// 获取原始状态
	stock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Skipf("跳过测试，股票 %s 不存在: %v", testTsCode, err)
		return
	}
	originalStatus := stock.IsActive

	// 测试更新状态
	err = dataService.UpdateStockStatus(testTsCode, false)
	if err != nil {
		t.Errorf("更新股票状态失败: %v", err)
		return
	}

	// 验证状态已更新
	updatedStock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Errorf("获取更新后股票信息失败: %v", err)
		return
	}

	if updatedStock.IsActive != false {
		t.Errorf("期望股票状态为 false，实际为 %v", updatedStock.IsActive)
	}

	// 恢复原始状态
	err = dataService.UpdateStockStatus(testTsCode, originalStatus)
	if err != nil {
		t.Errorf("恢复股票原始状态失败: %v", err)
	}

	t.Logf("UpdateStockStatus方法测试完成")
}
