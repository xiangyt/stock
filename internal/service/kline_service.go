package service

import (
	"fmt"
	"time"

	"stock/internal/collector"
	"stock/internal/model"
	"stock/internal/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// KLineService K线数据服务
type KLineService struct {
	db               *gorm.DB
	logger           *logrus.Logger
	collectorManager *collector.CollectorManager
	dailyManager     *DailyKLineManager
}

// NewKLineService 创建K线数据服务
func NewKLineService(db *gorm.DB, logger *logrus.Logger, collectorManager *collector.CollectorManager) *KLineService {
	utilsLogger := &utils.Logger{Logger: logger}
	return &KLineService{
		db:               db,
		logger:           logger,
		collectorManager: collectorManager,
		dailyManager:     NewDailyKLineManager(db, utilsLogger),
	}
}

// dateToInt 将time.Time转换为YYYYMMDD格式的int
func dateToInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// intToDate 将YYYYMMDD格式的int转换为time.Time
func intToDate(dateInt int) time.Time {
	year := dateInt / 10000
	month := (dateInt % 10000) / 100
	day := dateInt % 100
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// formatDateInt 将YYYYMMDD格式的int转换为字符串
func formatDateInt(dateInt int) string {
	return intToDate(dateInt).Format("2006-01-02")
}

// GetKLineData 从数据库获取K线数据（只查询，不刷新）
func (s *KLineService) GetKLineData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	dailyData, err := s.dailyManager.GetDailyData(tsCode, startDate, endDate, 0)
	if err != nil {
		s.logger.Errorf("Failed to get daily data from database: %v", err)
		return nil, err
	}

	s.logger.Infof("Retrieved %d daily data records from database for %s", len(dailyData), tsCode)
	return dailyData, nil
}

// RefreshKLineData 从API刷新K线数据并保存到数据库
func (s *KLineService) RefreshKLineData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	s.logger.Infof("Refreshing daily data from API for %s", tsCode)

	collector, err := s.collectorManager.GetCollector("eastmoney")
	if err != nil {
		return nil, fmt.Errorf("eastmoney collector not found: %w", err)
	}

	apiData, err := collector.GetDailyKLine(tsCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily data from API: %w", err)
	}

	// 保存到数据库
	err = s.saveDailyDataToDB(tsCode, apiData, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to save data to database: %w", err)
	}

	s.logger.Infof("Successfully refreshed and saved %d daily data records for %s", len(apiData), tsCode)
	return apiData, nil
}

// GetDataRange 获取数据库中K线数据的时间范围和数量
func (s *KLineService) GetDataRange(tsCode string) (startDate, endDate time.Time, count int64, err error) {
	// 获取数据数量
	count, err = s.dailyManager.GetDailyDataCount(tsCode)
	if err != nil {
		return
	}

	if count == 0 {
		return
	}

	// 获取时间范围
	startDate, endDate, err = s.dailyManager.GetDateRange(tsCode)
	return
}

// saveDailyDataToDB 保存日线数据到数据库
func (s *KLineService) saveDailyDataToDB(tsCode string, data []model.DailyData, startDate, endDate time.Time) error {
	if len(data) == 0 {
		return nil
	}

	// 设置创建时间
	for i := range data {
		data[i].TsCode = tsCode
		data[i].CreatedAt = time.Now().Unix()
	}

	// 使用DailyKLineManager保存数据到对应的交易所表
	err := s.dailyManager.UpsertDailyData(data)
	if err != nil {
		s.logger.Errorf("Failed to save daily data for %s: %v", tsCode, err)
		return err
	}

	s.logger.Infof("Saved %d daily data records for %s", len(data), tsCode)
	return nil
}

// GetDataStats 获取数据统计信息
func (s *KLineService) GetDataStats(tsCode string) (map[string]interface{}, error) {
	count, err := s.dailyManager.GetDailyDataCount(tsCode)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return map[string]interface{}{
			"ts_code":          tsCode,
			"daily_data_count": 0,
			"message":          "No data found",
		}, nil
	}

	startDate, endDate, err := s.dailyManager.GetDateRange(tsCode)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"ts_code":          tsCode,
		"daily_data_count": count,
		"data_start_date":  startDate.Format("2006-01-02"),
		"data_end_date":    endDate.Format("2006-01-02"),
		"last_updated":     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// CheckDataFreshness 检查数据新鲜度
func (s *KLineService) CheckDataFreshness(tsCode string) (map[string]interface{}, error) {
	startDate, endDate, count, err := s.GetDataRange(tsCode)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return map[string]interface{}{
			"ts_code":     tsCode,
			"has_data":    false,
			"need_update": true,
			"message":     "No data found, need initial fetch",
		}, nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 检查是否需要更新
	needUpdate := false
	reason := ""

	// 如果最新数据不是今天的，且当前时间已过交易时间，需要更新
	if endDate.Before(today) && now.Hour() >= 15 && now.Minute() >= 30 {
		// 检查今天是否是工作日
		weekday := now.Weekday()
		if weekday >= time.Monday && weekday <= time.Friday {
			needUpdate = true
			reason = "Data is outdated, need to fetch latest data"
		}
	}

	return map[string]interface{}{
		"ts_code":      tsCode,
		"has_data":     true,
		"count":        count,
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Format("2006-01-02"),
		"need_update":  needUpdate,
		"reason":       reason,
		"last_checked": now.Format("2006-01-02 15:04:05"),
	}, nil
}

// SyncDailyDataForStock 为指定股票同步日K数据（通用方法）
func (s *KLineService) SyncDailyDataForStock(stockCode string, months int) error {
	s.logger.Infof("Starting to sync daily K-line data for %s to database...", stockCode)

	// 获取采集器
	c, err := s.collectorManager.GetCollector("eastmoney")
	if err != nil {
		return fmt.Errorf("eastmoney collector not found: %w", err)
	}

	// 设置时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0) // 指定月数前

	s.logger.Infof("Fetching daily K-line data for %s from %s to %s",
		stockCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 从API获取日K数据
	dailyData, err := c.GetDailyKLine(stockCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get daily K-line data from API: %w", err)
	}

	if len(dailyData) == 0 {
		s.logger.Warnf("No daily data found for %s", stockCode)
		return nil
	}

	s.logger.Infof("Fetched %d daily K-line records from API for %s", len(dailyData), stockCode)

	// 保存到数据库
	err = s.saveDailyDataToDB(stockCode, dailyData, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to save daily data to database: %w", err)
	}

	// 获取保存后的统计信息
	stats, err := s.GetDataStats(stockCode)
	if err != nil {
		s.logger.Errorf("Failed to get data stats for %s: %v", stockCode, err)
	} else {
		s.logger.Infof("Daily K-line data sync completed for %s: %+v", stockCode, stats)
	}

	return nil
}
