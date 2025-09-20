package repository

import (
	"errors"
	"time"

	"stock/internal/model"
	"stock/internal/utils"

	"gorm.io/gorm"
)

// YearlyDataRepository 年K线数据仓库
type YearlyDataRepository struct {
	db     *gorm.DB
	logger *utils.Logger
}

// NewYearlyDataRepository 创建年K线数据仓库
func NewYearlyDataRepository(db *gorm.DB, logger *utils.Logger) *YearlyDataRepository {
	return &YearlyDataRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建年K线数据
func (r *YearlyDataRepository) Create(data *model.YearlyData) error {
	data.CreatedAt = time.Now().Unix()
	if err := r.db.Create(data).Error; err != nil {
		r.logger.Errorf("Failed to create yearly data: %v", err)
		return err
	}
	r.logger.Debugf("Created yearly data: %s %d", data.TsCode, data.TradeDate)
	return nil
}

// BatchCreate 批量创建年K线数据
func (r *YearlyDataRepository) BatchCreate(dataList []model.YearlyData) error {
	if len(dataList) == 0 {
		return nil
	}

	now := time.Now().Unix()
	for i := range dataList {
		dataList[i].CreatedAt = now
	}

	if err := r.db.CreateInBatches(dataList, 100).Error; err != nil {
		r.logger.Errorf("Failed to batch create yearly data: %v", err)
		return err
	}

	r.logger.Debugf("Batch created %d yearly data records", len(dataList))
	return nil
}

// Upsert 更新或插入年K线数据
func (r *YearlyDataRepository) Upsert(data *model.YearlyData) error {
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
		r.logger.Errorf("Failed to upsert yearly data: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		r.logger.Debugf("Upserted yearly data: %s %d", data.TsCode, data.TradeDate)
	}
	return nil
}

// BatchUpsert 批量更新或插入年K线数据
func (r *YearlyDataRepository) BatchUpsert(dataList []model.YearlyData) error {
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
				r.logger.Errorf("Failed to upsert yearly data in batch: %v", result.Error)
				return result.Error
			}
		}

		r.logger.Debugf("Batch upserted %d yearly data records", len(dataList))
		return nil
	})
}

// Update 更新年K线数据
func (r *YearlyDataRepository) Update(data *model.YearlyData) error {
	if err := r.db.Save(data).Error; err != nil {
		r.logger.Errorf("Failed to update yearly data: %v", err)
		return err
	}
	r.logger.Debugf("Updated yearly data: %s %d", data.TsCode, data.TradeDate)
	return nil
}

// GetYearlyDataByTsCode 根据股票代码获取年K线数据
func (r *YearlyDataRepository) GetYearlyDataByTsCode(tsCode string, startDate, endDate time.Time, limit int) ([]model.YearlyData, error) {
	var dataList []model.YearlyData
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
		r.logger.Errorf("Failed to get yearly data: %v", err)
		return nil, err
	}

	return dataList, nil
}

// GetLatestYearlyData 获取最新的年K线数据
func (r *YearlyDataRepository) GetLatestYearlyData(tsCode string) (*model.YearlyData, error) {
	var data model.YearlyData
	if err := r.db.Where("ts_code = ?", tsCode).Order("trade_date DESC").First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Errorf("Failed to get latest yearly data: %v", err)
		return nil, err
	}
	return &data, nil
}

// DeleteYearlyData 删除年K线数据
func (r *YearlyDataRepository) DeleteYearlyData(tsCode string, tradeDate time.Time) error {
	db := r.db.Where("ts_code = ?", tsCode)
	if !tradeDate.IsZero() {
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()
		db = db.Where("trade_date = ?", tradeDateInt)
	}
	if err := db.Delete(&model.YearlyData{}).Error; err != nil {
		r.logger.Errorf("Failed to delete yearly data: %v", err)
		return err
	}
	r.logger.Debugf("Deleted yearly data: %s %s", tsCode, tradeDate.Format("2006-01-02"))
	return nil
}

// GetYearlyDataCount 获取年K线数据总数
func (r *YearlyDataRepository) GetYearlyDataCount(tsCode string) (int64, error) {
	var count int64
	query := r.db.Model(&model.YearlyData{})
	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}
	if err := query.Count(&count).Error; err != nil {
		r.logger.Errorf("Failed to get yearly data count: %v", err)
		return 0, err
	}
	return count, nil
}

// GetDateRange 获取数据的日期范围
func (r *YearlyDataRepository) GetDateRange(tsCode string) (startDate, endDate time.Time, err error) {
	var startDateInt, endDateInt int
	query := r.db.Model(&model.YearlyData{})
	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}

	if err = query.Select("MIN(trade_date)").Scan(&startDateInt).Error; err != nil {
		r.logger.Errorf("Failed to get min trade date: %v", err)
		return
	}

	if err = query.Select("MAX(trade_date)").Scan(&endDateInt).Error; err != nil {
		r.logger.Errorf("Failed to get max trade date: %v", err)
		return
	}

	startDate = r.intToDate(startDateInt)
	endDate = r.intToDate(endDateInt)
	return
}

// intToDate 将YYYYMMDD格式的int转换为time.Time
func (r *YearlyDataRepository) intToDate(dateInt int) time.Time {
	year := dateInt / 10000
	month := (dateInt % 10000) / 100
	day := dateInt % 100
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
