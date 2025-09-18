# Service单例模式实现文档

## 概述

本文档描述了股票系统中所有service的单例模式实现，确保在整个应用程序生命周期中每个service只有一个实例。

## 实现的单例服务

### 1. DataService (数据服务)
- **获取方法**: `GetDataService(db *gorm.DB, logger *utils.Logger)`
- **向后兼容**: `NewDataService()` 内部调用 `GetDataService()`
- **功能**: 股票数据同步、K线数据管理、实时数据更新

### 2. KLinePersistenceService (K线数据持久化服务)
- **获取方法**: `GetKLinePersistenceService(db *gorm.DB, logger *utils.Logger)`
- **向后兼容**: `NewKLinePersistenceService()` 内部调用 `GetKLinePersistenceService()`
- **功能**: 日/周/月/年K线数据的保存和查询

### 3. StockService (股票数据服务)
- **获取方法**: `GetStockService(db *gorm.DB, logger *logrus.Logger, collectorManager *collector.CollectorManager)`
- **向后兼容**: `NewStockService()` 内部调用 `GetStockService()`
- **功能**: 股票基础信息管理、股票列表同步

### 4. KLineService (K线数据服务)
- **获取方法**: `GetKLineService(db *gorm.DB, logger *logrus.Logger, collectorManager *collector.CollectorManager)`
- **向后兼容**: `NewKLineService()` 内部调用 `GetKLineService()`
- **功能**: K线数据获取、刷新、统计

### 5. PerformanceService (业绩报表服务)
- **获取方法**: `GetPerformanceService(repo, stockRepo, collector, logger)`
- **向后兼容**: `NewPerformanceService()` 内部调用 `GetPerformanceService()`
- **功能**: 业绩报表数据采集和管理

### 6. TaskService (任务服务)
- **获取方法**: `GetTaskService(db *gorm.DB, logger *logrus.Logger)`
- **向后兼容**: `NewTaskService()` 内部调用 `GetTaskService()`
- **功能**: 异步任务管理和执行

### 7. ShareholderService (股东户数服务)
- **获取方法**: `GetShareholderService(repo, collector)`
- **向后兼容**: `NewShareholderService()` 内部调用 `GetShareholderService()`
- **功能**: 股东户数数据采集和管理

### 8. DailyKLineManager (日K线数据管理器)
- **获取方法**: `GetDailyKLineManager(db *gorm.DB, logger *utils.Logger)`
- **向后兼容**: `NewDailyKLineManager()` 内部调用 `GetDailyKLineManager()`
- **功能**: 按交易所分表管理日K线数据

### 9. DatabaseService (数据库服务)
- **获取方法**: `GetDatabaseService(cfg *config.Config, logger *utils.Logger)`
- **向后兼容**: `NewDatabaseService()` 内部调用 `GetDatabaseService()`
- **功能**: 数据库初始化和迁移

### 10. DataCollectorService (数据采集服务)
- **获取方法**: `GetDataCollectorService(cfg *config.Config, logger *utils.Logger)`
- **向后兼容**: `NewDataCollectorService()` 内部调用 `GetDataCollectorService()`
- **功能**: 数据采集任务管理

### 11. TechnicalAnalyzerService (技术分析服务)
- **获取方法**: `GetTechnicalAnalyzerService(cfg *config.Config, logger *utils.Logger)`
- **向后兼容**: `NewTechnicalAnalyzerService()` 内部调用 `GetTechnicalAnalyzerService()`
- **功能**: 技术指标计算

### 12. StrategyEngineService (策略引擎服务)
- **获取方法**: `GetStrategyEngineService(cfg *config.Config, logger *utils.Logger)`
- **向后兼容**: `NewStrategyEngineService()` 内部调用 `GetStrategyEngineService()`
- **功能**: 选股策略执行

### 13. BacktestEngineService (回测引擎服务)
- **获取方法**: `GetBacktestEngineService(cfg *config.Config, logger *utils.Logger)`
- **向后兼容**: `NewBacktestEngineService()` 内部调用 `GetBacktestEngineService()`
- **功能**: 策略回测

## 实现原理

### 单例模式实现
每个service都使用以下模式实现单例：

```go
var (
    serviceInstance *ServiceType
    serviceOnce     sync.Once
)

func GetService(params...) *ServiceType {
    serviceOnce.Do(func() {
        serviceInstance = &ServiceType{
            // 初始化字段
        }
    })
    return serviceInstance
}

func NewService(params...) *ServiceType {
    return GetService(params...)
}
```

### 关键特性

1. **线程安全**: 使用 `sync.Once` 确保只初始化一次
2. **延迟初始化**: 只在第一次调用时创建实例
3. **向后兼容**: 保留原有的 `New*` 函数，内部调用单例方法
4. **内存效率**: 整个应用程序生命周期中只有一个实例

## 使用示例

### 推荐用法（使用单例方法）
```go
// 获取数据服务单例
dataService := service.GetDataService(db, logger)

// 获取K线持久化服务单例
klineService := service.GetKLinePersistenceService(db, logger)
```

### 兼容用法（原有方法仍可用）
```go
// 这些方法内部会调用单例方法
dataService := service.NewDataService(db, logger)
klineService := service.NewKLinePersistenceService(db, logger)
```

## 测试验证

### 单例测试
- 文件: `test/unit/service_singleton_simple_test.go`
- 验证内容:
  - 多次获取同一service返回相同实例
  - 新旧方法返回相同实例
  - Services集合中的服务为单例

### 测试结果
```
=== RUN   TestServiceSingletonSimple
=== RUN   TestServiceSingletonSimple/TestDatabaseServiceSingleton
    DatabaseService单例模式验证成功，实例地址: 0x140001b1e30
=== RUN   TestServiceSingletonSimple/TestDataCollectorServiceSingleton
    DataCollectorService单例模式验证成功，实例地址: 0x140001b1ef0
...
--- PASS: TestServiceSingletonSimple (0.00s)
```

## 优势

1. **内存节省**: 避免重复创建相同功能的service实例
2. **状态一致**: 所有地方使用的都是同一个service实例，状态一致
3. **性能提升**: 减少对象创建和垃圾回收的开销
4. **配置统一**: 单例确保配置和依赖的一致性
5. **向后兼容**: 不破坏现有代码，平滑迁移

## 注意事项

1. **初始化参数**: 单例只在第一次调用时初始化，后续调用的参数会被忽略
2. **状态管理**: 由于是单例，需要注意并发访问时的状态管理
3. **测试隔离**: 在单元测试中可能需要考虑单例对测试隔离的影响
4. **依赖注入**: 确保依赖的服务也正确初始化

## 迁移指南

### 现有代码迁移
1. 将 `New*Service()` 调用改为 `Get*Service()` （可选，向后兼容）
2. 确保在应用启动时正确初始化所有依赖
3. 移除重复的service创建代码

### 新代码开发
1. 优先使用 `Get*Service()` 方法
2. 在需要service的地方直接获取单例
3. 避免在多个地方重复创建相同的service

## 总结

通过实现单例模式，股票系统的service层现在具有更好的内存效率和状态一致性。所有service都支持单例访问，同时保持向后兼容性，确保现有代码无需修改即可继续工作。