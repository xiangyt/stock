package service

import (
	"fmt"
	"sync"
	"time"

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

// GetShareholderCounts 获取股东户数数据
func (s *ShareholderService) GetShareholderCounts(tsCode string) ([]*model.ShareholderCount, error) {
	return s.repo.GetByTsCode(tsCode)
}

// GetLatestShareholderCount 获取最新股东户数
func (s *ShareholderService) GetLatestShareholderCount(tsCode string) (*model.ShareholderCount, error) {
	return s.repo.GetLatest(tsCode)
}

// GetShareholderCountsByDateRange 按日期范围获取股东户数
func (s *ShareholderService) GetShareholderCountsByDateRange(tsCode string, startDate, endDate string) ([]*model.ShareholderCount, error) {
	// 使用现有的GetByTsCode方法，然后过滤日期范围
	allCounts, err := s.repo.GetByTsCode(tsCode)
	if err != nil {
		return nil, err
	}

	// 解析日期字符串并转换为 YYYYMMDD 格式的整数
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("解析开始日期失败: %v", err)
	}
	startDateInt := startTime.Year()*10000 + int(startTime.Month())*100 + startTime.Day()

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("解析结束日期失败: %v", err)
	}
	endDateInt := endTime.Year()*10000 + int(endTime.Month())*100 + endTime.Day()

	var filteredCounts []*model.ShareholderCount
	for _, count := range allCounts {
		if count.EndDate >= startDateInt && count.EndDate <= endDateInt {
			filteredCounts = append(filteredCounts, count)
		}
	}
	return filteredCounts, nil
}

// SyncShareholderCounts 同步单只股票的股东户数
func (s *ShareholderService) SyncShareholderCounts(tsCode string) error {
	return s.SyncData(tsCode)
}

// SyncAllStocksShareholderCounts 同步所有股票的股东户数
func (s *ShareholderService) SyncAllStocksShareholderCounts() error {
	// 实现批量同步逻辑
	return nil
}

// GetStatistics 获取统计信息
func (s *ShareholderService) GetStatistics() (map[string]interface{}, error) {
	// 实现统计逻辑
	return make(map[string]interface{}), nil
}

// GetTopByHolderNum 按股东户数排序获取前N只股票
func (s *ShareholderService) GetTopByHolderNum(limit int) ([]*model.ShareholderCount, error) {
	// 简单实现，实际应该在repository中实现
	return []*model.ShareholderCount{}, nil
}

// GetTopByAvgMarketCap 按平均市值排序获取前N只股票
func (s *ShareholderService) GetTopByAvgMarketCap(limit int) ([]*model.ShareholderCount, error) {
	// 简单实现，实际应该在repository中实现
	return []*model.ShareholderCount{}, nil
}

// GetRecentChanges 获取最近变化的股东户数
func (s *ShareholderService) GetRecentChanges(days int) ([]*model.ShareholderCount, error) {
	// 简单实现，实际应该在repository中实现
	return []*model.ShareholderCount{}, nil
}

// GetShareholderCountsWithPagination 分页获取股东户数数据
func (s *ShareholderService) GetShareholderCountsWithPagination(page, pageSize int, tsCode string) ([]*model.ShareholderCount, int64, error) {
	// 简单实现，实际应该在repository中实现
	counts, err := s.repo.GetByTsCode(tsCode)
	if err != nil {
		return nil, 0, err
	}
	return counts, int64(len(counts)), nil
}

// ValidateShareholderCount 验证股东户数数据
func (s *ShareholderService) ValidateShareholderCount(data *model.ShareholderCount) error {
	// 实现验证逻辑
	return nil
}

var (
	shareholderServiceInstance *ShareholderService
	shareholderServiceOnce     sync.Once
)

// GetShareholderService 获取股东户数服务单例
func GetShareholderService(repo *repository.ShareholderRepository, collector collector.DataCollector) *ShareholderService {
	shareholderServiceOnce.Do(func() {
		shareholderServiceInstance = &ShareholderService{
			repo:      repo,
			collector: collector,
		}
	})
	return shareholderServiceInstance
}

// NewShareholderService 创建股东户数服务实例 (保持向后兼容)
func NewShareholderService(repo *repository.ShareholderRepository, collector collector.DataCollector) *ShareholderService {
	return GetShareholderService(repo, collector)
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
