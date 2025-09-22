package repository

import (
	"errors"
	"time"

	logger "stock/internal/logger"
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

// Create 创建月K线数据
func (r *MonthlyData) Create(data *model.MonthlyData) error {
	data.CreatedAt = time.Now().Unix()
	if err := r.db.Create(data).Error; err != nil {
		logger.Errorf("Failed to create monthly data: %v", err)
		return err
	}
	logger.Debugf("Created monthly data: %s %d", data.TsCode, data.TradeDate)
	return nil
}

// BatchCreate 批量创建月K线数据
func (r *MonthlyData) BatchCreate(dataList []model.MonthlyData) error {
	if len(dataList) == 0 {
		return nil
	}

	now := time.Now().Unix()
	for i := range dataList {
		dataList[i].CreatedAt = now
	}

	if err := r.db.CreateInBatches(dataList, 100).Error; err != nil {
		logger.Errorf("Failed to batch create monthly data: %v", err)
		return err
	}

	logger.Debugf("Batch created %d monthly data records", len(dataList))
	return nil
}

// Upsert 更新或插入月K线数据
func (r *MonthlyData) Upsert(data *model.MonthlyData) error {
	now := time.Now().Unix()
	data.UpdatedAt = now

	// 使用联合主键进行upsert操作
	result := r.db.Where("ts_code = ? AND trade_date = ?", data.TsCode, data.TradeDate).
		Assign(map[string]interface{}{
			"open":       data.Open,
			"high":       data.High,
			"low":        data.Low,
			"close":      data.Close,
			"volume":     data.Volume,
			"amount":     data.Amount,
			"updated_at": now,
		}).
		FirstOrCreate(data)

	if result.Error != nil {
		logger.Errorf("Failed to upsert monthly data: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Debugf("Upserted monthly data: %s %d", data.TsCode, data.TradeDate)
	}
	return nil
}

// BatchUpsert 批量更新或插入月K线数据
func (r *MonthlyData) BatchUpsert(dataList []model.MonthlyData) error {
	if len(dataList) == 0 {
		return nil
	}

	now := time.Now().Unix()

	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range dataList {
			dataList[i].UpdatedAt = now

			result := tx.Where("ts_code = ? AND trade_date = ?", dataList[i].TsCode, dataList[i].TradeDate).
				Assign(map[string]interface{}{
					"open":       dataList[i].Open,
					"high":       dataList[i].High,
					"low":        dataList[i].Low,
					"close":      dataList[i].Close,
					"volume":     dataList[i].Volume,
					"amount":     dataList[i].Amount,
					"updated_at": now,
				}).
				FirstOrCreate(&dataList[i])

			if result.Error != nil {
				logger.Errorf("Failed to upsert monthly data in batch: %v", result.Error)
				return result.Error
			}
		}

		logger.Debugf("Batch upserted %d monthly data records", len(dataList))
		return nil
	})
}

// Update 更新月K线数据
func (r *MonthlyData) Update(data *model.MonthlyData) error {
	if err := r.db.Save(data).Error; err != nil {
		logger.Errorf("Failed to update monthly data: %v", err)
		return err
	}
	logger.Debugf("Updated monthly data: %s %d", data.TsCode, data.TradeDate)
	return nil
}

// GetMonthlyDataByTsCode 根据股票代码获取月K线数据
func (r *MonthlyData) GetMonthlyDataByTsCode(tsCode string, startDate, endDate time.Time, limit int) ([]model.MonthlyData, error) {
	var dataList []model.MonthlyData
	query := r.db.Where("ts_code = ?", tsCode)

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
		logger.Errorf("Failed to get monthly data: %v", err)
		return nil, err
	}

	return dataList, nil
}

// GetLatestMonthlyData 获取最新的月K线数据
func (r *MonthlyData) GetLatestMonthlyData(tsCode string) (*model.MonthlyData, error) {
	var data model.MonthlyData
	if err := r.db.Where("ts_code = ?", tsCode).Order("trade_date DESC").First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Errorf("Failed to get latest monthly data: %v", err)
		return nil, err
	}
	return &data, nil
}

// DeleteMonthlyData 删除月K线数据
func (r *MonthlyData) DeleteMonthlyData(tsCode string, tradeDate time.Time) error {
	db := r.db.Where("ts_code = ? ", tsCode)
	if !tradeDate.IsZero() {
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()
		db = db.Where("trade_date = ?", tradeDateInt)
	}
	if err := db.Delete(&model.MonthlyData{}).Error; err != nil {
		logger.Errorf("Failed to delete monthly data: %v", err)
		return err
	}
	logger.Debugf("Deleted monthly data: %s %s", tsCode, tradeDate.Format("2006-01-02"))
	return nil
}

// GetMonthlyDataCount 获取月K线数据总数
func (r *MonthlyData) GetMonthlyDataCount(tsCode string) (int64, error) {
	var count int64
	query := r.db.Model(&model.MonthlyData{})
	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}
	if err := query.Count(&count).Error; err != nil {
		logger.Errorf("Failed to get monthly data count: %v", err)
		return 0, err
	}
	return count, nil
}

// GetDateRange 获取数据的日期范围
func (r *MonthlyData) GetDateRange(tsCode string) (startDate, endDate time.Time, err error) {
	var startDateInt, endDateInt int
	query := r.db.Model(&model.MonthlyData{})
	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}

	if err = query.Select("MIN(trade_date)").Scan(&startDateInt).Error; err != nil {
		logger.Errorf("Failed to get min trade date: %v", err)
		return
	}

	if err = query.Select("MAX(trade_date)").Scan(&endDateInt).Error; err != nil {
		logger.Errorf("Failed to get max trade date: %v", err)
		return
	}

	startDate = r.intToDate(startDateInt)
	endDate = r.intToDate(endDateInt)
	return
}

// intToDate 将YYYYMMDD格式的int转换为time.Time
func (r *MonthlyData) intToDate(dateInt int) time.Time {
	year := dateInt / 10000
	month := (dateInt % 10000) / 100
	day := dateInt % 100
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
