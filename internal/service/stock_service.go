package service

import (
	"fmt"
	"time"

	"stock/internal/collector"
	"stock/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StockService 股票数据服务
type StockService struct {
	db               *gorm.DB
	logger           *logrus.Logger
	collectorManager *collector.CollectorManager
}

// NewStockService 创建股票数据服务
func NewStockService(db *gorm.DB, logger *logrus.Logger, collectorManager *collector.CollectorManager) *StockService {
	return &StockService{
		db:               db,
		logger:           logger,
		collectorManager: collectorManager,
	}
}

// SyncAllStocks 同步所有股票数据到数据库
func (s *StockService) SyncAllStocks() error {
	s.logger.Info("Starting to sync all stocks to database...")

	// 获取采集器
	collector, err := s.collectorManager.GetCollector("eastmoney")
	if err != nil {
		return fmt.Errorf("eastmoney collector not found: %w", err)
	}

	// 从API获取股票列表
	stocks, err := collector.GetStockList()
	if err != nil {
		return fmt.Errorf("failed to get stock list from API: %w", err)
	}

	s.logger.Infof("Fetched %d stocks from API, starting to save to database...", len(stocks))

	// 批量保存到数据库
	savedCount := 0
	updatedCount := 0
	errorCount := 0

	for i, stock := range stocks {
		// 设置创建和更新时间
		now := time.Now()
		stock.CreatedAt = now
		stock.UpdatedAt = now

		// 使用UPSERT操作：如果存在则更新，不存在则创建
		result := s.db.Where("ts_code = ?", stock.TsCode).First(&model.Stock{})
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				if err := s.db.Create(&stock).Error; err != nil {
					s.logger.Errorf("Failed to create stock %s (%s): %v", stock.TsCode, stock.Name, err)
					errorCount++
				} else {
					savedCount++
				}
			} else {
				s.logger.Errorf("Database error for stock %s: %v", stock.TsCode, result.Error)
				errorCount++
			}
		} else {
			// 记录存在，更新记录
			if err := s.db.Model(&model.Stock{}).Where("ts_code = ?", stock.TsCode).Updates(map[string]interface{}{
				"symbol":     stock.Symbol,
				"name":       stock.Name,
				"area":       stock.Area,
				"industry":   stock.Industry,
				"market":     stock.Market,
				"list_date":  stock.ListDate,
				"is_active":  stock.IsActive,
				"updated_at": now,
			}).Error; err != nil {
				s.logger.Errorf("Failed to update stock %s (%s): %v", stock.TsCode, stock.Name, err)
				errorCount++
			} else {
				updatedCount++
			}
		}

		// 每处理100条记录输出一次进度
		if (i+1)%100 == 0 {
			s.logger.Infof("Progress: %d/%d processed (saved: %d, updated: %d, errors: %d)",
				i+1, len(stocks), savedCount, updatedCount, errorCount)
		}
	}

	s.logger.Infof("Stock sync completed: %d total, %d saved, %d updated, %d errors",
		len(stocks), savedCount, updatedCount, errorCount)

	return nil
}

// GetStockList 从数据库获取股票列表
func (s *StockService) GetStockList(limit, offset int) ([]model.Stock, int64, error) {
	var stocks []model.Stock
	var total int64

	// 获取总数
	if err := s.db.Model(&model.Stock{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	query := s.db.Where("is_active = ?", true).Order("ts_code ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&stocks).Error; err != nil {
		return nil, 0, err
	}

	return stocks, total, nil
}

// GetStockByCode 根据股票代码获取股票信息
func (s *StockService) GetStockByCode(tsCode string) (*model.Stock, error) {
	var stock model.Stock
	if err := s.db.Where("ts_code = ?", tsCode).First(&stock).Error; err != nil {
		return nil, err
	}
	return &stock, nil
}

// SearchStocks 搜索股票
func (s *StockService) SearchStocks(keyword string, limit int) ([]model.Stock, error) {
	var stocks []model.Stock

	query := s.db.Where("is_active = ?", true)
	if keyword != "" {
		query = query.Where("ts_code LIKE ? OR symbol LIKE ? OR name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("ts_code ASC").Find(&stocks).Error; err != nil {
		return nil, err
	}

	return stocks, nil
}

// GetStockStats 获取股票统计信息
func (s *StockService) GetStockStats() (map[string]interface{}, error) {
	var totalCount, activeCount int64

	// 总股票数
	if err := s.db.Model(&model.Stock{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// 活跃股票数
	if err := s.db.Model(&model.Stock{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		return nil, err
	}

	// 按市场统计
	var marketStats []struct {
		Market string `json:"market"`
		Count  int64  `json:"count"`
	}
	if err := s.db.Model(&model.Stock{}).
		Select("market, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("market").
		Find(&marketStats).Error; err != nil {
		return nil, err
	}

	// 按行业统计（前10）
	var industryStats []struct {
		Industry string `json:"industry"`
		Count    int64  `json:"count"`
	}
	if err := s.db.Model(&model.Stock{}).
		Select("industry, COUNT(*) as count").
		Where("is_active = ? AND industry != ''", true).
		Group("industry").
		Order("count DESC").
		Limit(10).
		Find(&industryStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_stocks":   totalCount,
		"active_stocks":  activeCount,
		"market_stats":   marketStats,
		"industry_stats": industryStats,
		"last_updated":   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// RefreshStockDetail 刷新单个股票的详细信息
func (s *StockService) RefreshStockDetail(tsCode string) (*model.Stock, error) {
	s.logger.Infof("Refreshing stock detail for %s", tsCode)

	// 获取采集器
	collector, err := s.collectorManager.GetCollector("eastmoney")
	if err != nil {
		return nil, fmt.Errorf("eastmoney collector not found: %w", err)
	}

	// 从API获取股票详情
	stockDetail, err := collector.GetStockDetail(tsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock detail from API: %w", err)
	}

	// 更新数据库
	now := time.Now()
	stockDetail.UpdatedAt = now

	if err := s.db.Model(&model.Stock{}).Where("ts_code = ?", tsCode).Updates(map[string]interface{}{
		"symbol":     stockDetail.Symbol,
		"name":       stockDetail.Name,
		"area":       stockDetail.Area,
		"industry":   stockDetail.Industry,
		"market":     stockDetail.Market,
		"list_date":  stockDetail.ListDate,
		"is_active":  stockDetail.IsActive,
		"updated_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update stock in database: %w", err)
	}

	s.logger.Infof("Successfully refreshed stock detail for %s", tsCode)
	return stockDetail, nil
}
