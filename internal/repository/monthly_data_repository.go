package repository

import (
	"errors"
	"fmt"
	"time"

	"stock/internal/logger"
	"stock/internal/model"

	"gorm.io/gorm"
)

// MonthlyData 月K线数据仓库
type MonthlyData struct {
	db *gorm.DB
}

// NewMonthlyData 创建月K线数据仓库
func NewMonthlyData(db *gorm.DB) *MonthlyData {
	return &MonthlyData{
		db: db,
	}
}

// getTableName 根据股票代码前三位确定表名
func (r *MonthlyData) getTableName(tsCode string) string {
	s := model.MonthlyData{TsCode: tsCode}
	return s.TableName()
}

// SaveMonthlyData 保存月K线数据到对应的分表
func (r *MonthlyData) SaveMonthlyData(data []model.MonthlyData) error {
	if len(data) == 0 {
		return nil
	}

	// 按表名分组
	tableGroups := make(map[string][]model.MonthlyData)

	for _, item := range data {
		tableName := item.TableName()
		if tableGroups[tableName] == nil {
			tableGroups[tableName] = make([]model.MonthlyData, 0)
		}
		tableGroups[tableName] = append(tableGroups[tableName], item)
	}

	// 分别保存到对应的表
	for tableName, tableData := range tableGroups {
		if len(tableData) > 0 {
			if err := r.db.Table(tableName).CreateInBatches(tableData, 1000).Error; err != nil {
				return fmt.Errorf("failed to save data to %s: %w", tableName, err)
			}
			logger.Infof("Saved %d records to %s", len(tableData), tableName)
		}
	}

	return nil
}

// UpsertMonthlyData 更新或插入月K线数据到对应的分表（支持分批处理）
func (r *MonthlyData) UpsertMonthlyData(data []model.MonthlyData) error {
	if len(data) == 0 {
		return nil
	}

	// 按表名分组
	tableGroups := make(map[string][]model.MonthlyData)

	for _, item := range data {
		tableName := item.TableName()
		if tableGroups[tableName] == nil {
			tableGroups[tableName] = make([]model.MonthlyData, 0)
		}
		tableGroups[tableName] = append(tableGroups[tableName], item)
	}

	// 分批处理各个表的数据
	for tableName, tableData := range tableGroups {
		if len(tableData) > 0 {
			if err := r.upsertDataInBatches(tableName, tableData); err != nil {
				return fmt.Errorf("failed to upsert data to %s: %w", tableName, err)
			}
			logger.Infof("Upserted %d records to %s", len(tableData), tableName)
		}
	}

	return nil
}

// upsertDataInBatches 分批执行 upsert 操作
func (r *MonthlyData) upsertDataInBatches(tableName string, data []model.MonthlyData) error {
	const batchSize = 500 // 每批处理500条记录

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]

		// 使用 ON DUPLICATE KEY UPDATE 进行批量 upsert
		if err := r.db.Table(tableName).Save(&batch).Error; err != nil {
			return fmt.Errorf("failed to upsert batch %d-%d: %w", i, end-1, err)
		}

		logger.Debugf("Upserted batch %d-%d (%d records) to %s", i, end-1, len(batch), tableName)
	}

	return nil
}

// GetMonthlyData 获取指定股票的月K线数据
func (r *MonthlyData) GetMonthlyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.MonthlyData, error) {
	var dataList []model.MonthlyData

	// 根据股票代码确定表名
	tableName := r.getTableName(tsCode)

	query := r.db.Table(tableName).Where("ts_code = ?", tsCode)

	if !startDate.IsZero() {
		startDateInt := startDate.Year()*10000 + int(startDate.Month())*100 + startDate.Day()
		query = query.Where("trade_date >= ?", startDateInt)
	}

	if !endDate.IsZero() {
		endDateInt := endDate.Year()*10000 + int(endDate.Month())*100 + endDate.Day()
		query = query.Where("trade_date <= ?", endDateInt)
	}

	query = query.Order("trade_date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dataList).Error; err != nil {
		logger.Errorf("Failed to get monthly data for %s: %v", tsCode, err)
		return nil, err
	}

	return dataList, nil
}

// GetLatestMonthlyData 获取最新的月K线数据
func (r *MonthlyData) GetLatestMonthlyData(tsCode string) (*model.MonthlyData, error) {
	var data model.MonthlyData

	// 根据股票代码确定表名
	tableName := r.getTableName(tsCode)

	if err := r.db.Table(tableName).Where("ts_code = ?", tsCode).
		Order("trade_date DESC").First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("Failed to get latest monthly data for %s: %v", tsCode, err)
		return nil, err
	}

	return &data, nil
}

// DeleteMonthlyData 删除月K线数据
func (r *MonthlyData) DeleteMonthlyData(tsCode string, tradeDate time.Time) error {
	// 根据股票代码确定表名
	tableName := r.getTableName(tsCode)
	db := r.db.Table(tableName).Where("ts_code = ?", tsCode)
	if !tradeDate.IsZero() {
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()
		db = db.Where("trade_date = ?", tradeDateInt)
	}
	if err := db.Delete(&model.MonthlyData{}).Error; err != nil {
		logger.Errorf("Failed to delete monthly data %s %s: %v", tsCode, tradeDate.Format("2006-01-02"), err)
		return err
	}

	logger.Debugf("Deleted monthly data: %s %s", tsCode, tradeDate.Format("2006-01-02"))
	return nil
}

// GetMonthlyDataCount 获取月K线数据总数
func (r *MonthlyData) GetMonthlyDataCount(tsCode string) (int64, error) {
	if tsCode == "" {
		// 获取所有表的数据总数
		tableNames := []string{
			"monthly_data_000", "monthly_data_001", "monthly_data_002",
			"monthly_data_300", "monthly_data_301", "monthly_data_600",
			"monthly_data_601", "monthly_data_603", "monthly_data_605",
			"monthly_data_688", "monthly_data_other",
		}

		var totalCount int64
		for _, tableName := range tableNames {
			var count int64
			if err := r.db.Table(tableName).Count(&count).Error; err != nil {
				// 如果表不存在，跳过
				continue
			}
			totalCount += count
		}

		return totalCount, nil
	}

	// 根据股票代码确定表名
	tableName := r.getTableName(tsCode)
	var count int64

	if err := r.db.Table(tableName).Where("ts_code = ?", tsCode).Count(&count).Error; err != nil {
		logger.Errorf("Failed to get monthly data count: %v", err)
		return 0, err
	}

	return count, nil
}

// GetDateRange 获取数据的日期范围
func (r *MonthlyData) GetDateRange(tsCode string) (startDate, endDate time.Time, err error) {
	if tsCode != "" {
		// 根据股票代码确定表名
		tableName := r.getTableName(tsCode)

		var startDateInt, endDateInt int
		query := r.db.Table(tableName).Where("ts_code = ?", tsCode)

		// 获取最早日期
		if err := query.Select("MIN(trade_date)").Scan(&startDateInt).Error; err != nil {
			logger.Errorf("Failed to get start date: %v", err)
			return time.Time{}, time.Time{}, err
		}

		// 获取最晚日期
		if err := query.Select("MAX(trade_date)").Scan(&endDateInt).Error; err != nil {
			logger.Errorf("Failed to get end date: %v", err)
			return time.Time{}, time.Time{}, err
		}

		// 转换为time.Time
		if startDateInt > 0 {
			year := startDateInt / 10000
			month := (startDateInt % 10000) / 100
			day := startDateInt % 100
			startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}

		if endDateInt > 0 {
			year := endDateInt / 10000
			month := (endDateInt % 10000) / 100
			day := endDateInt % 100
			endDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	} else {
		// 获取所有表的日期范围
		tableNames := []string{
			"monthly_data_000", "monthly_data_001", "monthly_data_002",
			"monthly_data_300", "monthly_data_301", "monthly_data_600",
			"monthly_data_601", "monthly_data_603", "monthly_data_605",
			"monthly_data_688", "monthly_data_other",
		}

		var globalStartDate, globalEndDate int

		for _, tableName := range tableNames {
			var tableStartDate, tableEndDate int

			// 获取表的最早日期
			if err := r.db.Table(tableName).Select("MIN(trade_date)").Scan(&tableStartDate).Error; err != nil {
				continue // 表不存在或无数据，跳过
			}

			// 获取表的最晚日期
			if err := r.db.Table(tableName).Select("MAX(trade_date)").Scan(&tableEndDate).Error; err != nil {
				continue
			}

			// 更新全局最早和最晚日期
			if globalStartDate == 0 || (tableStartDate > 0 && tableStartDate < globalStartDate) {
				globalStartDate = tableStartDate
			}
			if globalEndDate == 0 || (tableEndDate > 0 && tableEndDate > globalEndDate) {
				globalEndDate = tableEndDate
			}
		}

		// 转换为time.Time
		if globalStartDate > 0 {
			year := globalStartDate / 10000
			month := (globalStartDate % 10000) / 100
			day := globalStartDate % 100
			startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}

		if globalEndDate > 0 {
			year := globalEndDate / 10000
			month := (globalEndDate % 10000) / 100
			day := globalEndDate % 100
			endDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	}

	return startDate, endDate, nil
}

// GetAllTableStats 获取所有分表的统计信息
func (r *MonthlyData) GetAllTableStats() (map[string]interface{}, error) {
	tableNames := []string{
		"monthly_data_000", "monthly_data_001", "monthly_data_002",
		"monthly_data_300", "monthly_data_301", "monthly_data_600",
		"monthly_data_601", "monthly_data_603", "monthly_data_605",
		"monthly_data_688", "monthly_data_other",
	}

	tableStats := make(map[string]int64)
	var totalCount int64

	for _, tableName := range tableNames {
		var count int64
		if err := r.db.Table(tableName).Count(&count).Error; err != nil {
			// 如果表不存在，设置为0
			tableStats[tableName] = 0
			continue
		}
		tableStats[tableName] = count
		totalCount += count
	}

	return map[string]interface{}{
		"table_stats": tableStats,
		"total_count": totalCount,
		"table_names": tableNames,
	}, nil
}
