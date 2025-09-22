package repository

import (
	"fmt"

	"stock/internal/model"

	"gorm.io/gorm"
)

// Shareholder 股东户数数据仓库
type Shareholder struct {
	db *gorm.DB
}

// NewShareholder 创建股东户数仓库实例
func NewShareholder(db *gorm.DB) *Shareholder {
	return &Shareholder{
		db: db,
	}
}

// Create 创建股东户数记录
func (r *Shareholder) Create(count *model.ShareholderCount) error {
	return r.db.Create(count).Error
}

// GetByTsCode 根据股票代码获取股东户数记录
func (r *Shareholder) GetByTsCode(tsCode string) ([]*model.ShareholderCount, error) {
	var counts []*model.ShareholderCount
	err := r.db.Where("ts_code = ?", tsCode).
		Order("end_date DESC").
		Find(&counts).Error
	return counts, err
}

// GetLatest 获取最新的股东户数记录
func (r *Shareholder) GetLatest(tsCode string) (*model.ShareholderCount, error) {
	var count model.ShareholderCount
	err := r.db.Where("ts_code = ?", tsCode).
		Order("end_date DESC").
		First(&count).Error
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// UpsertBatch 批量插入或更新股东户数记录
func (r *Shareholder) UpsertBatch(counts []*model.ShareholderCount) error {
	if len(counts) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, count := range counts {
			// 检查记录是否存在
			var existing model.ShareholderCount
			err := tx.Where("ts_code = ? AND end_date = ?", count.TsCode, count.EndDate).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				if err := tx.Create(count).Error; err != nil {
					return fmt.Errorf("创建股东户数记录失败: %v", err)
				}
			} else if err != nil {
				return fmt.Errorf("查询股东户数记录失败: %v", err)
			} else {
				// 记录存在，更新记录
				count.CreatedAt = existing.CreatedAt
				if err := tx.Save(count).Error; err != nil {
					return fmt.Errorf("更新股东户数记录失败: %v", err)
				}
			}
		}
		return nil
	})
}
