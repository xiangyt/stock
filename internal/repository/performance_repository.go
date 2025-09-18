package repository

import (
	"fmt"
	"time"

	"stock/internal/model"

	"gorm.io/gorm"
)

// PerformanceRepository 业绩报表数据仓库
type PerformanceRepository struct {
	db *gorm.DB
}

// NewPerformanceRepository 创建业绩报表仓库实例
func NewPerformanceRepository(db *gorm.DB) *PerformanceRepository {
	return &PerformanceRepository{
		db: db,
	}
}

// Create 创建业绩报表记录
func (r *PerformanceRepository) Create(report *model.PerformanceReport) error {
	return r.db.Create(report).Error
}

// CreateBatch 批量创建业绩报表记录
func (r *PerformanceRepository) CreateBatch(reports []model.PerformanceReport) error {
	if len(reports) == 0 {
		return nil
	}
	return r.db.CreateInBatches(reports, 100).Error
}

// GetByTsCode 根据股票代码获取业绩报表
func (r *PerformanceRepository) GetByTsCode(tsCode string) ([]model.PerformanceReport, error) {
	var reports []model.PerformanceReport
	err := r.db.Where("ts_code = ?", tsCode).
		Order("report_date DESC").
		Find(&reports).Error
	return reports, err
}

// GetLatestByTsCode 获取指定股票的最新业绩报表
func (r *PerformanceRepository) GetLatestByTsCode(tsCode string) (*model.PerformanceReport, error) {
	var report model.PerformanceReport
	err := r.db.Where("ts_code = ?", tsCode).
		Order("report_date DESC").
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetByTsCodeAndDateRange 根据股票代码和日期范围获取业绩报表
func (r *PerformanceRepository) GetByTsCodeAndDateRange(tsCode string, startDate, endDate time.Time) ([]model.PerformanceReport, error) {
	var reports []model.PerformanceReport
	err := r.db.Where("ts_code = ? AND report_date BETWEEN ? AND ?", tsCode, startDate, endDate).
		Order("report_date DESC").
		Find(&reports).Error
	return reports, err
}

// GetByReportDate 根据报告日期获取所有股票的业绩报表
func (r *PerformanceRepository) GetByReportDate(reportDate time.Time) ([]model.PerformanceReport, error) {
	var reports []model.PerformanceReport
	err := r.db.Where("report_date = ?", reportDate).
		Find(&reports).Error
	return reports, err
}

// Update 更新业绩报表记录
func (r *PerformanceRepository) Update(report *model.PerformanceReport) error {
	return r.db.Save(report).Error
}

// Delete 删除业绩报表记录
func (r *PerformanceRepository) Delete(id uint) error {
	return r.db.Delete(&model.PerformanceReport{}, id).Error
}

// DeleteByTsCode 删除指定股票的所有业绩报表
func (r *PerformanceRepository) DeleteByTsCode(tsCode string) error {
	return r.db.Where("ts_code = ?", tsCode).Delete(&model.PerformanceReport{}).Error
}

// Exists 检查业绩报表是否存在
func (r *PerformanceRepository) Exists(tsCode string, reportDate int) (bool, error) {
	var count int64
	err := r.db.Model(&model.PerformanceReport{}).
		Where("ts_code = ? AND report_date = ?", tsCode, reportDate).
		Count(&count).Error
	return count > 0, err
}

// UpsertBatch 批量插入或更新业绩报表（如果存在则更新，不存在则插入）
func (r *PerformanceRepository) UpsertBatch(reports []model.PerformanceReport) error {
	if len(reports) == 0 {
		return nil
	}

	// 使用事务处理批量upsert
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, report := range reports {
			// 检查记录是否存在
			var existingReport model.PerformanceReport
			err := tx.Where("ts_code = ? AND report_date = ?", report.TsCode, report.ReportDate).
				First(&existingReport).Error

			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				if err := tx.Create(&report).Error; err != nil {
					return fmt.Errorf("failed to create performance report for %s: %w", report.TsCode, err)
				}
			} else if err != nil {
				// 查询出错
				return fmt.Errorf("failed to query existing performance report: %w", err)
			} else {
				// 记录存在，更新记录
				report.CreatedAt = existingReport.CreatedAt // 保持原有创建时间
				if err := tx.Save(&report).Error; err != nil {
					return fmt.Errorf("failed to update performance report for %s: %w", report.TsCode, err)
				}
			}
		}
		return nil
	})
}

// GetStatistics 获取业绩报表统计信息
func (r *PerformanceRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总记录数
	var totalCount int64
	if err := r.db.Model(&model.PerformanceReport{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total reports: %w", err)
	}
	stats["total_reports"] = totalCount

	// 股票数量
	var stockCount int64
	if err := r.db.Model(&model.PerformanceReport{}).
		Distinct("ts_code").
		Count(&stockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count distinct stocks: %w", err)
	}
	stats["stock_count"] = stockCount

	// 最新报告日期
	var latestDate time.Time
	if err := r.db.Model(&model.PerformanceReport{}).
		Select("MAX(report_date)").
		Scan(&latestDate).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest report date: %w", err)
	}
	stats["latest_report_date"] = latestDate

	// 最早报告日期
	var earliestDate time.Time
	if err := r.db.Model(&model.PerformanceReport{}).
		Select("MIN(report_date)").
		Scan(&earliestDate).Error; err != nil {
		return nil, fmt.Errorf("failed to get earliest report date: %w", err)
	}
	stats["earliest_report_date"] = earliestDate

	return stats, nil
}

// GetTopPerformers 获取业绩表现最好的股票
func (r *PerformanceRepository) GetTopPerformers(limit int, orderBy string) ([]model.PerformanceReport, error) {
	var reports []model.PerformanceReport

	// 验证排序字段
	validOrderFields := map[string]bool{
		"eps":            true,
		"roe":            true,
		"roa":            true,
		"gross_margin":   true,
		"dividend_yield": true,
		"revenue":        true,
		"net_profit":     true,
	}

	if !validOrderFields[orderBy] {
		orderBy = "eps" // 默认按每股收益排序
	}

	err := r.db.Where("report_date = (SELECT MAX(report_date) FROM performance_reports pr2 WHERE pr2.ts_code = performance_reports.ts_code)").
		Order(fmt.Sprintf("%s DESC", orderBy)).
		Limit(limit).
		Find(&reports).Error

	return reports, err
}
