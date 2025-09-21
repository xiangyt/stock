package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"stock/internal/collector"
	"stock/internal/logger"
	"stock/internal/model"
	"stock/internal/repository"
)

// DataService 数据服务
type DataService struct {
	db               *gorm.DB
	logger           *logger.Logger
	stockRepo        *repository.StockRepository
	dailyDataRepo    *repository.DailyDataRepository
	weeklyDataRepo   *repository.WeeklyDataRepository
	monthlyDataRepo  *repository.MonthlyDataRepository
	yearlyDataRepo   *repository.YearlyDataRepository
	collectorFactory *collector.CollectorFactory
}

var (
	dataServiceInstance *DataService
	dataServiceOnce     sync.Once
)

// GetDataService 获取数据服务单例
func GetDataService(db *gorm.DB, logger *logger.Logger) *DataService {
	dataServiceOnce.Do(func() {
		dataServiceInstance = &DataService{
			db:               db,
			logger:           logger,
			stockRepo:        repository.NewStockRepository(db, logger),
			dailyDataRepo:    repository.NewDailyDataRepository(db, logger),
			weeklyDataRepo:   repository.NewWeeklyDataRepository(db, logger),
			monthlyDataRepo:  repository.NewMonthlyDataRepository(db, logger),
			yearlyDataRepo:   repository.NewYearlyDataRepository(db, logger),
			collectorFactory: collector.NewCollectorFactory(logger),
		}
	})
	return dataServiceInstance
}

// NewDataService 创建数据服务 (保持向后兼容)
func NewDataService(db *gorm.DB, logger *logger.Logger) *DataService {
	return GetDataService(db, logger)
}

// GetDB 获取数据库连接
func (s *DataService) GetDB() *gorm.DB {
	return s.db
}

// SyncStockList 同步股票列表
func (s *DataService) SyncStockList() error {
	s.logger.Info("Starting stock list synchronization...")

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return fmt.Errorf("failed to create EastMoney collector: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return fmt.Errorf("failed to connect to EastMoney: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取股票列表
	stocks, err := eastMoney.GetStockList()
	if err != nil {
		return fmt.Errorf("failed to get stock list: %v", err)
	}

	s.logger.Infof("Fetched %d stocks from EastMoney", len(stocks))

	for _, stock := range stocks {
		if strings.HasPrefix(stock.Name, "XD") { // 除权日清理所有k线数据
			_ = s.dailyDataRepo.DeleteDailyData(stock.TsCode, time.Time{})
			_ = s.weeklyDataRepo.DeleteWeeklyData(stock.TsCode, time.Time{})
			_ = s.monthlyDataRepo.DeleteMonthlyData(stock.TsCode, time.Time{})
			_ = s.yearlyDataRepo.DeleteYearlyData(stock.TsCode, time.Time{})
		}
	}
	// 批量更新或插入股票数据
	if err := s.stockRepo.UpsertStocks(stocks); err != nil {
		return fmt.Errorf("failed to upsert stocks: %v", err)
	}

	s.logger.Infof("Successfully synchronized %d stocks", len(stocks))
	return nil
}

// GetAllStocks 获取所有股票列表
func (s *DataService) GetAllStocks() ([]*model.Stock, error) {
	stocks, err := s.stockRepo.GetAllStocks()
	if err != nil {
		return nil, err
	}

	// 转换为指针切片
	result := make([]*model.Stock, len(stocks))
	for i := range stocks {
		result[i] = &stocks[i]
	}
	return result, nil
}

// UpdateStockStatus 更新股票状态
func (s *DataService) UpdateStockStatus(tsCode string, isActive bool) error {
	s.logger.Infof("更新股票 %s 状态为: %v", tsCode, isActive)

	// 使用数据库直接更新股票状态
	result := s.db.Model(&model.Stock{}).Where("ts_code = ?", tsCode).Update("is_active", isActive)
	if result.Error != nil {
		return fmt.Errorf("更新股票状态失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到股票 %s", tsCode)
	}

	s.logger.Infof("成功更新股票 %s 状态为: %v", tsCode, isActive)
	return nil
}

// SyncDailyData 同步日K线数据
func (s *DataService) SyncDailyData(tsCode string, startDate, endDate time.Time) (int, error) {
	s.logger.Infof("开始同步股票 %s 的日K线数据，时间范围: %s 到 %s",
		tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return 0, fmt.Errorf("创建采集器失败: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return 0, fmt.Errorf("连接数据源失败: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取日K线数据
	klineData, err := eastMoney.GetDailyKLine(tsCode, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("获取日K线数据失败: %v", err)
	}

	if len(klineData) == 0 {
		s.logger.Debugf("股票 %s 在指定时间范围内没有日K线数据", tsCode)
		return 0, nil
	}

	// 批量保存数据
	if err := s.dailyDataRepo.UpsertDailyData(klineData); err != nil {
		return 0, fmt.Errorf("保存日K线数据失败: %v", err)
	}

	s.logger.Infof("成功同步股票 %s 的日K线数据，共 %d 条记录", tsCode, len(klineData))
	return len(klineData), nil
}

// SyncWeeklyData 同步周K线数据
func (s *DataService) SyncWeeklyData(tsCode string, startDate, endDate time.Time) (int, error) {
	s.logger.Infof("开始同步股票 %s 的周K线数据，时间范围: %s 到 %s",
		tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return 0, fmt.Errorf("创建采集器失败: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return 0, fmt.Errorf("连接数据源失败: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取周K线数据
	klineData, err := eastMoney.GetWeeklyKLine(tsCode, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("获取周K线数据失败: %v", err)
	}

	if len(klineData) == 0 {
		s.logger.Debugf("股票 %s 在指定时间范围内没有周K线数据", tsCode)
		return 0, nil
	}

	// 使用KLinePersistenceService保存周K线数据
	klinePersistence := GetKLinePersistenceService(s.db, s.logger)
	for _, data := range klineData {
		weeklyData := model.WeeklyData{
			TsCode:    data.TsCode,
			TradeDate: data.TradeDate,
			Open:      data.Open,
			High:      data.High,
			Low:       data.Low,
			Close:     data.Close,
			Volume:    data.Volume,
			Amount:    data.Amount,
		}
		if err := klinePersistence.SaveWeeklyData(weeklyData); err != nil {
			return 0, fmt.Errorf("保存周K线数据失败: %v", err)
		}
	}

	s.logger.Infof("成功同步股票 %s 的周K线数据，共 %d 条记录", tsCode, len(klineData))
	return len(klineData), nil
}

// SyncMonthlyData 同步月K线数据
func (s *DataService) SyncMonthlyData(tsCode string, startDate, endDate time.Time) (int, error) {
	s.logger.Infof("开始同步股票 %s 的月K线数据，时间范围: %s 到 %s",
		tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return 0, fmt.Errorf("创建采集器失败: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return 0, fmt.Errorf("连接数据源失败: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取月K线数据
	klineData, err := eastMoney.GetMonthlyKLine(tsCode, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("获取月K线数据失败: %v", err)
	}

	if len(klineData) == 0 {
		s.logger.Debugf("股票 %s 在指定时间范围内没有月K线数据", tsCode)
		return 0, nil
	}

	// 使用KLinePersistenceService保存月K线数据
	klinePersistence := GetKLinePersistenceService(s.db, s.logger)
	for _, data := range klineData {
		monthlyData := model.MonthlyData{
			TsCode:    data.TsCode,
			TradeDate: data.TradeDate,
			Open:      data.Open,
			High:      data.High,
			Low:       data.Low,
			Close:     data.Close,
			Volume:    data.Volume,
			Amount:    data.Amount,
		}
		if err := klinePersistence.SaveMonthlyData(monthlyData); err != nil {
			return 0, fmt.Errorf("保存月K线数据失败: %v", err)
		}
	}

	s.logger.Infof("成功同步股票 %s 的月K线数据，共 %d 条记录", tsCode, len(klineData))
	return len(klineData), nil
}

// SyncYearlyData 同步年K线数据
func (s *DataService) SyncYearlyData(tsCode string, startDate, endDate time.Time) (int, error) {
	s.logger.Infof("开始同步股票 %s 的年K线数据，时间范围: %s 到 %s",
		tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return 0, fmt.Errorf("创建采集器失败: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return 0, fmt.Errorf("连接数据源失败: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取年K线数据
	klineData, err := eastMoney.GetYearlyKLine(tsCode, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("获取年K线数据失败: %v", err)
	}

	if len(klineData) == 0 {
		s.logger.Debugf("股票 %s 在指定时间范围内没有年K线数据", tsCode)
		return 0, nil
	}

	// 使用KLinePersistenceService保存年K线数据
	klinePersistence := GetKLinePersistenceService(s.db, s.logger)
	for _, data := range klineData {
		yearlyData := model.YearlyData{
			TsCode:    data.TsCode,
			TradeDate: data.TradeDate,
			Open:      data.Open,
			High:      data.High,
			Low:       data.Low,
			Close:     data.Close,
			Volume:    data.Volume,
			Amount:    data.Amount,
		}
		if err := klinePersistence.SaveYearlyData(yearlyData); err != nil {
			return 0, fmt.Errorf("保存年K线数据失败: %v", err)
		}
	}

	s.logger.Infof("成功同步股票 %s 的年K线数据，共 %d 条记录", tsCode, len(klineData))
	return len(klineData), nil
}

// SyncRealtimeData 同步实时数据
func (s *DataService) SyncRealtimeData(tsCodes []string) error {
	s.logger.Infof("Starting realtime data synchronization for %d stocks", len(tsCodes))

	// 创建东方财富采集器
	eastMoney, err := s.collectorFactory.CreateCollector(collector.CollectorTypeEastMoney)
	if err != nil {
		return fmt.Errorf("failed to create EastMoney collector: %v", err)
	}

	// 连接数据源
	if err := eastMoney.Connect(); err != nil {
		return fmt.Errorf("failed to connect to EastMoney: %v", err)
	}
	defer eastMoney.Disconnect()

	// 获取实时数据
	realtimeData, err := eastMoney.GetRealtimeData(tsCodes)
	if err != nil {
		return fmt.Errorf("failed to get realtime data: %v", err)
	}

	s.logger.Infof("Fetched realtime data for %d stocks", len(realtimeData))

	// 批量更新或插入日线数据
	if err := s.dailyDataRepo.UpsertDailyData(realtimeData); err != nil {
		return fmt.Errorf("failed to upsert daily data: %v", err)
	}

	s.logger.Infof("Successfully synchronized realtime data for %d stocks", len(realtimeData))
	return nil
}

// SyncAllRealtimeData 同步所有活跃股票的实时数据
func (s *DataService) SyncAllRealtimeData() error {
	s.logger.Info("Starting full realtime data synchronization...")

	// 获取所有活跃股票
	stocks, err := s.stockRepo.GetAllStocks()
	if err != nil {
		return fmt.Errorf("failed to get all stocks: %v", err)
	}

	// 提取股票代码
	tsCodes := make([]string, len(stocks))
	for i, stock := range stocks {
		tsCodes[i] = stock.TsCode
	}

	// 分批同步实时数据（每批100只股票）
	batchSize := 100
	for i := 0; i < len(tsCodes); i += batchSize {
		end := i + batchSize
		if end > len(tsCodes) {
			end = len(tsCodes)
		}

		batch := tsCodes[i:end]
		if err := s.SyncRealtimeData(batch); err != nil {
			s.logger.Errorf("Failed to sync batch %d-%d: %v", i, end, err)
			continue // 继续处理下一批
		}

		// 避免请求过于频繁
		time.Sleep(1 * time.Second)
	}

	s.logger.Info("Completed full realtime data synchronization")
	return nil
}

// GetStockList 获取股票列表
func (s *DataService) GetStockList(market string, industry string, limit int) ([]model.Stock, error) {
	if market != "" {
		return s.stockRepo.GetStocksByMarket(market)
	}

	if industry != "" {
		return s.stockRepo.GetStocksByIndustry(industry)
	}

	return s.stockRepo.GetAllStocks()
}

// GetStockInfo 获取股票信息
func (s *DataService) GetStockInfo(tsCode string) (*model.Stock, error) {
	return s.stockRepo.GetStockByTsCode(tsCode)
}

// GetDailyData 获取日线数据
func (s *DataService) GetDailyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.DailyData, error) {
	return s.dailyDataRepo.GetDailyData(tsCode, startDate, endDate, limit)
}

// GetLatestPrice 获取最新价格
func (s *DataService) GetLatestPrice(tsCode string) (*model.DailyData, error) {
	return s.dailyDataRepo.GetLatestDailyData(tsCode)
}

// SearchStocks 搜索股票
func (s *DataService) SearchStocks(keyword string, limit int) ([]model.Stock, error) {
	return s.stockRepo.SearchStocks(keyword, limit)
}

// GetDataStats 获取数据统计信息
func (s *DataService) GetDataStats() (map[string]interface{}, error) {
	stockCount, err := s.stockRepo.GetStockCount()
	if err != nil {
		return nil, err
	}

	dailyDataCount, err := s.dailyDataRepo.GetDailyDataCount("")
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := s.dailyDataRepo.GetDateRange("")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"stock_count":      stockCount,
		"daily_data_count": dailyDataCount,
		"data_start_date":  startDate.Format("2006-01-02"),
		"data_end_date":    endDate.Format("2006-01-02"),
		"last_updated":     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
