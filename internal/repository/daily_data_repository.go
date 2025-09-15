package repository

import (
	"time"

	"stock/internal/model"
	"stock/internal/utils"

	"gorm.io/gorm"
)

// DailyDataRepository 日线数据仓库
type DailyDataRepository struct {
	db     *gorm.DB
	logger *utils.Logger
}

// NewDailyDataRepository 创建日线数据仓库
func NewDailyDataRepository(db *gorm.DB, logger *utils.Logger) *DailyDataRepository {
	return &DailyDataRepository{
		db:     db,
		logger: logger,
	}
}

// CreateDailyData 创建日线数据
func (r *DailyDataRepository) CreateDailyData(data *model.DailyData) error {
	if err := r.db.Create(data).Error; err != nil {
		r.logger.Errorf("Failed to create daily data for %s: %v", data.TsCode, err)
		return err
	}
	r.logger.Debugf("Created daily data: %s %d", data.TsCode, data.TradeDate)
	return nil
}

// BatchCreateDailyData 批量创建日线数据
func (r *DailyDataRepository) BatchCreateDailyData(dataList []model.DailyData) error {
	if len(dataList) == 0 {
		return nil
	}

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	batchSize := 1000
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]
		if err := tx.CreateInBatches(batch, batchSize).Error; err != nil {
			tx.Rollback()
			r.logger.Errorf("Failed to batch create daily data: %v", err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Errorf("Failed to commit batch create daily data: %v", err)
		return err
	}

	r.logger.Infof("Successfully batch created %d daily data records", len(dataList))
	return nil
}

// UpsertDailyData 批量更新或插入日线数据
func (r *DailyDataRepository) UpsertDailyData(dataList []model.DailyData) error {
	if len(dataList) == 0 {
		return nil
	}

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, data := range dataList {
		if err := tx.Save(&data).Error; err != nil {
			tx.Rollback()
			r.logger.Errorf("Failed to upsert daily data %s %d: %v",
				data.TsCode, data.TradeDate, err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Errorf("Failed to commit upsert daily data: %v", err)
		return err
	}

	r.logger.Infof("Successfully upserted %d daily data records", len(dataList))
	return nil
}

// GetDailyDataByTsCode 获取指定股票的日线数据
func (r *DailyDataRepository) GetDailyDataByTsCode(tsCode string, startDate, endDate time.Time, limit int) ([]model.DailyData, error) {
	var dataList []model.DailyData
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
		r.logger.Errorf("Failed to get daily data for %s: %v", tsCode, err)
		return nil, err
	}

	return dataList, nil
}

// GetLatestDailyData 获取最新的日线数据
func (r *DailyDataRepository) GetLatestDailyData(tsCode string) (*model.DailyData, error) {
	var data model.DailyData
	if err := r.db.Where("ts_code = ?", tsCode).Order("trade_date DESC").First(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Errorf("Failed to get latest daily data for %s: %v", tsCode, err)
		return nil, err
	}
	return &data, nil
}

// GetDailyDataByDate 获取指定日期的所有股票数据
func (r *DailyDataRepository) GetDailyDataByDate(tradeDate time.Time) ([]model.DailyData, error) {
	var dataList []model.DailyData
	if err := r.db.Where("trade_date = ?", tradeDate).Find(&dataList).Error; err != nil {
		r.logger.Errorf("Failed to get daily data for date %s: %v", tradeDate.Format("2006-01-02"), err)
		return nil, err
	}
	return dataList, nil
}

// DeleteDailyData 删除日线数据
func (r *DailyDataRepository) DeleteDailyData(tsCode string, tradeDate time.Time) error {
	if err := r.db.Where("ts_code = ? AND trade_date = ?", tsCode, tradeDate).Delete(&model.DailyData{}).Error; err != nil {
		r.logger.Errorf("Failed to delete daily data %s %s: %v",
			tsCode, tradeDate.Format("2006-01-02"), err)
		return err
	}
	r.logger.Debugf("Deleted daily data: %s %s", tsCode, tradeDate.Format("2006-01-02"))
	return nil
}

// GetDailyDataCount 获取日线数据总数
func (r *DailyDataRepository) GetDailyDataCount(tsCode string) (int64, error) {
	var count int64
	query := r.db.Model(&model.DailyData{})

	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}

	if err := query.Count(&count).Error; err != nil {
		r.logger.Errorf("Failed to get daily data count: %v", err)
		return 0, err
	}
	return count, nil
}

// GetDateRange 获取数据的日期范围
func (r *DailyDataRepository) GetDateRange(tsCode string) (startDate, endDate time.Time, err error) {
	query := r.db.Model(&model.DailyData{})

	if tsCode != "" {
		query = query.Where("ts_code = ?", tsCode)
	}

	// 获取最早日期
	if err := query.Select("MIN(trade_date)").Scan(&startDate).Error; err != nil {
		r.logger.Errorf("Failed to get start date: %v", err)
		return time.Time{}, time.Time{}, err
	}

	// 获取最晚日期
	if err := query.Select("MAX(trade_date)").Scan(&endDate).Error; err != nil {
		r.logger.Errorf("Failed to get end date: %v", err)
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}
