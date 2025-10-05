package service

import (
	"context"
	"fmt"
	"stock/internal/logger"
	"sync"
	"time"

	"stock/internal/collector"
	"stock/internal/model"
	"stock/internal/repository"
)

// PerformanceService 业绩报表服务
type PerformanceService struct {
	repo      *repository.Performance
	stockRepo *repository.Stock
	collector collector.DataCollector
}

var (
	performanceServiceInstance *PerformanceService
	performanceServiceOnce     sync.Once
)

// GetPerformanceService 获取业绩报表服务单例
func GetPerformanceService(repo *repository.Performance, stockRepo *repository.Stock, collector collector.DataCollector,
) *PerformanceService {
	performanceServiceOnce.Do(func() {
		performanceServiceInstance = &PerformanceService{
			repo:      repo,
			stockRepo: stockRepo,
			collector: collector,
		}
	})
	return performanceServiceInstance
}

// NewPerformanceService 创建业绩报表服务实例 (保持向后兼容)
func NewPerformanceService(repo *repository.Performance, stockRepo *repository.Stock,
	collector collector.DataCollector) *PerformanceService {
	return GetPerformanceService(repo, stockRepo, collector)
}

// GetPerformanceReports 获取业绩报表数据
func (s *PerformanceService) GetPerformanceReports(ctx context.Context, tsCode string) ([]model.PerformanceReport, error) {
	logger.Infof("Getting performance reports for stock: %s", tsCode)

	// 首先从数据库查询
	reports, err := s.repo.GetByTsCode(tsCode)
	if err != nil {
		logger.Errorf("Failed to query performance reports from database: %v", err)
		return nil, fmt.Errorf("failed to query performance reports: %w", err)
	}

	// 如果数据库中没有数据，从采集器获取
	if len(reports) == 0 {
		logger.Infof("No performance reports found in database, fetching from collector")
		reports, err = s.collector.GetPerformanceReports(tsCode)
		if err != nil {
			logger.Errorf("Failed to get performance reports from collector: %v", err)
			return nil, fmt.Errorf("failed to fetch performance reports: %w", err)
		}

		// 保存到数据库
		if len(reports) > 0 {
			if err := s.repo.CreateBatch(reports); err != nil {
				logger.Warnf("Failed to save performance reports to database: %v", err)
			} else {
				logger.Infof("Saved %d performance reports to database", len(reports))
			}
		}
	}

	return reports, nil
}

// GetLatestPerformanceReport 获取最新业绩报表
func (s *PerformanceService) GetLatestPerformanceReport(ctx context.Context, tsCode string) (*model.PerformanceReport, error) {
	return s.repo.GetLatestByTsCode(tsCode)
}

// SyncPerformanceReports 同步业绩报表数据
func (s *PerformanceService) SyncPerformanceReports(ctx context.Context, tsCode string) error {
	logger.Infof("Syncing performance reports for stock: %s", tsCode)

	// 从采集器获取最新数据
	reports, err := s.collector.GetPerformanceReports(tsCode)
	if err != nil {
		logger.Errorf("Failed to fetch performance reports from collector: %v", err)
		return fmt.Errorf("failed to fetch performance reports: %w", err)
	}

	if len(reports) == 0 {
		logger.Infof("No performance reports available for stock: %s", tsCode)
		return nil
	}

	// 批量插入或更新数据
	if err := s.repo.UpsertBatch(reports); err != nil {
		logger.Errorf("Failed to upsert performance reports: %v", err)
		return fmt.Errorf("failed to save performance reports: %w", err)
	}

	logger.Infof("Successfully synced %d performance reports for stock: %s", len(reports), tsCode)
	return nil
}

// SyncAllStocksPerformanceReports 同步所有股票的业绩报表数据
func (s *PerformanceService) SyncAllStocksPerformanceReports(ctx context.Context) error {
	logger.Info("Starting to sync performance reports for all stocks")

	// 获取所有股票
	stocks, err := s.stockRepo.GetAllStocks()
	if err != nil {
		logger.Errorf("Failed to get all stocks: %v", err)
		return fmt.Errorf("failed to get stocks: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, stock := range stocks {
		select {
		case <-ctx.Done():
			logger.Info("Sync cancelled by context")
			return ctx.Err()
		default:
		}

		if err := s.SyncPerformanceReports(ctx, stock.TsCode); err != nil {
			logger.Errorf("Failed to sync performance reports for %s: %v", stock.TsCode, err)
			errorCount++
		} else {
			successCount++
		}

		// 添加延迟避免请求过于频繁
		time.Sleep(100 * time.Millisecond)
	}

	logger.Infof("Sync completed. Success: %d, Errors: %d", successCount, errorCount)
	return nil
}

// GetPerformanceReportsByDateRange 根据日期范围获取业绩报表
func (s *PerformanceService) GetPerformanceReportsByDateRange(ctx context.Context, tsCode string, startDate, endDate time.Time) ([]model.PerformanceReport, error) {
	logger.Infof("Getting performance reports for stock %s from %s to %s", tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	reports, err := s.repo.GetByTsCodeAndDateRange(tsCode, startDate, endDate)
	if err != nil {
		logger.Errorf("Failed to get performance reports by date range: %v", err)
		return nil, fmt.Errorf("failed to get performance reports: %w", err)
	}

	return reports, nil
}

// GetTopPerformers 获取业绩表现最好的股票
func (s *PerformanceService) GetTopPerformers(ctx context.Context, limit int, orderBy string) ([]model.PerformanceReport, error) {
	logger.Infof("Getting top %d performers ordered by %s", limit, orderBy)

	reports, err := s.repo.GetTopPerformers(limit, orderBy)
	if err != nil {
		logger.Errorf("Failed to get top performers: %v", err)
		return nil, fmt.Errorf("failed to get top performers: %w", err)
	}

	return reports, nil
}

// GetStatistics 获取业绩报表统计信息
func (s *PerformanceService) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	logger.Info("Getting performance reports statistics")

	stats, err := s.repo.GetStatistics()
	if err != nil {
		logger.Errorf("Failed to get statistics: %v", err)
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return stats, nil
}

// DeletePerformanceReports 删除指定股票的业绩报表数据
func (s *PerformanceService) DeletePerformanceReports(ctx context.Context, tsCode string) error {
	logger.Infof("Deleting performance reports for stock: %s", tsCode)

	if err := s.repo.DeleteByTsCode(tsCode); err != nil {
		logger.Errorf("Failed to delete performance reports: %v", err)
		return fmt.Errorf("failed to delete performance reports: %w", err)
	}

	logger.Infof("Successfully deleted performance reports for stock: %s", tsCode)
	return nil
}

// ValidatePerformanceReport 验证业绩报表数据
func (s *PerformanceService) ValidatePerformanceReport(report *model.PerformanceReport) error {
	if report.TsCode == "" {
		return fmt.Errorf("ts_code is required")
	}

	if report.ReportDate == 0 {
		return fmt.Errorf("report_date is required")
	}

	// 验证股票是否存在
	stock, err := s.stockRepo.GetStockByTsCode(report.TsCode)
	if err != nil {
		return fmt.Errorf("failed to check stock existence: %w", err)
	}
	if stock == nil {
		return fmt.Errorf("stock with ts_code %s does not exist", report.TsCode)
	}

	return nil
}

// CreatePerformanceReport 创建业绩报表记录
func (s *PerformanceService) CreatePerformanceReport(ctx context.Context, report *model.PerformanceReport) error {
	logger.Infof("Creating performance report for stock: %s", report.TsCode)

	// 验证数据
	if err := s.ValidatePerformanceReport(report); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 检查是否已存在
	exists, err := s.repo.Exists(report.TsCode, report.ReportDate)
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if exists {
		return fmt.Errorf("performance report already exists for %s on %d", report.TsCode, report.ReportDate)
	}

	// 创建记录
	if err := s.repo.Create(report); err != nil {
		logger.Errorf("Failed to create performance report: %v", err)
		return fmt.Errorf("failed to create performance report: %w", err)
	}

	logger.Infof("Successfully created performance report for stock: %s", report.TsCode)
	return nil
}

// UpdatePerformanceReport 更新业绩报表记录
func (s *PerformanceService) UpdatePerformanceReport(ctx context.Context, report *model.PerformanceReport) error {
	logger.Infof("Updating performance report for stock: %s", report.TsCode)

	// 验证数据
	if err := s.ValidatePerformanceReport(report); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 更新记录
	if err := s.repo.Update(report); err != nil {
		logger.Errorf("Failed to update performance report: %v", err)
		return fmt.Errorf("failed to update performance report: %w", err)
	}

	logger.Infof("Successfully updated performance report for stock: %s", report.TsCode)
	return nil
}
