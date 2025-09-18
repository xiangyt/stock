package integration

import (
	"testing"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestSimpleUpdateStockStatus(t *testing.T) {
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

	// 测试用例：找一个存在的股票进行测试
	testTsCode := "000001.SZ" // 平安银行

	// 1. 获取股票当前状态
	stock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Skipf("跳过测试，股票 %s 不存在: %v", testTsCode, err)
		return
	}

	originalStatus := stock.IsActive
	t.Logf("股票 %s 原始状态: %v", testTsCode, originalStatus)

	// 2. 测试更新为非活跃状态
	err = dataService.UpdateStockStatus(testTsCode, false)
	if err != nil {
		t.Fatalf("更新股票状态为非活跃失败: %v", err)
	}

	// 3. 验证状态已更新
	updatedStock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Fatalf("获取更新后股票信息失败: %v", err)
	}

	if updatedStock.IsActive != false {
		t.Errorf("期望股票状态为 false，实际为 %v", updatedStock.IsActive)
	}
	t.Logf("成功将股票 %s 状态更新为: %v", testTsCode, updatedStock.IsActive)

	// 4. 测试更新为活跃状态
	err = dataService.UpdateStockStatus(testTsCode, true)
	if err != nil {
		t.Fatalf("更新股票状态为活跃失败: %v", err)
	}

	// 5. 验证状态已恢复
	restoredStock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Fatalf("获取恢复后股票信息失败: %v", err)
	}

	if restoredStock.IsActive != true {
		t.Errorf("期望股票状态为 true，实际为 %v", restoredStock.IsActive)
	}
	t.Logf("成功将股票 %s 状态恢复为: %v", testTsCode, restoredStock.IsActive)

	// 6. 恢复原始状态
	err = dataService.UpdateStockStatus(testTsCode, originalStatus)
	if err != nil {
		t.Fatalf("恢复股票原始状态失败: %v", err)
	}
	t.Logf("已恢复股票 %s 原始状态: %v", testTsCode, originalStatus)

	// 7. 测试不存在的股票
	err = dataService.UpdateStockStatus("999999.SZ", false)
	if err == nil {
		t.Error("期望更新不存在股票时返回错误，但没有返回错误")
	} else {
		t.Logf("正确处理不存在股票的情况: %v", err)
	}
}

func TestSimpleMarkStockInactive(t *testing.T) {
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

	// 创建Services结构体（模拟worker中的services）
	services := &service.Services{
		DataService: dataService,
	}

	// 测试markStockInactive函数的逻辑
	testTsCode := "000002.SZ" // 万科A

	// 获取原始状态
	stock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Skipf("跳过测试，股票 %s 不存在: %v", testTsCode, err)
		return
	}
	originalStatus := stock.IsActive

	// 确保股票是活跃状态，以便测试标记为非活跃
	if !originalStatus {
		err = dataService.UpdateStockStatus(testTsCode, true)
		if err != nil {
			t.Fatalf("设置股票为活跃状态失败: %v", err)
		}
	}

	// 模拟markStockInactive函数的逻辑
	markStockInactive := func(services *service.Services, tsCode string, logger *utils.Logger) error {
		logger.Infof("标记股票 %s 为非活跃状态", tsCode)

		// 获取股票信息
		stock, err := services.DataService.GetStockInfo(tsCode)
		if err != nil {
			return err
		}

		// 检查股票是否已经是非活跃状态
		if !stock.IsActive {
			logger.Debugf("股票 %s 已经是非活跃状态", tsCode)
			return nil
		}

		// 更新股票状态为非活跃
		err = services.DataService.UpdateStockStatus(tsCode, false)
		if err != nil {
			return err
		}

		logger.Infof("成功标记股票 %s 为非活跃状态", tsCode)
		return nil
	}

	// 执行标记为非活跃
	err = markStockInactive(services, testTsCode, logger)
	if err != nil {
		t.Fatalf("标记股票为非活跃失败: %v", err)
	}

	// 验证状态已更新
	updatedStock, err := dataService.GetStockInfo(testTsCode)
	if err != nil {
		t.Fatalf("获取更新后股票信息失败: %v", err)
	}

	if updatedStock.IsActive {
		t.Errorf("期望股票状态为非活跃，但仍为活跃状态")
	}

	// 再次调用，应该不会重复标记
	err = markStockInactive(services, testTsCode, logger)
	if err != nil {
		t.Fatalf("重复标记股票为非活跃失败: %v", err)
	}

	// 恢复原始状态
	err = dataService.UpdateStockStatus(testTsCode, originalStatus)
	if err != nil {
		t.Fatalf("恢复股票原始状态失败: %v", err)
	}

	t.Logf("markStockInactive函数测试完成")
}
