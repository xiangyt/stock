package service

import (
	"fmt"
	"stock/internal/logger"
	"sync"
	"time"

	"gorm.io/gorm"
	"stock/internal/model"
	"stock/internal/repository"
)

// KLinePersistenceService K线数据持久化服务
type KLinePersistenceService struct {
	db            *gorm.DB
	logger        *logger.Logger
	dailyDataRepo *repository.DailyDataRepository
	weeklyRepo    *repository.WeeklyDataRepository
	monthlyRepo   *repository.MonthlyDataRepository
	yearlyRepo    *repository.YearlyDataRepository
}

var (
	klinePersistenceServiceInstance *KLinePersistenceService
	klinePersistenceServiceOnce     sync.Once
)

// GetKLinePersistenceService 获取K线数据持久化服务单例
func GetKLinePersistenceService(db *gorm.DB, logger *logger.Logger) *KLinePersistenceService {
	klinePersistenceServiceOnce.Do(func() {
		klinePersistenceServiceInstance = &KLinePersistenceService{
			db:            db,
			logger:        logger,
			dailyDataRepo: repository.NewDailyDataRepository(db, logger),
			weeklyRepo:    repository.NewWeeklyDataRepository(db, logger),
			monthlyRepo:   repository.NewMonthlyDataRepository(db, logger),
			yearlyRepo:    repository.NewYearlyDataRepository(db, logger),
		}
	})
	return klinePersistenceServiceInstance
}

// NewKLinePersistenceService 创建K线数据持久化服务 (保持向后兼容)
func NewKLinePersistenceService(db *gorm.DB, logger *logger.Logger) *KLinePersistenceService {
	return GetKLinePersistenceService(db, logger)
}

// SaveDailyData 保存日K线数据
func (s *KLinePersistenceService) SaveDailyData(data model.DailyData) error {
	return s.dailyDataRepo.SaveDailyData([]model.DailyData{data})
}

// SaveWeeklyData 保存周K线数据 (使用Upsert)
func (s *KLinePersistenceService) SaveWeeklyData(data model.WeeklyData) error {
	return s.weeklyRepo.Upsert(&data)
}

// SaveMonthlyData 保存月K线数据 (使用Upsert)
func (s *KLinePersistenceService) SaveMonthlyData(data model.MonthlyData) error {
	return s.monthlyRepo.Upsert(&data)
}

// SaveYearlyData 保存年K线数据 (使用Upsert)
func (s *KLinePersistenceService) SaveYearlyData(data model.YearlyData) error {
	return s.yearlyRepo.Upsert(&data)
}

// BatchSaveDailyData 批量保存日K线数据
func (s *KLinePersistenceService) BatchSaveDailyData(dataList []model.DailyData) error {
	return s.dailyDataRepo.SaveDailyData(dataList)
}

// BatchSaveWeeklyData 批量保存周K线数据 (使用BatchUpsert)
func (s *KLinePersistenceService) BatchSaveWeeklyData(dataList []model.WeeklyData) error {
	return s.weeklyRepo.BatchUpsert(dataList)
}

// BatchSaveMonthlyData 批量保存月K线数据 (使用BatchUpsert)
func (s *KLinePersistenceService) BatchSaveMonthlyData(dataList []model.MonthlyData) error {
	return s.monthlyRepo.BatchUpsert(dataList)
}

// BatchSaveYearlyData 批量保存年K线数据 (使用BatchUpsert)
func (s *KLinePersistenceService) BatchSaveYearlyData(dataList []model.YearlyData) error {
	return s.yearlyRepo.BatchUpsert(dataList)
}

// GetDailyData 获取日K线数据
func (s *KLinePersistenceService) GetDailyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.DailyData, error) {
	return s.dailyDataRepo.GetDailyData(tsCode, startDate, endDate, limit)
}

// GetWeeklyData 获取周K线数据
func (s *KLinePersistenceService) GetWeeklyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.WeeklyData, error) {
	return s.weeklyRepo.GetWeeklyDataByTsCode(tsCode, startDate, endDate, limit)
}

// GetMonthlyData 获取月K线数据
func (s *KLinePersistenceService) GetMonthlyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.MonthlyData, error) {
	return s.monthlyRepo.GetMonthlyDataByTsCode(tsCode, startDate, endDate, limit)
}

// GetYearlyData 获取年K线数据
func (s *KLinePersistenceService) GetYearlyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.YearlyData, error) {
	return s.yearlyRepo.GetYearlyDataByTsCode(tsCode, startDate, endDate, limit)
}

// GetLatestDailyData 获取最新日K线数据
func (s *KLinePersistenceService) GetLatestDailyData(tsCode string) (*model.DailyData, error) {
	return s.dailyDataRepo.GetLatestDailyData(tsCode)
}

// GetLatestWeeklyData 获取最新周K线数据
func (s *KLinePersistenceService) GetLatestWeeklyData(tsCode string) (*model.WeeklyData, error) {
	return s.weeklyRepo.GetLatestWeeklyData(tsCode)
}

// GetLatestMonthlyData 获取最新月K线数据
func (s *KLinePersistenceService) GetLatestMonthlyData(tsCode string) (*model.MonthlyData, error) {
	return s.monthlyRepo.GetLatestMonthlyData(tsCode)
}

// GetLatestYearlyData 获取最新年K线数据
func (s *KLinePersistenceService) GetLatestYearlyData(tsCode string) (*model.YearlyData, error) {
	return s.yearlyRepo.GetLatestYearlyData(tsCode)
}

// GetDataStats 获取所有K线数据统计信息
func (s *KLinePersistenceService) GetDataStats(tsCode string) (map[string]interface{}, error) {
	dailyCount, err := s.dailyDataRepo.GetDailyDataCount(tsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily count: %w", err)
	}

	weeklyCount, err := s.weeklyRepo.GetWeeklyDataCount(tsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly count: %w", err)
	}

	monthlyCount, err := s.monthlyRepo.GetMonthlyDataCount(tsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly count: %w", err)
	}

	yearlyCount, err := s.yearlyRepo.GetYearlyDataCount(tsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get yearly count: %w", err)
	}

	stats := map[string]interface{}{
		"ts_code":       tsCode,
		"daily_count":   dailyCount,
		"weekly_count":  weeklyCount,
		"monthly_count": monthlyCount,
		"yearly_count":  yearlyCount,
		"total_count":   dailyCount + weeklyCount + monthlyCount + yearlyCount,
		"last_updated":  time.Now().Format("2006-01-02 15:04:05"),
	}

	return stats, nil
}

// DeleteData 删除指定时间的K线数据
func (s *KLinePersistenceService) DeleteData(tsCode string, tradeDate time.Time, dataType string) error {
	switch dataType {
	case "daily":
		return s.dailyDataRepo.DeleteDailyData(tsCode, tradeDate)
	case "weekly":
		return s.weeklyRepo.DeleteWeeklyData(tsCode, tradeDate)
	case "monthly":
		return s.monthlyRepo.DeleteMonthlyData(tsCode, tradeDate)
	case "yearly":
		return s.yearlyRepo.DeleteYearlyData(tsCode, tradeDate)
	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
}
