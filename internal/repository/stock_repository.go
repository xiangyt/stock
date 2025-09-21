package repository

import (
	"stock/internal/logger"
	"stock/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StockRepository 股票数据仓库
type StockRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewStockRepository 创建股票仓库
func NewStockRepository(db *gorm.DB, logger *logger.Logger) *StockRepository {
	return &StockRepository{
		db:     db,
		logger: logger,
	}
}

// CreateStock 创建股票记录
func (r *StockRepository) CreateStock(stock *model.Stock) error {
	if err := r.db.Create(stock).Error; err != nil {
		r.logger.Errorf("Failed to create stock %s: %v", stock.TsCode, err)
		return err
	}
	r.logger.Debugf("Created stock: %s", stock.TsCode)
	return nil
}

// BatchCreateStocks 批量创建股票记录
func (r *StockRepository) BatchCreateStocks(stocks []model.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	// 使用事务批量插入
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 批量插入，每次处理1000条记录
	batchSize := 1000
	for i := 0; i < len(stocks); i += batchSize {
		end := i + batchSize
		if end > len(stocks) {
			end = len(stocks)
		}

		batch := stocks[i:end]
		if err := tx.CreateInBatches(batch, batchSize).Error; err != nil {
			tx.Rollback()
			r.logger.Errorf("Failed to batch create stocks: %v", err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Errorf("Failed to commit batch create stocks: %v", err)
		return err
	}

	r.logger.Infof("Successfully batch created %d stocks", len(stocks))
	return nil
}

// UpsertStocks 批量更新或插入股票记录
func (r *StockRepository) UpsertStocks(stocks []model.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用批量处理，每次处理100条记录
	batchSize := 100
	for i := 0; i < len(stocks); i += batchSize {
		end := i + batchSize
		if end > len(stocks) {
			end = len(stocks)
		}

		batch := stocks[i:end]

		// 使用Clauses来实现ON DUPLICATE KEY UPDATE
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ts_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "area", "industry", "market", "list_date", "is_active", "updated_at"}),
		}).Create(&batch).Error; err != nil {
			tx.Rollback()
			r.logger.Errorf("Failed to upsert stock batch: %v", err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.Errorf("Failed to commit upsert stocks: %v", err)
		return err
	}

	r.logger.Infof("Successfully upserted %d stocks", len(stocks))
	return nil
}

// GetStockByTsCode 根据股票代码获取股票信息
func (r *StockRepository) GetStockByTsCode(tsCode string) (*model.Stock, error) {
	var stock model.Stock
	if err := r.db.Where("ts_code = ?", tsCode).First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Errorf("Failed to get stock %s: %v", tsCode, err)
		return nil, err
	}
	return &stock, nil
}

// GetAllStocks 获取所有股票列表
func (r *StockRepository) GetAllStocks() ([]model.Stock, error) {
	var stocks []model.Stock
	if err := r.db.Where("is_active = ?", true).Find(&stocks).Error; err != nil {
		r.logger.Errorf("Failed to get all stocks: %v", err)
		return nil, err
	}
	return stocks, nil
}

// GetStocksByMarket 根据市场获取股票列表
func (r *StockRepository) GetStocksByMarket(market string) ([]model.Stock, error) {
	var stocks []model.Stock
	if err := r.db.Where("market = ? AND is_active = ?", market, true).Find(&stocks).Error; err != nil {
		r.logger.Errorf("Failed to get stocks by market %s: %v", market, err)
		return nil, err
	}
	return stocks, nil
}

// GetStocksByIndustry 根据行业获取股票列表
func (r *StockRepository) GetStocksByIndustry(industry string) ([]model.Stock, error) {
	var stocks []model.Stock
	if err := r.db.Where("industry = ? AND is_active = ?", industry, true).Find(&stocks).Error; err != nil {
		r.logger.Errorf("Failed to get stocks by industry %s: %v", industry, err)
		return nil, err
	}
	return stocks, nil
}

// UpdateStock 更新股票信息
func (r *StockRepository) UpdateStock(stock *model.Stock) error {
	if err := r.db.Save(stock).Error; err != nil {
		r.logger.Errorf("Failed to update stock %s: %v", stock.TsCode, err)
		return err
	}
	r.logger.Debugf("Updated stock: %s", stock.TsCode)
	return nil
}

// DeleteStock 删除股票记录
func (r *StockRepository) DeleteStock(tsCode string) error {
	if err := r.db.Where("ts_code = ?", tsCode).Delete(&model.Stock{}).Error; err != nil {
		r.logger.Errorf("Failed to delete stock %s: %v", tsCode, err)
		return err
	}
	r.logger.Debugf("Deleted stock: %s", tsCode)
	return nil
}

// GetStockCount 获取股票总数
func (r *StockRepository) GetStockCount() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Stock{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		r.logger.Errorf("Failed to get stock count: %v", err)
		return 0, err
	}
	return count, nil
}

// SearchStocks 搜索股票
func (r *StockRepository) SearchStocks(keyword string, limit int) ([]model.Stock, error) {
	var stocks []model.Stock
	query := r.db.Where("is_active = ?", true)

	if keyword != "" {
		query = query.Where("ts_code LIKE ? OR symbol LIKE ? OR name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&stocks).Error; err != nil {
		r.logger.Errorf("Failed to search stocks with keyword %s: %v", keyword, err)
		return nil, err
	}

	return stocks, nil
}
