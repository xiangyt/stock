package service

import (
	"fmt"

	"stock/internal/collector"
	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/utils"
)

// ShareholderService 股东户数服务
type ShareholderService struct {
	repo      *repository.ShareholderRepository
	collector collector.DataCollector
}

// NewShareholderService 创建股东户数服务实例
func NewShareholderService(repo *repository.ShareholderRepository, collector collector.DataCollector) *ShareholderService {
	return &ShareholderService{
		repo:      repo,
		collector: collector,
	}
}

// SyncData 同步股东户数数据
func (s *ShareholderService) SyncData(tsCode string) error {
	// 确保股票代码格式正确
	tsCode = utils.ConvertToTsCode(tsCode)

	// 从采集器获取数据
	counts, err := s.collector.GetShareholderCounts(tsCode)
	if err != nil {
		return fmt.Errorf("获取股东户数数据失败: %v", err)
	}

	if len(counts) == 0 {
		return fmt.Errorf("未获取到股东户数数据")
	}

	// 转换为指针切片
	countPtrs := make([]*model.ShareholderCount, len(counts))
	for i, count := range counts {
		countPtrs[i] = &count
	}

	// 批量插入或更新数据
	if err := s.repo.UpsertBatch(countPtrs); err != nil {
		return fmt.Errorf("保存股东户数数据失败: %v", err)
	}

	return nil
}

// GetLatest 获取最新股东户数数据
func (s *ShareholderService) GetLatest(tsCode string) (*model.ShareholderCount, error) {
	tsCode = utils.ConvertToTsCode(tsCode)
	return s.repo.GetLatest(tsCode)
}

// GetByTsCode 获取指定股票的股东户数历史数据
func (s *ShareholderService) GetByTsCode(tsCode string) ([]*model.ShareholderCount, error) {
	tsCode = utils.ConvertToTsCode(tsCode)
	return s.repo.GetByTsCode(tsCode)
}
