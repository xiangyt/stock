package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	logger "stock/internal/logger"
	"stock/internal/model"

	"gorm.io/gorm"
)

// DailyData 日线数据仓库
type DailyData struct {
	db *gorm.DB
}

// NewDailyData 创建日线数据仓库
func NewDailyData(db *gorm.DB) *DailyData {
	return &DailyData{
		db: db,
	}
}

// getExchange 根据股票代码获取交易所类型
func (r *DailyData) getExchange(tsCode string) string {
	if strings.HasSuffix(tsCode, ".SH") {
		return "SH"
	} else if strings.HasSuffix(tsCode, ".SZ") {
		return "SZ"
	}
	// 默认根据代码前缀判断
	if strings.HasPrefix(tsCode, "6") {
		return "SH"
	} else if strings.HasPrefix(tsCode, "0") || strings.HasPrefix(tsCode, "3") {
		return "SZ"
	}
	return "SH" // 默认上海
}

// SaveDailyData 保存日K线数据到对应的交易所表
func (r *DailyData) SaveDailyData(data []model.DailyData) error {
	if len(data) == 0 {
		return nil
	}

	// 按交易所分组
	shData := make([]model.DailyData, 0)
	szData := make([]model.DailyData, 0)

	for _, item := range data {
		exchange := r.getExchange(item.TsCode)
		if exchange == "SH" {
			shData = append(shData, item)
		} else {
			szData = append(szData, item)
		}
	}

	// 分别保存到对应的表
	if len(shData) > 0 {
		if err := r.db.Table("daily_data_sh").CreateInBatches(shData, 1000).Error; err != nil {
			return fmt.Errorf("failed to save SH daily data: %w", err)
		}
		logger.Infof("Saved %d SH daily data records", len(shData))
	}

	if len(szData) > 0 {
		if err := r.db.Table("daily_data_sz").CreateInBatches(szData, 1000).Error; err != nil {
			return fmt.Errorf("failed to save SZ daily data: %w", err)
		}
		logger.Infof("Saved %d SZ daily data records", len(szData))
	}

	return nil
}

// UpsertDailyData 更新或插入日K线数据到对应的交易所表（支持分批处理）
func (r *DailyData) UpsertDailyData(data []model.DailyData) error {
	if len(data) == 0 {
		return nil
	}

	// 按交易所分组
	shData := make([]model.DailyData, 0)
	szData := make([]model.DailyData, 0)

	for _, item := range data {
		exchange := r.getExchange(item.TsCode)
		if exchange == "SH" {
			shData = append(shData, item)
		} else {
			szData = append(szData, item)
		}
	}

	// 分批处理上海交易所数据
	if len(shData) > 0 {
		if err := r.upsertDataInBatches("daily_data_sh", shData); err != nil {
			return fmt.Errorf("failed to upsert SH daily data: %w", err)
		}
		logger.Infof("Upserted %d SH daily data records", len(shData))
	}

	// 分批处理深圳交易所数据
	if len(szData) > 0 {
		if err := r.upsertDataInBatches("daily_data_sz", szData); err != nil {
			return fmt.Errorf("failed to upsert SZ daily data: %w", err)
		}
		logger.Infof("Upserted %d SZ daily data records", len(szData))
	}

	return nil
}

// upsertDataInBatches 分批执行 upsert 操作
func (r *DailyData) upsertDataInBatches(tableName string, data []model.DailyData) error {
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

// GetDailyData 获取指定股票的日K线数据
func (r *DailyData) GetDailyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.DailyData, error) {
	var dataList []model.DailyData

	// 创建一个临时的DailyData对象来确定表名
	tempData := model.DailyData{TsCode: tsCode}

	query := r.db.Table(tempData.TableName()).Where("ts_code = ?", tsCode)

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
		logger.Errorf("Failed to get daily data for %s: %v", tsCode, err)
		return nil, err
	}

	return dataList, nil
}

// GetLatestDailyData 获取最新的日K线数据
func (r *DailyData) GetLatestDailyData(tsCode string) (*model.DailyData, error) {
	var data model.DailyData

	// 创建一个临时的DailyData对象来确定表名
	tempData := model.DailyData{TsCode: tsCode}

	if err := r.db.Table(tempData.TableName()).Where("ts_code = ?", tsCode).
		Order("trade_date DESC").First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("Failed to get latest daily data for %s: %v", tsCode, err)
		return nil, err
	}

	return &data, nil
}

// DeleteDailyData 删除日K线数据
func (r *DailyData) DeleteDailyData(tsCode string, tradeDate time.Time) error {
	// 创建一个临时的DailyData对象来确定表名
	tempData := model.DailyData{TsCode: tsCode}
	db := r.db.Table(tempData.TableName()).Where("ts_code = ?", tsCode)
	if !tradeDate.IsZero() {
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()
		db = db.Where("trade_date = ?", tradeDateInt)
	}
	if err := db.Delete(&model.DailyData{}).Error; err != nil {
		logger.Errorf("Failed to delete daily data %s %s: %v", tsCode, tradeDate.Format("2006-01-02"), err)
		return err
	}

	logger.Debugf("Deleted daily data: %s %s", tsCode, tradeDate.Format("2006-01-02"))
	return nil
}

// GetDailyDataCount 获取日K线数据总数
func (r *DailyData) GetDailyDataCount(tsCode string) (int64, error) {
	if tsCode == "" {
		// 获取所有数据总数
		var shCount, szCount int64

		if err := r.db.Table("daily_data_sh").Count(&shCount).Error; err != nil {
			return 0, err
		}

		if err := r.db.Table("daily_data_sz").Count(&szCount).Error; err != nil {
			return 0, err
		}

		return shCount + szCount, nil
	}

	// 创建一个临时的DailyData对象来确定表名
	tempData := model.DailyData{TsCode: tsCode}
	var count int64

	if err := r.db.Table(tempData.TableName()).Where("ts_code = ?", tsCode).Count(&count).Error; err != nil {
		logger.Errorf("Failed to get daily data count: %v", err)
		return 0, err
	}

	return count, nil
}

// GetDateRange 获取数据的日期范围
func (r *DailyData) GetDateRange(tsCode string) (startDate, endDate time.Time, err error) {
	// 创建一个临时的DailyData对象来确定表名
	tempData := model.DailyData{TsCode: tsCode}

	var startDateInt, endDateInt int
	query := r.db.Table(tempData.TableName())

	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}

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

	return startDate, endDate, nil
}

// GetAllExchangeStats 获取所有交易所的统计信息
func (r *DailyData) GetAllExchangeStats() (map[string]interface{}, error) {
	var shCount, szCount int64

	if err := r.db.Table("daily_data_sh").Count(&shCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get SH count: %w", err)
	}

	if err := r.db.Table("daily_data_sz").Count(&szCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get SZ count: %w", err)
	}

	return map[string]interface{}{
		"sh_count":    shCount,
		"sz_count":    szCount,
		"total_count": shCount + szCount,
		"exchanges":   []string{"SH", "SZ"},
	}, nil
}
