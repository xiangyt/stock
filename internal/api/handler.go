package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/internal/collector"
	"stock/internal/model"
	"stock/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler API处理器
type Handler struct {
	collectorManager *collector.CollectorManager
	logger           *logrus.Logger
	klineService     *service.KLineService
	stockService     *service.StockService
	taskService      *service.TaskService
	db               *gorm.DB
}

// NewHandler 创建新的API处理器
func NewHandler(collectorManager *collector.CollectorManager, logger *logrus.Logger, db *gorm.DB) *Handler {
	taskService := service.NewTaskService(db, logger)
	return &Handler{
		collectorManager: collectorManager,
		logger:           logger,
		klineService:     service.NewKLineService(db, logger, collectorManager),
		stockService:     service.NewStockService(db, logger, collectorManager),
		taskService:      taskService,
		db:               db,
	}
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// GetStockList 获取股票列表
func (h *Handler) GetStockList(c *gin.Context) {
	h.logger.Info("API: Getting stock list")

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 20
	}

	// 从东方财富获取股票列表
	stocks, err := h.collectorManager.GetStockListFromSource("eastmoney")
	if err != nil {
		h.logger.Errorf("Failed to get stock list: %v", err)
		Error(c, 1001, "获取股票列表失败")
		return
	}

	// 分页处理
	total := len(stocks)
	start := (page - 1) * size
	end := start + size

	if start >= total {
		Success(c, gin.H{
			"stocks": []interface{}{},
			"total":  total,
			"page":   page,
			"size":   size,
		})
		return
	}

	if end > total {
		end = total
	}

	pagedStocks := stocks[start:end]

	Success(c, gin.H{
		"stocks": pagedStocks,
		"total":  total,
		"page":   page,
		"size":   size,
	})
}

// GetStockDetail 获取股票详情
func (h *Handler) GetStockDetail(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	h.logger.Infof("API: Getting stock detail for %s", tsCode)

	// 获取股票详情
	collector, err := h.collectorManager.GetCollector("eastmoney")
	if err != nil {
		h.logger.Errorf("Failed to get EastMoney collector: %v", err)
		Error(c, 1004, "数据源不可用")
		return
	}

	stock, err := collector.GetStockDetail(tsCode)
	if err != nil {
		h.logger.Errorf("Failed to get stock detail: %v", err)
		Error(c, 1004, "获取股票详情失败")
		return
	}

	Success(c, stock)
}

// GetKLineData 获取K线数据（只从数据库查询，不刷新）
func (h *Handler) GetKLineData(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	// 获取查询参数
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 1000 {
		days = 30
	}

	h.logger.Infof("API: Getting K-line data from database for %s (last %d days)", tsCode, days)

	// 计算时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// 只从数据库获取K线数据
	klineData, err := h.klineService.GetKLineData(tsCode, startDate, endDate)
	if err != nil {
		h.logger.Errorf("Failed to get K-line data from database: %v", err)
		Error(c, 1005, "获取K线数据失败")
		return
	}

	Success(c, gin.H{
		"code":   tsCode,
		"days":   days,
		"count":  len(klineData),
		"kline":  klineData,
		"start":  startDate.Format("2006-01-02"),
		"end":    endDate.Format("2006-01-02"),
		"source": "database", // 数据来源：数据库
	})
}

// RefreshKLineData 刷新K线数据（从API获取并保存到数据库）
func (h *Handler) RefreshKLineData(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	// 获取查询参数
	daysStr := c.DefaultQuery("days", "365") // 默认刷新一年数据
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 2000 {
		days = 365
	}

	h.logger.Infof("API: Refreshing K-line data from API for %s (last %d days)", tsCode, days)

	// 计算时间范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// 从API刷新K线数据
	klineData, err := h.klineService.RefreshKLineData(tsCode, startDate, endDate)
	if err != nil {
		h.logger.Errorf("Failed to refresh K-line data: %v", err)
		Error(c, 1005, "刷新K线数据失败")
		return
	}

	Success(c, gin.H{
		"code":      tsCode,
		"days":      days,
		"count":     len(klineData),
		"start":     startDate.Format("2006-01-02"),
		"end":       endDate.Format("2006-01-02"),
		"source":    "api",
		"message":   "数据已从API刷新并保存到数据库",
		"refreshed": true,
	})
}

// GetKLineDataRange 获取数据库中K线数据的范围信息
func (h *Handler) GetKLineDataRange(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	h.logger.Infof("API: Getting K-line data range for %s", tsCode)

	startDate, endDate, count, err := h.klineService.GetDataRange(tsCode)
	if err != nil {
		h.logger.Errorf("Failed to get data range: %v", err)
		Error(c, 1005, "获取数据范围失败")
		return
	}

	if count == 0 {
		Success(c, gin.H{
			"code":    tsCode,
			"count":   0,
			"message": "数据库中没有该股票的K线数据",
		})
		return
	}

	Success(c, gin.H{
		"code":       tsCode,
		"count":      count,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"message":    "数据库中的K线数据范围",
	})
}

// CheckKLineDataFreshness 检查K线数据新鲜度
func (h *Handler) CheckKLineDataFreshness(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	h.logger.Infof("API: Checking K-line data freshness for %s", tsCode)

	freshness, err := h.klineService.CheckDataFreshness(tsCode)
	if err != nil {
		h.logger.Errorf("Failed to check data freshness: %v", err)
		Error(c, 1005, "检查数据新鲜度失败")
		return
	}

	Success(c, freshness)
}

// GetRealtimeData 获取实时数据
func (h *Handler) GetRealtimeData(c *gin.Context) {
	codesParam := c.Query("codes")
	if codesParam == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 解析股票代码列表
	codes := strings.Split(codesParam, ",")
	var tsCodes []string
	for _, code := range codes {
		code = strings.TrimSpace(strings.ToUpper(code))
		if code != "" && strings.Contains(code, ".") {
			tsCodes = append(tsCodes, code)
		}
	}

	if len(tsCodes) == 0 {
		Error(c, 1003, "没有有效的股票代码")
		return
	}

	h.logger.Infof("API: Getting realtime data for %d stocks", len(tsCodes))

	// 获取实时数据
	collector, err := h.collectorManager.GetCollector("eastmoney")
	if err != nil {
		h.logger.Errorf("Failed to get EastMoney collector: %v", err)
		Error(c, 1007, "数据源不可用")
		return
	}

	realtimeData, err := collector.GetRealtimeData(tsCodes)
	if err != nil {
		h.logger.Errorf("Failed to get realtime data: %v", err)
		Error(c, 1007, "获取实时数据失败")
		return
	}

	Success(c, gin.H{
		"codes":    tsCodes,
		"count":    len(realtimeData),
		"realtime": realtimeData,
	})
}

// SyncAllStocks 同步所有股票数据到数据库
func (h *Handler) SyncAllStocks(c *gin.Context) {
	h.logger.Info("API: Starting to sync all stocks to database")

	// 异步执行同步操作
	go func() {
		if err := h.stockService.SyncAllStocks(); err != nil {
			h.logger.Errorf("Failed to sync all stocks: %v", err)
		}
	}()

	Success(c, gin.H{
		"message": "Stock synchronization started",
		"status":  "running",
	})
}

// GetStockStats 获取股票统计信息
func (h *Handler) GetStockStats(c *gin.Context) {
	h.logger.Info("API: Getting stock statistics")

	stats, err := h.stockService.GetStockStats()
	if err != nil {
		h.logger.Errorf("Failed to get stock stats: %v", err)
		Error(c, 1006, "获取股票统计信息失败")
		return
	}

	Success(c, stats)
}

// SearchStocks 搜索股票
func (h *Handler) SearchStocks(c *gin.Context) {
	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	h.logger.Infof("API: Searching stocks with keyword: %s, limit: %d", keyword, limit)

	stocks, err := h.stockService.SearchStocks(keyword, limit)
	if err != nil {
		h.logger.Errorf("Failed to search stocks: %v", err)
		Error(c, 1007, "搜索股票失败")
		return
	}

	Success(c, gin.H{
		"keyword": keyword,
		"count":   len(stocks),
		"stocks":  stocks,
	})
}

// ===== 异步任务相关API =====

// SyncAllStocksAsync 异步同步全量股票数据
func (h *Handler) SyncAllStocksAsync(c *gin.Context) {
	h.logger.Info("API: Starting async sync of all stocks")

	// 创建异步任务
	task, err := h.taskService.CreateTask(model.TaskTypeSyncAllStocks, map[string]interface{}{
		"source": "api_request",
	})
	if err != nil {
		h.logger.Errorf("Failed to create sync task: %v", err)
		Error(c, 1005, "创建同步任务失败")
		return
	}

	// 启动异步执行
	h.taskService.StartTask(task.ID, func(ctx context.Context, task *model.Task, updateProgress func(int, string)) error {
		return h.stockService.SyncAllStocks()
	})

	Success(c, gin.H{
		"task_id": task.ID,
		"message": "股票数据同步任务已启动",
		"status":  "running",
	})
}

// GetTaskStatus 获取任务状态
func (h *Handler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		Error(c, 1001, "任务ID不能为空")
		return
	}

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		h.logger.Errorf("Failed to get task %s: %v", taskID, err)
		Error(c, 1004, "任务不存在")
		return
	}

	Success(c, task)
}

// ListTasks 获取任务列表
func (h *Handler) ListTasks(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	statusStr := c.DefaultQuery("status", "")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var status model.TaskStatus
	if statusStr != "" {
		status = model.TaskStatus(statusStr)
	}

	tasks, total, err := h.taskService.ListTasks(limit, offset, status)
	if err != nil {
		h.logger.Errorf("Failed to list tasks: %v", err)
		Error(c, 1005, "获取任务列表失败")
		return
	}

	Success(c, gin.H{
		"tasks":  tasks,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CancelTask 取消任务
func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		Error(c, 1001, "任务ID不能为空")
		return
	}

	err := h.taskService.CancelTask(taskID)
	if err != nil {
		h.logger.Errorf("Failed to cancel task %s: %v", taskID, err)
		Error(c, 1005, "取消任务失败")
		return
	}

	Success(c, gin.H{
		"task_id": taskID,
		"message": "任务已取消",
	})
}

// SyncSingleStockAsync 异步刷新单只股票的日K数据
func (h *Handler) SyncSingleStockAsync(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, 1002, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		Error(c, 1003, "股票代码格式错误，应为：000001.SZ 或 600000.SH")
		return
	}

	h.logger.Infof("API: Starting async sync of single stock: %s", tsCode)

	// 创建异步任务
	task, err := h.taskService.CreateTask(model.TaskTypeSyncSingleStock, map[string]interface{}{
		"source":    "api_request",
		"stockCode": tsCode,
	})
	if err != nil {
		h.logger.Errorf("Failed to create sync task: %v", err)
		Error(c, 1005, "创建同步任务失败")
		return
	}

	// 启动异步执行
	h.taskService.StartTask(task.ID, func(ctx context.Context, task *model.Task, updateProgress func(int, string)) error {
		// 设置时间范围：获取最近1年的数据
		endDate := time.Now()
		startDate := endDate.AddDate(-1, 0, 0) // 1年前

		// 更新进度
		updateProgress(10, "开始获取日K数据")

		// 从API获取日K数据
		dailyData, err := h.klineService.RefreshKLineData(tsCode, startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to refresh K-line data: %w", err)
		}

		// 更新进度
		updateProgress(100, fmt.Sprintf("成功获取 %d 条日K数据", len(dailyData)))

		return nil
	})

	Success(c, gin.H{
		"task_id": task.ID,
		"message": "单只股票日K数据同步任务已启动",
		"status":  "running",
		"stock":   tsCode,
	})
}

// GetPerformanceReports 获取业绩报表数据
func (h *Handler) GetPerformanceReports(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换为标准格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		// 如果没有交易所后缀，尝试添加
		if len(code) == 6 {
			if code[0] == '6' {
				tsCode = code + ".SH"
			} else {
				tsCode = code + ".SZ"
			}
		}
	}

	h.logger.Infof("Getting performance reports for stock: %s", tsCode)

	// 从数据库查询业绩报表数据
	var reports []model.PerformanceReport
	if err := h.db.Where("ts_code = ?", tsCode).Order("report_date DESC").Find(&reports).Error; err != nil {
		h.logger.Errorf("Failed to query performance reports from database: %v", err)
		Error(c, http.StatusInternalServerError, "查询业绩报表数据失败")
		return
	}

	// 如果数据库中没有数据，尝试从数据源获取
	if len(reports) == 0 {
		collector, err := h.collectorManager.GetCollector("eastmoney")
		if err != nil {
			h.logger.Errorf("Failed to get collector: %v", err)
			Error(c, http.StatusInternalServerError, "获取数据采集器失败")
			return
		}

		reports, err = collector.GetPerformanceReports(tsCode)
		if err != nil {
			h.logger.Errorf("Failed to get performance reports from collector: %v", err)
			Error(c, http.StatusInternalServerError, "获取业绩报表数据失败")
			return
		}

		// 保存到数据库
		if len(reports) > 0 {
			for i := range reports {
				reports[i].CreatedAt = time.Now()
				reports[i].UpdatedAt = time.Now()
			}
			if err := h.db.Create(&reports).Error; err != nil {
				h.logger.Warnf("Failed to save performance reports to database: %v", err)
			}
		}
	}

	Success(c, gin.H{
		"ts_code": tsCode,
		"count":   len(reports),
		"reports": reports,
	})
}

// GetLatestPerformanceReport 获取最新业绩报表数据
func (h *Handler) GetLatestPerformanceReport(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换为标准格式
	tsCode := strings.ToUpper(code)
	if !strings.Contains(tsCode, ".") {
		// 如果没有交易所后缀，尝试添加
		if len(code) == 6 {
			if code[0] == '6' {
				tsCode = code + ".SH"
			} else {
				tsCode = code + ".SZ"
			}
		}
	}

	h.logger.Infof("Getting latest performance report for stock: %s", tsCode)

	// 从数据库查询最新业绩报表数据
	var report model.PerformanceReport
	if err := h.db.Where("ts_code = ?", tsCode).Order("report_date DESC").First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 数据库中没有数据，尝试从数据源获取
			collector, err := h.collectorManager.GetCollector("eastmoney")
			if err != nil {
				h.logger.Errorf("Failed to get collector: %v", err)
				Error(c, http.StatusInternalServerError, "获取数据采集器失败")
				return
			}

			latestReport, err := collector.GetLatestPerformanceReport(tsCode)
			if err != nil {
				h.logger.Errorf("Failed to get latest performance report from collector: %v", err)
				Error(c, http.StatusInternalServerError, "获取最新业绩报表数据失败")
				return
			}

			// 保存到数据库
			latestReport.CreatedAt = time.Now()
			latestReport.UpdatedAt = time.Now()
			if err := h.db.Create(latestReport).Error; err != nil {
				h.logger.Warnf("Failed to save latest performance report to database: %v", err)
			}

			report = *latestReport
		} else {
			h.logger.Errorf("Failed to query latest performance report from database: %v", err)
			Error(c, http.StatusInternalServerError, "查询最新业绩报表数据失败")
			return
		}
	}

	Success(c, gin.H{
		"ts_code": tsCode,
		"report":  report,
	})
}
