package repository

import (
	"fmt"

	"stock/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TechnicalIndicatorRepository 技术指标数据访问层
type TechnicalIndicatorRepository struct {
	db *gorm.DB
}

// NewTechnicalIndicatorRepository 创建技术指标仓库实例
func NewTechnicalIndicatorRepository(db *gorm.DB) *TechnicalIndicatorRepository {
	return &TechnicalIndicatorRepository{db: db}
}

// Create 创建技术指标记录
func (r *TechnicalIndicatorRepository) Create(indicator *model.TechnicalIndicator) error {
	return r.db.Table(indicator.TableName()).Create(indicator).Error
}

// BatchCreate 批量创建技术指标记录
func (r *TechnicalIndicatorRepository) BatchCreate(indicators []*model.TechnicalIndicator) error {
	if len(indicators) == 0 {
		return nil
	}

	// 按周期分组
	groupedIndicators := make(map[model.TechnicalIndicatorPeriod][]*model.TechnicalIndicator)
	for _, indicator := range indicators {
		period := indicator.Period
		groupedIndicators[period] = append(groupedIndicators[period], indicator)
	}
	batchSize := 500
	// 分组批量插入
	for period, periodIndicators := range groupedIndicators {
		if len(periodIndicators) == 0 {
			continue
		}

		tableName := periodIndicators[0].TableName()

		// 分批处理
		for i := 0; i < len(periodIndicators); i += batchSize {
			end := i + batchSize
			if end > len(periodIndicators) {
				end = len(periodIndicators)
			}

			batch := periodIndicators[i:end]
			if err := r.db.Table(tableName).Create(batch).Error; err != nil {
				return fmt.Errorf("failed to batch create %s indicators: %v", period, err)
			}
		}
	}

	return nil
}

// GetBySymbolAndDate 根据股票代码和交易日期获取技术指标
func (r *TechnicalIndicatorRepository) GetBySymbolAndDate(symbol string, tradeDate int,
	period model.TechnicalIndicatorPeriod) (*model.TechnicalIndicator, error) {
	indicator := model.NewTechnicalIndicator(period)

	err := r.db.Table(indicator.TableName()).
		Where("symbol = ? AND trade_date = ?", symbol, tradeDate).
		First(indicator).Error

	if err != nil {
		return nil, err
	}

	return indicator, nil
}

// GetBySymbol 根据股票代码获取技术指标列表
func (r *TechnicalIndicatorRepository) GetBySymbol(symbol string, period model.TechnicalIndicatorPeriod) (
	[]*model.TechnicalIndicator, error) {
	var indicators []*model.TechnicalIndicator
	indicator := model.NewTechnicalIndicator(period)

	query := r.db.Table(indicator.TableName()).
		Where("symbol = ?", symbol).
		Order("trade_date")

	err := query.Find(&indicators).Error
	if err != nil {
		return nil, err
	}

	// 设置周期类型
	for _, ind := range indicators {
		ind.Period = period
	}

	return indicators, nil
}

// GetByDateRange 根据日期范围获取技术指标
func (r *TechnicalIndicatorRepository) GetByDateRange(symbol string, startDate, endDate int, period model.TechnicalIndicatorPeriod) ([]*model.TechnicalIndicator, error) {
	var indicators []*model.TechnicalIndicator
	indicator := model.NewTechnicalIndicator(period)

	err := r.db.Table(indicator.TableName()).
		Where("symbol = ? AND trade_date >= ? AND trade_date <= ?", symbol, startDate, endDate).
		Order("trade_date ASC").
		Find(&indicators).Error

	if err != nil {
		return nil, err
	}

	// 设置周期类型
	for _, ind := range indicators {
		ind.Period = period
	}

	return indicators, nil
}

// Update 更新技术指标记录
func (r *TechnicalIndicatorRepository) Update(indicator *model.TechnicalIndicator) error {
	return r.db.Table(indicator.TableName()).
		Where("symbol = ? AND trade_date = ?", indicator.Symbol, indicator.TradeDate).
		Updates(indicator).Error
}

// UpsertMacd 更新macd
func (r *TechnicalIndicatorRepository) UpsertMacd(indicators []*model.TechnicalIndicator) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range indicators {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "symbol"}, {Name: "trade_date"}}, // 冲突检测列
				DoUpdates: clause.Assignments(map[string]interface{}{ // 显式赋值
					"macd":      v.Macd,
					"macd_ema1": v.MacdEma1,
					"macd_ema2": v.MacdEma2,
					"macd_dif":  v.MacdDif,
					"macd_dea":  v.MacdDea,
				}),
			}).Create(v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpsertKdj 更新kdj
func (r *TechnicalIndicatorRepository) UpsertKdj(indicators []*model.TechnicalIndicator) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range indicators {
			if err := tx.Table(v.TableName()).Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "symbol"}, {Name: "trade_date"}}, // 冲突检测列
				DoUpdates: clause.Assignments(map[string]interface{}{ // 显式赋值
					"kdj_k": v.KdjK,
					"kdj_d": v.KdjD,
					"kdj_j": v.KdjJ,
				}),
			}).Create(v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Upsert 插入或更新技术指标记录
func (r *TechnicalIndicatorRepository) Upsert(indicator *model.TechnicalIndicator) error {
	// 尝试更新
	result := r.db.Table(indicator.TableName()).
		Where("symbol = ? AND trade_date = ?", indicator.Symbol, indicator.TradeDate).
		Updates(indicator)

	if result.Error != nil {
		return result.Error
	}

	// 如果没有更新任何记录，则插入新记录
	if result.RowsAffected == 0 {
		return r.db.Table(indicator.TableName()).Create(indicator).Error
	}

	return nil
}

// Delete 删除技术指标记录
func (r *TechnicalIndicatorRepository) Delete(symbol string, tradeDate int, period model.TechnicalIndicatorPeriod) error {
	indicator := model.NewTechnicalIndicator(period)

	return r.db.Table(indicator.TableName()).
		Where("symbol = ? AND trade_date = ?", symbol, tradeDate).
		Delete(&model.TechnicalIndicator{}).Error
}

// DeleteBySymbol 删除指定股票的所有技术指标记录
func (r *TechnicalIndicatorRepository) DeleteBySymbol(symbol string, period model.TechnicalIndicatorPeriod) error {
	indicator := model.NewTechnicalIndicator(period)

	return r.db.Table(indicator.TableName()).
		Where("symbol = ?", symbol).
		Delete(&model.TechnicalIndicator{}).Error
}

// DeleteByDate 删除指定日期以前的技术指标记录
func (r *TechnicalIndicatorRepository) DeleteByDate(symbol string, date int,
	period model.TechnicalIndicatorPeriod) error {
	indicator := model.NewTechnicalIndicator(period)
	db := r.db.Table(indicator.TableName())
	if symbol != "" {
		db = db.Where("symbol = ?", symbol)
	}
	return db.Where("trade_date < ?", date).
		Delete(&model.TechnicalIndicator{}).Error
}

// Count 统计技术指标记录数量
func (r *TechnicalIndicatorRepository) Count(symbol string, period model.TechnicalIndicatorPeriod) (int64, error) {
	var count int64
	indicator := model.NewTechnicalIndicator(period)

	query := r.db.Table(indicator.TableName())
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetLatest 获取最新的技术指标记录
func (r *TechnicalIndicatorRepository) GetLatest(symbol string, period model.TechnicalIndicatorPeriod) (*model.TechnicalIndicator, error) {
	indicator := model.NewTechnicalIndicator(period)

	err := r.db.Table(indicator.TableName()).
		Where("symbol = ?", symbol).
		Order("trade_date DESC").
		First(indicator).Error

	if err != nil {
		return nil, err
	}

	return indicator, nil
}

// GetByMA20Range 根据MA20范围获取股票
func (r *TechnicalIndicatorRepository) GetByMA20Range(minMA20, maxMA20 float64, limit int, period model.TechnicalIndicatorPeriod) ([]*model.TechnicalIndicator, error) {
	var indicators []*model.TechnicalIndicator
	indicator := model.NewTechnicalIndicator(period)

	err := r.db.Table(indicator.TableName()).
		Where("ma20 >= ? AND ma20 <= ?", minMA20, maxMA20).
		Order("ma20 DESC").
		Limit(limit).
		Find(&indicators).Error

	if err != nil {
		return nil, err
	}

	// 设置周期类型
	for _, ind := range indicators {
		ind.Period = period
	}

	return indicators, nil
}

// GetSymbolsByDate 获取指定日期有技术指标数据的股票代码列表
func (r *TechnicalIndicatorRepository) GetSymbolsByDate(tradeDate int, period model.TechnicalIndicatorPeriod) ([]string, error) {
	var symbols []string
	indicator := model.NewTechnicalIndicator(period)

	err := r.db.Table(indicator.TableName()).
		Select("DISTINCT symbol").
		Where("trade_date = ?", tradeDate).
		Pluck("symbol", &symbols).Error

	return symbols, err
}

// GetDateRange 获取指定股票的技术指标日期范围
func (r *TechnicalIndicatorRepository) GetDateRange(symbol string, period model.TechnicalIndicatorPeriod) (minDate, maxDate int, err error) {
	indicator := model.NewTechnicalIndicator(period)

	type DateRange struct {
		MinDate int `json:"min_date"`
		MaxDate int `json:"max_date"`
	}

	var result DateRange
	err = r.db.Table(indicator.TableName()).
		Select("MIN(trade_date) as min_date, MAX(trade_date) as max_date").
		Where("symbol = ?", symbol).
		Scan(&result).Error

	return result.MinDate, result.MaxDate, err
}
