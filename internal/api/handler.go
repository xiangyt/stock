package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/internal/collector"
	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler API处理器
type Handler struct {
	collectorFactory *collector.CollectorFactory
	logger           *logrus.Logger
	klineService     *service.KLineService
	stockService     *service.StockService
	taskService      *service.TaskService
	userService      *service.UserService
	roleRepo         *repository.RoleRepository
	db               *gorm.DB
}

// NewHandler 创建新的API处理器
func NewHandler(collectorFactory *collector.CollectorFactory, logger *logrus.Logger, db *gorm.DB) *Handler {
	taskService := service.NewTaskService(db, logger)

	// 创建用户相关的repository
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := service.NewUserService(userRepo, roleRepo)

	return &Handler{
		collectorFactory: collectorFactory,
		logger:           logger,
		klineService:     service.NewKLineService(db, logger, collectorFactory),
		stockService:     service.NewStockService(db, logger, collectorFactory),
		taskService:      taskService,
		userService:      userService,
		roleRepo:         roleRepo,
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
func Error(c *gin.Context, httpCode int, message string, details ...string) {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
		Data:    detail,
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

	// 从数据库获取股票列表
	stocks, total, err := h.stockService.GetStockListWithPagination(page, size)
	if err != nil {
		h.logger.Errorf("Failed to get stock list from database: %v", err)
		Error(c, 1001, "获取股票列表失败")
		return
	}

	Success(c, gin.H{
		"stocks": stocks,
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
	collector, err := h.collectorFactory.GetCollector("eastmoney")
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
	collector, err := h.collectorFactory.GetCollector("eastmoney")
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
		collector, err := h.collectorFactory.GetCollector("eastmoney")
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
			collector, err := h.collectorFactory.GetCollector("eastmoney")
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

// ===== 认证模块 =====

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	Success(c, gin.H{
		"message": "Login endpoint - TODO: implement",
		"token":   "mock_token_123",
	})
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	Success(c, gin.H{
		"message": "Logout successful",
	})
}

// RefreshToken 刷新令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	Success(c, gin.H{
		"message": "Token refreshed",
		"token":   "new_mock_token_456",
	})
}

// GetProfile 获取用户资料
func (h *Handler) GetProfile(c *gin.Context) {
	Success(c, gin.H{
		"message": "User profile - TODO: implement",
		"user": gin.H{
			"id":       1,
			"username": "admin",
			"email":    "admin@example.com",
		},
	})
}

// ===== 用户管理模块 =====

// GetUsers 获取用户列表
func (h *Handler) GetUsers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	keyword := c.Query("keyword")

	pageInt := 1
	pageSizeInt := 10

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
		pageSizeInt = ps
	}

	result, err := h.userService.GetUserList(pageInt, pageSizeInt, keyword)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}

	Success(c, gin.H{
		"users":     result.Users,
		"total":     result.Total,
		"page":      pageInt,
		"page_size": pageSizeInt,
	})
}

// CreateUser 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建用户失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "用户创建成功",
		"user":    user,
	})
}

// GetUser 获取单个用户
func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		Error(c, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	Success(c, gin.H{
		"user": user,
	})
}

// UpdateUser 更新用户
func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	user, err := h.userService.UpdateUser(uint(id), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新用户失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "用户更新成功",
		"user":    user,
	})
}

// DeleteUser 删除用户
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	err = h.userService.DeleteUser(uint(id))
	if err != nil {
		Error(c, http.StatusInternalServerError, "删除用户失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "用户删除成功",
	})
}

// GetUserLogs 获取用户日志
func (h *Handler) GetUserLogs(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message": "User logs - TODO: implement",
		"user_id": id,
		"logs":    []gin.H{},
	})
}

// ===== 股票数据模块扩展 =====

// GetFinancialData 获取财务数据
func (h *Handler) GetFinancialData(c *gin.Context) {
	code := c.Param("code")
	Success(c, gin.H{
		"message": "Financial data - TODO: implement",
		"code":    code,
		"data":    gin.H{},
	})
}

// AddToWatchlist 添加到关注列表
func (h *Handler) AddToWatchlist(c *gin.Context) {
	code := c.Param("code")
	Success(c, gin.H{
		"message": "Added to watchlist - TODO: implement",
		"code":    code,
	})
}

// RemoveFromWatchlist 从关注列表移除
func (h *Handler) RemoveFromWatchlist(c *gin.Context) {
	code := c.Param("code")
	Success(c, gin.H{
		"message": "Removed from watchlist - TODO: implement",
		"code":    code,
	})
}

// ===== 策略管理模块 =====

// GetStrategies 获取策略列表
func (h *Handler) GetStrategies(c *gin.Context) {
	Success(c, gin.H{
		"message":    "Strategies list - TODO: implement",
		"strategies": []gin.H{},
		"total":      0,
	})
}

// CreateStrategy 创建策略
func (h *Handler) CreateStrategy(c *gin.Context) {
	Success(c, gin.H{
		"message":     "Strategy created - TODO: implement",
		"strategy_id": 1,
	})
}

// GetStrategy 获取单个策略
func (h *Handler) GetStrategy(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Strategy detail - TODO: implement",
		"strategy_id": id,
	})
}

// UpdateStrategy 更新策略
func (h *Handler) UpdateStrategy(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Strategy updated - TODO: implement",
		"strategy_id": id,
	})
}

// DeleteStrategy 删除策略
func (h *Handler) DeleteStrategy(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Strategy deleted - TODO: implement",
		"strategy_id": id,
	})
}

// RunBacktest 运行回测
func (h *Handler) RunBacktest(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Backtest started - TODO: implement",
		"strategy_id": id,
		"task_id":     "bt_123",
	})
}

// GetStrategyResults 获取策略结果
func (h *Handler) GetStrategyResults(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Strategy results - TODO: implement",
		"strategy_id": id,
		"results":     []gin.H{},
	})
}

// ExecuteStrategy 执行策略
func (h *Handler) ExecuteStrategy(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":     "Strategy executed - TODO: implement",
		"strategy_id": id,
		"task_id":     "exec_123",
	})
}

// ===== 选股结果模块 =====

// GetSelections 获取选股结果
func (h *Handler) GetSelections(c *gin.Context) {
	Success(c, gin.H{
		"message":    "Selections list - TODO: implement",
		"selections": []gin.H{},
		"total":      0,
	})
}

// GetTodaySelections 获取今日选股
func (h *Handler) GetTodaySelections(c *gin.Context) {
	Success(c, gin.H{
		"message":    "Today selections - TODO: implement",
		"selections": []gin.H{},
		"date":       "2024-01-01",
	})
}

// GetSelectionHistory 获取选股历史
func (h *Handler) GetSelectionHistory(c *gin.Context) {
	Success(c, gin.H{
		"message": "Selection history - TODO: implement",
		"history": []gin.H{},
		"total":   0,
	})
}

// CreateSelection 创建选股结果
func (h *Handler) CreateSelection(c *gin.Context) {
	Success(c, gin.H{
		"message":      "Selection created - TODO: implement",
		"selection_id": 1,
	})
}

// GetSelection 获取单个选股结果
func (h *Handler) GetSelection(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Selection detail - TODO: implement",
		"selection_id": id,
	})
}

// DeleteSelection 删除选股结果
func (h *Handler) DeleteSelection(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Selection deleted - TODO: implement",
		"selection_id": id,
	})
}

// ExportSelections 导出选股结果
func (h *Handler) ExportSelections(c *gin.Context) {
	Success(c, gin.H{
		"message":  "Export started - TODO: implement",
		"task_id":  "export_123",
		"download": "/api/v1/downloads/export_123.xlsx",
	})
}

// ===== 数据采集模块扩展 =====

// GetCollectors 获取采集器列表
func (h *Handler) GetCollectors(c *gin.Context) {
	Success(c, gin.H{
		"message":    "Collectors list - TODO: implement",
		"collectors": []gin.H{},
	})
}

// GetCollectorStatus 获取采集器状态
func (h *Handler) GetCollectorStatus(c *gin.Context) {
	Success(c, gin.H{
		"message": "Collector status - TODO: implement",
		"status":  "running",
	})
}

// SyncData 同步数据
func (h *Handler) SyncData(c *gin.Context) {
	Success(c, gin.H{
		"message": "Data sync started - TODO: implement",
		"task_id": "sync_123",
	})
}

// SyncDataFromSource 从指定数据源同步
func (h *Handler) SyncDataFromSource(c *gin.Context) {
	source := c.Param("source")
	Success(c, gin.H{
		"message": "Source sync started - TODO: implement",
		"source":  source,
		"task_id": "sync_" + source + "_123",
	})
}

// GetCollectorTasks 获取采集任务
func (h *Handler) GetCollectorTasks(c *gin.Context) {
	Success(c, gin.H{
		"message": "Collector tasks - TODO: implement",
		"tasks":   []gin.H{},
		"total":   0,
	})
}

// CreateCollectorTask 创建采集任务
func (h *Handler) CreateCollectorTask(c *gin.Context) {
	Success(c, gin.H{
		"message": "Task created - TODO: implement",
		"task_id": 1,
	})
}

// UpdateCollectorTask 更新采集任务
func (h *Handler) UpdateCollectorTask(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message": "Task updated - TODO: implement",
		"task_id": id,
	})
}

// DeleteCollectorTask 删除采集任务
func (h *Handler) DeleteCollectorTask(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message": "Task deleted - TODO: implement",
		"task_id": id,
	})
}

// ===== 通知管理模块 =====

// GetRobots 获取机器人列表
func (h *Handler) GetRobots(c *gin.Context) {
	Success(c, gin.H{
		"message": "Robots list - TODO: implement",
		"robots":  []gin.H{},
		"total":   0,
	})
}

// CreateRobot 创建机器人
func (h *Handler) CreateRobot(c *gin.Context) {
	Success(c, gin.H{
		"message":  "Robot created - TODO: implement",
		"robot_id": 1,
	})
}

// GetRobot 获取单个机器人
func (h *Handler) GetRobot(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":  "Robot detail - TODO: implement",
		"robot_id": id,
	})
}

// UpdateRobot 更新机器人
func (h *Handler) UpdateRobot(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":  "Robot updated - TODO: implement",
		"robot_id": id,
	})
}

// DeleteRobot 删除机器人
func (h *Handler) DeleteRobot(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":  "Robot deleted - TODO: implement",
		"robot_id": id,
	})
}

// TestRobot 测试机器人
func (h *Handler) TestRobot(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":  "Robot test sent - TODO: implement",
		"robot_id": id,
		"status":   "success",
	})
}

// SendNotification 发送通知
func (h *Handler) SendNotification(c *gin.Context) {
	Success(c, gin.H{
		"message":         "Notification sent - TODO: implement",
		"notification_id": 1,
	})
}

// GetNotificationLogs 获取通知日志
func (h *Handler) GetNotificationLogs(c *gin.Context) {
	Success(c, gin.H{
		"message": "Notification logs - TODO: implement",
		"logs":    []gin.H{},
		"total":   0,
	})
}

// GetMessageTemplates 获取消息模板
func (h *Handler) GetMessageTemplates(c *gin.Context) {
	Success(c, gin.H{
		"message":   "Message templates - TODO: implement",
		"templates": []gin.H{},
		"total":     0,
	})
}

// CreateMessageTemplate 创建消息模板
func (h *Handler) CreateMessageTemplate(c *gin.Context) {
	Success(c, gin.H{
		"message":     "Template created - TODO: implement",
		"template_id": 1,
	})
}

// ===== 投资组合模块 =====

// GetPortfolios 获取投资组合列表
func (h *Handler) GetPortfolios(c *gin.Context) {
	Success(c, gin.H{
		"message":    "Portfolios list - TODO: implement",
		"portfolios": []gin.H{},
		"total":      0,
	})
}

// CreatePortfolio 创建投资组合
func (h *Handler) CreatePortfolio(c *gin.Context) {
	Success(c, gin.H{
		"message":      "Portfolio created - TODO: implement",
		"portfolio_id": 1,
	})
}

// GetPortfolio 获取单个投资组合
func (h *Handler) GetPortfolio(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Portfolio detail - TODO: implement",
		"portfolio_id": id,
	})
}

// UpdatePortfolio 更新投资组合
func (h *Handler) UpdatePortfolio(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Portfolio updated - TODO: implement",
		"portfolio_id": id,
	})
}

// DeletePortfolio 删除投资组合
func (h *Handler) DeletePortfolio(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Portfolio deleted - TODO: implement",
		"portfolio_id": id,
	})
}

// GetPortfolioPerformance 获取投资组合表现
func (h *Handler) GetPortfolioPerformance(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Portfolio performance - TODO: implement",
		"portfolio_id": id,
		"performance":  gin.H{},
	})
}

// AddStockToPortfolio 添加股票到投资组合
func (h *Handler) AddStockToPortfolio(c *gin.Context) {
	id := c.Param("id")
	Success(c, gin.H{
		"message":      "Stock added to portfolio - TODO: implement",
		"portfolio_id": id,
	})
}

// RemoveStockFromPortfolio 从投资组合移除股票
func (h *Handler) RemoveStockFromPortfolio(c *gin.Context) {
	id := c.Param("id")
	stockID := c.Param("stock_id")
	Success(c, gin.H{
		"message":      "Stock removed from portfolio - TODO: implement",
		"portfolio_id": id,
		"stock_id":     stockID,
	})
}

// ===== 报表分析模块 =====

// GetStrategyPerformanceReport 获取策略表现报告
func (h *Handler) GetStrategyPerformanceReport(c *gin.Context) {
	Success(c, gin.H{
		"message": "Strategy performance report - TODO: implement",
		"report":  gin.H{},
	})
}

// GetMarketAnalysisReport 获取市场分析报告
func (h *Handler) GetMarketAnalysisReport(c *gin.Context) {
	Success(c, gin.H{
		"message": "Market analysis report - TODO: implement",
		"report":  gin.H{},
	})
}

// GetSelectionSuccessReport 获取选股成功率报告
func (h *Handler) GetSelectionSuccessReport(c *gin.Context) {
	Success(c, gin.H{
		"message": "Selection success report - TODO: implement",
		"report":  gin.H{},
	})
}

// GetRiskAnalysisReport 获取风险分析报告
func (h *Handler) GetRiskAnalysisReport(c *gin.Context) {
	Success(c, gin.H{
		"message": "Risk analysis report - TODO: implement",
		"report":  gin.H{},
	})
}

// GenerateCustomReport 生成自定义报告
func (h *Handler) GenerateCustomReport(c *gin.Context) {
	Success(c, gin.H{
		"message":   "Custom report generated - TODO: implement",
		"report_id": 1,
	})
}

// GetDashboardData 获取仪表板数据
func (h *Handler) GetDashboardData(c *gin.Context) {
	Success(c, gin.H{
		"message":   "Dashboard data - TODO: implement",
		"dashboard": gin.H{},
	})
}

// ===== 系统监控模块 =====

// GetSystemStatus 获取系统状态
func (h *Handler) GetSystemStatus(c *gin.Context) {
	Success(c, gin.H{
		"message": "System status - TODO: implement",
		"status":  "healthy",
		"uptime":  "24h",
	})
}

// GetAPIStats 获取API统计
func (h *Handler) GetAPIStats(c *gin.Context) {
	Success(c, gin.H{
		"message": "API stats - TODO: implement",
		"stats":   gin.H{},
	})
}

// GetSystemLogs 获取系统日志
func (h *Handler) GetSystemLogs(c *gin.Context) {
	Success(c, gin.H{
		"message": "System logs - TODO: implement",
		"logs":    []gin.H{},
		"total":   0,
	})
}

// GetPerformanceMetrics 获取性能指标
func (h *Handler) GetPerformanceMetrics(c *gin.Context) {
	Success(c, gin.H{
		"message": "Performance metrics - TODO: implement",
		"metrics": gin.H{},
	})
}

// GetErrorLogs 获取错误日志
func (h *Handler) GetErrorLogs(c *gin.Context) {
	Success(c, gin.H{
		"message": "Error logs - TODO: implement",
		"errors":  []gin.H{},
		"total":   0,
	})
}

// GetAlerts 获取系统告警
func (h *Handler) GetAlerts(c *gin.Context) {
	Success(c, gin.H{
		"message": "System alerts - TODO: implement",
		"alerts":  []gin.H{},
		"total":   0,
	})
}

// ===== 系统设置模块 =====

// GetSettings 获取系统设置
func (h *Handler) GetSettings(c *gin.Context) {
	Success(c, gin.H{
		"message":  "System settings - TODO: implement",
		"settings": gin.H{},
	})
}

// UpdateSettings 更新系统设置
func (h *Handler) UpdateSettings(c *gin.Context) {
	Success(c, gin.H{
		"message": "Settings updated - TODO: implement",
	})
}

// BackupSettings 备份设置
func (h *Handler) BackupSettings(c *gin.Context) {
	Success(c, gin.H{
		"message":   "Settings backup created - TODO: implement",
		"backup_id": "backup_123",
	})
}

// RestoreSettings 恢复设置
func (h *Handler) RestoreSettings(c *gin.Context) {
	Success(c, gin.H{
		"message": "Settings restored - TODO: implement",
	})
}

// ===== 角色管理模块 =====

// GetRoles 获取角色列表
func (h *Handler) GetRoles(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	keyword := c.Query("keyword")

	pageInt := 1
	pageSizeInt := 10

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
		pageSizeInt = ps
	}

	result, err := h.userService.GetRoles(pageInt, pageSizeInt, keyword)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取角色列表失败", err.Error())
		return
	}

	Success(c, gin.H{
		"roles":     result.Roles,
		"total":     result.Total,
		"page":      pageInt,
		"page_size": pageSizeInt,
	})
}

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	role, err := h.userService.CreateRole(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建角色失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "角色创建成功",
		"role":    role,
	})
}

// GetRole 获取单个角色
func (h *Handler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "角色ID格式错误")
		return
	}

	role, err := h.userService.GetRoleByID(uint(id))
	if err != nil {
		Error(c, http.StatusNotFound, "角色不存在", err.Error())
		return
	}

	Success(c, gin.H{
		"role": role,
	})
}

// UpdateRole 更新角色
func (h *Handler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "角色ID格式错误")
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	role, err := h.userService.UpdateRole(uint(id), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新角色失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "角色更新成功",
		"role":    role,
	})
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "角色ID格式错误")
		return
	}

	err = h.userService.DeleteRole(uint(id))
	if err != nil {
		Error(c, http.StatusInternalServerError, "删除角色失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "角色删除成功",
	})
}

// GetRolePermissions 获取角色权限
func (h *Handler) GetRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "角色ID格式错误")
		return
	}

	// 从数据库获取角色权限
	permissions, err := h.roleRepo.GetRolePermissions(uint(id))
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取角色权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"role_id":     uint(id),
		"permissions": permissions,
	})
}

// UpdateRolePermissions 更新角色权限
func (h *Handler) UpdateRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "角色ID格式错误")
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 调用repository层保存权限
	if err := h.roleRepo.AssignPermissions(uint(id), req.PermissionIDs); err != nil {
		Error(c, http.StatusInternalServerError, "更新角色权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message":        "角色权限更新成功",
		"role_id":        uint(id),
		"permission_ids": req.PermissionIDs,
	})
}

// ===== 权限管理模块 =====

// GetPermissions 获取权限列表
func (h *Handler) GetPermissions(c *gin.Context) {
	var permissions []model.Permission

	// 从数据库查询权限列表
	if err := h.db.Order("sort_order ASC, id ASC").Find(&permissions).Error; err != nil {
		Error(c, http.StatusInternalServerError, "获取权限列表失败", err.Error())
		return
	}

	Success(c, gin.H{
		"permissions": permissions,
		"total":       len(permissions),
	})
}

// GetUserMenuPermissions 获取当前用户的菜单权限
func (h *Handler) GetUserMenuPermissions(c *gin.Context) {
	// 尝试从多个来源获取用户ID
	var userID uint
	var userIDStr string

	// 1. 从JWT token中获取（生产环境）
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	} else if id, exists := c.Get("user_id"); exists {
		userID = id.(uint)
	} else {
		// 2. 从Header中获取（测试环境）
		userIDStr = c.GetHeader("X-User-ID")
		if userIDStr == "" {
			// 3. 从查询参数中获取（调试用）
			userIDStr = c.Query("user_id")
		}

		if userIDStr != "" {
			if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				userID = uint(id)
			}
		}
	}

	// 如果仍然无法获取用户ID，返回管理员权限（用于测试）
	if userID == 0 {
		h.logger.Warn("无法获取用户ID，返回管理员权限用于测试")

		// 获取所有菜单权限
		var allPermissions []model.Permission
		if err := h.db.Where("status = ? AND resource_type IN (?)", 1, []string{"menu", "module"}).
			Order("sort_order ASC").Find(&allPermissions).Error; err != nil {
			Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
			return
		}

		Success(c, gin.H{
			"permissions":    allPermissions,
			"user_role":      "管理员",
			"user_role_code": "admin",
			"message":        "测试模式：返回所有权限",
		})
		return
	}

	// 获取用户信息（带角色）
	userRepo := repository.NewUserRepository(h.db)
	user, err := userRepo.GetWithRole(userID)
	if err != nil {
		h.logger.Errorf("获取用户信息失败: %v", err)
		// 如果用户不存在，也返回管理员权限用于测试
		var allPermissions []model.Permission
		if err := h.db.Where("status = ? AND resource_type IN (?)", 1, []string{"menu", "module"}).
			Order("sort_order ASC").Find(&allPermissions).Error; err != nil {
			Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
			return
		}

		Success(c, gin.H{
			"permissions":    allPermissions,
			"user_role":      "管理员",
			"user_role_code": "admin",
			"message":        "用户不存在，返回管理员权限",
		})
		return
	}

	// 获取用户的菜单权限
	var menuPermissions []model.Permission
	roleName := "普通用户"
	roleCode := "user"

	if user.RoleID != nil && *user.RoleID > 0 {
		// 获取角色信息
		roleRepo := repository.NewRoleRepository(h.db)
		role, err := roleRepo.GetByID(*user.RoleID)
		if err == nil && role.Status == 1 {
			// 查询用户角色拥有的菜单权限
			err = h.db.Model(&model.Permission{}).
				Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
				Where("role_permissions.role_id = ?", *user.RoleID).
				Where("permissions.status = ?", 1).
				Where("permissions.resource_type IN (?)", []string{"menu", "module"}).
				Order("permissions.sort_order ASC").
				Find(&menuPermissions).Error

			if err != nil {
				Error(c, http.StatusInternalServerError, "获取用户权限失败", err.Error())
				return
			}

			roleName = role.RoleName
			roleCode = role.RoleCode
		} else {
			// 角色不存在或被禁用，返回基础权限
			err = h.db.Where("status = ? AND resource_type IN (?) AND permission_code IN (?)",
				1, []string{"menu", "module"}, []string{"dashboard:view"}).
				Order("sort_order ASC").
				Find(&menuPermissions).Error

			if err != nil {
				Error(c, http.StatusInternalServerError, "获取基础权限失败", err.Error())
				return
			}
		}
	} else {
		// 如果用户没有角色，返回基础权限
		err = h.db.Where("status = ? AND resource_type IN (?) AND permission_code IN (?)",
			1, []string{"menu", "module"}, []string{"dashboard:view"}).
			Order("sort_order ASC").
			Find(&menuPermissions).Error

		if err != nil {
			Error(c, http.StatusInternalServerError, "获取基础权限失败", err.Error())
			return
		}
	}

	Success(c, gin.H{
		"permissions":    menuPermissions,
		"user_role":      roleName,
		"user_role_code": roleCode,
	})
}

// CreatePermission 创建权限
func (h *Handler) CreatePermission(c *gin.Context) {
	var permission model.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 检查权限编码是否已存在
	var existingPermission model.Permission
	if err := h.db.Where("permission_code = ?", permission.PermissionCode).First(&existingPermission).Error; err == nil {
		Error(c, http.StatusBadRequest, "权限编码已存在")
		return
	}

	// 设置默认值
	if permission.Status == 0 {
		permission.Status = 1
	}
	if permission.SortOrder == 0 {
		permission.SortOrder = 999
	}

	// 创建权限
	if err := h.db.Create(&permission).Error; err != nil {
		Error(c, http.StatusInternalServerError, "创建权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message":    "权限创建成功",
		"permission": permission,
	})
}

// GetPermission 获取单个权限
func (h *Handler) GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "权限ID格式错误")
		return
	}

	var permission model.Permission
	if err := h.db.Preload("Parent").Preload("Children").First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "权限不存在")
			return
		}
		Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"permission": permission,
	})
}

// UpdatePermission 更新权限
func (h *Handler) UpdatePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "权限ID格式错误")
		return
	}

	var req struct {
		PermissionName string  `json:"permission_name"`
		PermissionCode string  `json:"permission_code"`
		ResourceType   *string `json:"resource_type"`
		ResourcePath   *string `json:"resource_path"`
		ParentID       *uint   `json:"parent_id"`
		Description    *string `json:"description"`
		Status         int8    `json:"status"`
		SortOrder      int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误", err.Error())
		return
	}

	// 获取现有权限
	var permission model.Permission
	if err := h.db.First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "权限不存在")
			return
		}
		Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}

	// 检查权限编码是否被其他权限使用
	if req.PermissionCode != "" && req.PermissionCode != permission.PermissionCode {
		var existingPermission model.Permission
		if err := h.db.Where("permission_code = ? AND id != ?", req.PermissionCode, id).First(&existingPermission).Error; err == nil {
			Error(c, http.StatusBadRequest, "权限编码已被其他权限使用")
			return
		}
		permission.PermissionCode = req.PermissionCode
	}

	// 更新字段
	if req.PermissionName != "" {
		permission.PermissionName = req.PermissionName
	}
	if req.ResourceType != nil {
		permission.ResourceType = req.ResourceType
	}
	if req.ResourcePath != nil {
		permission.ResourcePath = req.ResourcePath
	}
	if req.ParentID != nil {
		permission.ParentID = req.ParentID
	}
	if req.Description != nil {
		permission.Description = req.Description
	}
	if req.Status != 0 {
		permission.Status = req.Status
	}
	if req.SortOrder != 0 {
		permission.SortOrder = req.SortOrder
	}

	if err := h.db.Save(&permission).Error; err != nil {
		Error(c, http.StatusInternalServerError, "更新权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message":    "权限更新成功",
		"permission": permission,
	})
}

// DeletePermission 删除权限
func (h *Handler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "权限ID格式错误")
		return
	}

	// 检查权限是否存在
	var permission model.Permission
	if err := h.db.First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "权限不存在")
			return
		}
		Error(c, http.StatusInternalServerError, "获取权限失败", err.Error())
		return
	}

	// 检查是否有角色使用此权限
	var count int64
	if err := h.db.Model(&model.RolePermission{}).Where("permission_id = ?", id).Count(&count).Error; err != nil {
		Error(c, http.StatusInternalServerError, "检查权限使用情况失败", err.Error())
		return
	}

	if count > 0 {
		Error(c, http.StatusBadRequest, "该权限已被角色使用，无法删除")
		return
	}

	// 删除权限
	if err := h.db.Delete(&permission).Error; err != nil {
		Error(c, http.StatusInternalServerError, "删除权限失败", err.Error())
		return
	}

	Success(c, gin.H{
		"message": "权限删除成功",
	})
}

// ResetPermissions 重置权限配置 - 与前端菜单树结构一致
func (h *Handler) ResetPermissions(c *gin.Context) {
	// 删除现有权限
	if err := h.db.Exec("DELETE FROM permissions").Error; err != nil {
		Error(c, http.StatusInternalServerError, "删除现有权限失败", err.Error())
		return
	}

	// 创建新的权限配置 - 完全匹配前端菜单树结构
	permissions := []model.Permission{
		// 仪表板
		{
			PermissionName: "仪表板",
			PermissionCode: "dashboard:view",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/"),
			Description:    stringPtr("仪表板访问权限"),
			SortOrder:      1,
			Status:         1,
		},

		// 股票数据模块
		{
			PermissionName: "股票数据模块",
			PermissionCode: "stocks:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/stocks"),
			Description:    stringPtr("股票数据模块权限"),
			SortOrder:      2,
			Status:         1,
		},
		{
			PermissionName: "股票列表",
			PermissionCode: "stocks:list",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/stocks"),
			Description:    stringPtr("股票列表权限"),
			SortOrder:      21,
			Status:         1,
		},
		{
			PermissionName: "自选股",
			PermissionCode: "stocks:watchlist",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/stocks/watchlist"),
			Description:    stringPtr("自选股管理权限"),
			SortOrder:      22,
			Status:         1,
		},
		{
			PermissionName: "实时行情",
			PermissionCode: "stocks:realtime",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/stocks/realtime"),
			Description:    stringPtr("实时行情权限"),
			SortOrder:      23,
			Status:         1,
		},

		// 策略管理模块
		{
			PermissionName: "策略管理模块",
			PermissionCode: "strategies:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/strategies"),
			Description:    stringPtr("策略管理模块权限"),
			SortOrder:      3,
			Status:         1,
		},
		{
			PermissionName: "策略列表",
			PermissionCode: "strategies:list",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/strategies"),
			Description:    stringPtr("策略列表权限"),
			SortOrder:      31,
			Status:         1,
		},
		{
			PermissionName: "创建策略",
			PermissionCode: "strategies:create",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/strategies/create"),
			Description:    stringPtr("创建策略权限"),
			SortOrder:      32,
			Status:         1,
		},
		{
			PermissionName: "回测中心",
			PermissionCode: "strategies:backtest",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/strategies/backtest"),
			Description:    stringPtr("回测中心权限"),
			SortOrder:      33,
			Status:         1,
		},

		// 选股结果模块
		{
			PermissionName: "选股结果模块",
			PermissionCode: "selections:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/selections"),
			Description:    stringPtr("选股结果模块权限"),
			SortOrder:      4,
			Status:         1,
		},
		{
			PermissionName: "选股记录",
			PermissionCode: "selections:list",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/selections"),
			Description:    stringPtr("选股记录权限"),
			SortOrder:      41,
			Status:         1,
		},
		{
			PermissionName: "今日选股",
			PermissionCode: "selections:today",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/selections/today"),
			Description:    stringPtr("今日选股权限"),
			SortOrder:      42,
			Status:         1,
		},
		{
			PermissionName: "历史回顾",
			PermissionCode: "selections:history",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/selections/history"),
			Description:    stringPtr("历史回顾权限"),
			SortOrder:      43,
			Status:         1,
		},

		// 投资组合模块
		{
			PermissionName: "投资组合模块",
			PermissionCode: "portfolios:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/portfolios"),
			Description:    stringPtr("投资组合模块权限"),
			SortOrder:      5,
			Status:         1,
		},
		{
			PermissionName: "组合管理",
			PermissionCode: "portfolios:manage",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/portfolios"),
			Description:    stringPtr("组合管理权限"),
			SortOrder:      51,
			Status:         1,
		},
		{
			PermissionName: "业绩分析",
			PermissionCode: "portfolios:performance",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/portfolios/performance"),
			Description:    stringPtr("业绩分析权限"),
			SortOrder:      52,
			Status:         1,
		},

		// 数据采集模块
		{
			PermissionName: "数据采集模块",
			PermissionCode: "collectors:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/collectors"),
			Description:    stringPtr("数据采集模块权限"),
			SortOrder:      6,
			Status:         1,
		},
		{
			PermissionName: "采集器管理",
			PermissionCode: "collectors:manage",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/collectors"),
			Description:    stringPtr("采集器管理权限"),
			SortOrder:      61,
			Status:         1,
		},
		{
			PermissionName: "采集任务",
			PermissionCode: "collectors:tasks",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/collectors/tasks"),
			Description:    stringPtr("采集任务权限"),
			SortOrder:      62,
			Status:         1,
		},
		{
			PermissionName: "数据同步",
			PermissionCode: "collectors:sync",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/collectors/sync"),
			Description:    stringPtr("数据同步权限"),
			SortOrder:      63,
			Status:         1,
		},

		// 通知管理模块
		{
			PermissionName: "通知管理模块",
			PermissionCode: "notifications:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/notifications"),
			Description:    stringPtr("通知管理模块权限"),
			SortOrder:      7,
			Status:         1,
		},
		{
			PermissionName: "机器人配置",
			PermissionCode: "notifications:robots",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/notifications/robots"),
			Description:    stringPtr("机器人配置权限"),
			SortOrder:      71,
			Status:         1,
		},
		{
			PermissionName: "消息模板",
			PermissionCode: "notifications:templates",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/notifications/templates"),
			Description:    stringPtr("消息模板权限"),
			SortOrder:      72,
			Status:         1,
		},
		{
			PermissionName: "发送日志",
			PermissionCode: "notifications:logs",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/notifications/logs"),
			Description:    stringPtr("发送日志权限"),
			SortOrder:      73,
			Status:         1,
		},

		// 报表分析模块
		{
			PermissionName: "报表分析模块",
			PermissionCode: "reports:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/reports"),
			Description:    stringPtr("报表分析模块权限"),
			SortOrder:      8,
			Status:         1,
		},
		{
			PermissionName: "策略报告",
			PermissionCode: "reports:strategy",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/reports/strategy"),
			Description:    stringPtr("策略报告权限"),
			SortOrder:      81,
			Status:         1,
		},
		{
			PermissionName: "市场分析",
			PermissionCode: "reports:market",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/reports/market"),
			Description:    stringPtr("市场分析权限"),
			SortOrder:      82,
			Status:         1,
		},
		{
			PermissionName: "风险分析",
			PermissionCode: "reports:risk",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/reports/risk"),
			Description:    stringPtr("风险分析权限"),
			SortOrder:      83,
			Status:         1,
		},

		// 系统管理模块
		{
			PermissionName: "系统管理模块",
			PermissionCode: "system:module",
			ResourceType:   stringPtr("module"),
			ResourcePath:   stringPtr("/system"),
			Description:    stringPtr("系统管理模块权限"),
			SortOrder:      9,
			Status:         1,
		},
		{
			PermissionName: "用户管理",
			PermissionCode: "system:users",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/system/users"),
			Description:    stringPtr("用户管理权限"),
			SortOrder:      91,
			Status:         1,
		},
		{
			PermissionName: "角色管理",
			PermissionCode: "system:roles",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/system/roles"),
			Description:    stringPtr("角色管理权限"),
			SortOrder:      92,
			Status:         1,
		},
		{
			PermissionName: "系统监控",
			PermissionCode: "system:monitoring",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/system/monitoring"),
			Description:    stringPtr("系统监控权限"),
			SortOrder:      93,
			Status:         1,
		},
		{
			PermissionName: "系统设置",
			PermissionCode: "system:settings",
			ResourceType:   stringPtr("menu"),
			ResourcePath:   stringPtr("/system/settings"),
			Description:    stringPtr("系统设置权限"),
			SortOrder:      94,
			Status:         1,
		},
	}

	// 批量插入权限
	if err := h.db.Create(&permissions).Error; err != nil {
		Error(c, http.StatusInternalServerError, "创建权限失败", err.Error())
		return
	}

	// 统计模块和菜单数量
	moduleCount := 0
	menuCount := 0
	for _, perm := range permissions {
		if perm.ResourceType != nil {
			if *perm.ResourceType == "module" {
				moduleCount++
			} else if *perm.ResourceType == "menu" {
				menuCount++
			}
		}
	}

	Success(c, gin.H{
		"message": "权限配置重置成功 - 已与前端菜单树结构完全匹配",
		"count":   len(permissions),
		"modules": moduleCount,
		"menus":   menuCount,
	})
}

func stringPtr(s string) *string {
	return &s
}
