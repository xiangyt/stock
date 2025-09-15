package handler

import (
	"net/http"
	"strconv"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"
	"stock/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handlers 处理器集合
type Handlers struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services

	Stock    *StockHandler
	Analysis *AnalysisHandler
	Strategy *StrategyHandler
	Backtest *BacktestHandler
	Web      *WebHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(cfg *config.Config, logger *utils.Logger) *Handlers {
	services, _ := service.NewServices(cfg, logger)

	return &Handlers{
		cfg:      cfg,
		logger:   logger,
		services: services,
		Stock:    NewStockHandler(cfg, logger, services),
		Analysis: NewAnalysisHandler(cfg, logger, services),
		Strategy: NewStrategyHandler(cfg, logger, services),
		Backtest: NewBacktestHandler(cfg, logger, services),
		Web:      NewWebHandler(cfg, logger, services),
	}
}

// Health 健康检查
func (h *Handlers) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"service": "stock",
		"version": h.cfg.App.Version,
	})
}

// StockHandler 股票处理器
type StockHandler struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services
}

// NewStockHandler 创建股票处理器
func NewStockHandler(cfg *config.Config, logger *utils.Logger, services *service.Services) *StockHandler {
	return &StockHandler{
		cfg:      cfg,
		logger:   logger,
		services: services,
	}
}

// GetStocks 获取股票列表
func (h *StockHandler) GetStocks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	market := c.Query("market")

	h.logger.Infof("Getting stocks: page=%d, limit=%d, market=%s", page, limit, market)

	// TODO: 从数据库获取股票列表
	stocks := []gin.H{
		{"ts_code": "000001.SZ", "name": "平安银行", "industry": "银行", "market": "SZ"},
		{"ts_code": "000002.SZ", "name": "万科A", "industry": "房地产开发", "market": "SZ"},
		{"ts_code": "600000.SH", "name": "浦发银行", "industry": "银行", "market": "SH"},
		{"ts_code": "600036.SH", "name": "招商银行", "industry": "银行", "market": "SH"},
	}

	response.SuccessPaged(c, stocks, int64(len(stocks)), page, limit)
}

// GetStock 获取股票详情
func (h *StockHandler) GetStock(c *gin.Context) {
	code := c.Param("code")
	h.logger.Infof("Getting stock detail: %s", code)

	// TODO: 从数据库获取股票详情
	stock := gin.H{
		"ts_code":   code,
		"name":      "平安银行",
		"industry":  "银行",
		"market":    "SZ",
		"list_date": "1991-04-03",
	}

	response.Success(c, stock)
}

// GetStockData 获取股票数据
func (h *StockHandler) GetStockData(c *gin.Context) {
	code := c.Param("code")
	period := c.DefaultQuery("period", "1m")

	h.logger.Infof("Getting stock data: %s, period=%s", code, period)

	// TODO: 从数据库获取股票数据
	data := gin.H{
		"ts_code": code,
		"period":  period,
		"data": []gin.H{
			{"date": "2023-12-01", "open": 12.34, "high": 12.56, "low": 12.12, "close": 12.45, "volume": 1000000},
			{"date": "2023-12-02", "open": 12.45, "high": 12.67, "low": 12.23, "close": 12.56, "volume": 1200000},
		},
	}

	response.Success(c, data)
}

// AnalysisHandler 分析处理器
type AnalysisHandler struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services
}

// NewAnalysisHandler 创建分析处理器
func NewAnalysisHandler(cfg *config.Config, logger *utils.Logger, services *service.Services) *AnalysisHandler {
	return &AnalysisHandler{
		cfg:      cfg,
		logger:   logger,
		services: services,
	}
}

// GetTechnicalAnalysis 获取技术分析
func (h *AnalysisHandler) GetTechnicalAnalysis(c *gin.Context) {
	code := c.Param("code")
	period := c.DefaultQuery("period", "1m")

	h.logger.Infof("Getting technical analysis: %s, period=%s", code, period)

	// TODO: 实现技术分析逻辑
	analysis := gin.H{
		"ts_code": code,
		"indicators": gin.H{
			"ma5":  12.34,
			"ma20": 11.89,
			"rsi":  45.6,
			"macd": 0.12,
		},
		"signals": gin.H{
			"ma_signal":   "bullish",
			"rsi_signal":  "neutral",
			"macd_signal": "golden_cross",
		},
		"score": 75.5,
	}

	response.Success(c, analysis)
}

// GetFundamentalAnalysis 获取基本面分析
func (h *AnalysisHandler) GetFundamentalAnalysis(c *gin.Context) {
	code := c.Param("code")

	h.logger.Infof("Getting fundamental analysis: %s", code)

	// TODO: 实现基本面分析逻辑
	analysis := gin.H{
		"ts_code": code,
		"financial": gin.H{
			"roe":      0.15,
			"roa":      0.08,
			"pe_ratio": 12.5,
			"pb_ratio": 1.2,
		},
		"score": 68.2,
	}

	response.Success(c, analysis)
}

// StrategyHandler 策略处理器
type StrategyHandler struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services
}

// NewStrategyHandler 创建策略处理器
func NewStrategyHandler(cfg *config.Config, logger *utils.Logger, services *service.Services) *StrategyHandler {
	return &StrategyHandler{
		cfg:      cfg,
		logger:   logger,
		services: services,
	}
}

// GetStrategies 获取策略列表
func (h *StrategyHandler) GetStrategies(c *gin.Context) {
	strategies := []gin.H{
		{"name": "technical", "description": "技术面策略", "enabled": true},
		{"name": "fundamental", "description": "基本面策略", "enabled": true},
		{"name": "combined", "description": "综合策略", "enabled": true},
	}

	response.Success(c, strategies)
}

// SelectionRequest 选股请求
type SelectionRequest struct {
	Strategy   string                 `json:"strategy" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Limit      int                    `json:"limit"`
}

// ExecuteSelection 执行选股
func (h *StrategyHandler) ExecuteSelection(c *gin.Context) {
	var req SelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	h.logger.Infof("Executing selection: strategy=%s, limit=%d", req.Strategy, req.Limit)

	// 调用策略引擎执行选股
	results, err := h.services.StrategyEngine.ExecuteStrategy(req.Strategy, req.Limit)
	if err != nil {
		h.logger.Errorf("Strategy execution failed: %v", err)
		response.InternalServerError(c, "Strategy execution failed")
		return
	}

	response.Success(c, gin.H{
		"strategy": req.Strategy,
		"count":    len(results),
		"results":  results,
	})
}

// GetResults 获取选股结果
func (h *StrategyHandler) GetResults(c *gin.Context) {
	strategy := c.Query("strategy")
	date := c.Query("date")

	h.logger.Infof("Getting selection results: strategy=%s, date=%s", strategy, date)

	// TODO: 从数据库获取选股结果
	results := []gin.H{
		{"ts_code": "000001.SZ", "name": "平安银行", "score": 85.6, "reason": "技术指标良好"},
		{"ts_code": "000002.SZ", "name": "万科A", "score": 78.9, "reason": "基本面稳健"},
	}

	response.Success(c, results)
}

// BacktestHandler 回测处理器
type BacktestHandler struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services
}

// NewBacktestHandler 创建回测处理器
func NewBacktestHandler(cfg *config.Config, logger *utils.Logger, services *service.Services) *BacktestHandler {
	return &BacktestHandler{
		cfg:      cfg,
		logger:   logger,
		services: services,
	}
}

// BacktestRequest 回测请求
type BacktestRequest struct {
	Strategy  string  `json:"strategy" binding:"required"`
	StartDate string  `json:"start_date" binding:"required"`
	EndDate   string  `json:"end_date" binding:"required"`
	Capital   float64 `json:"capital"`
}

// RunBacktest 运行回测
func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	var req BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters")
		return
	}

	h.logger.Infof("Running backtest: strategy=%s, period=%s to %s", req.Strategy, req.StartDate, req.EndDate)

	// TODO: 实现回测逻辑
	result := gin.H{
		"strategy": req.Strategy,
		"period": gin.H{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
		},
		"performance": gin.H{
			"total_return":  0.25,
			"annual_return": 0.15,
			"max_drawdown":  0.08,
			"sharpe_ratio":  1.2,
			"win_rate":      0.65,
		},
	}

	response.Success(c, result)
}

// GetResult 获取回测结果
func (h *BacktestHandler) GetResult(c *gin.Context) {
	id := c.Param("id")

	h.logger.Infof("Getting backtest result: %s", id)

	// TODO: 从数据库获取回测结果
	result := gin.H{
		"id":       id,
		"strategy": "technical",
		"performance": gin.H{
			"total_return":  0.25,
			"annual_return": 0.15,
		},
	}

	response.Success(c, result)
}

// WebHandler Web页面处理器
type WebHandler struct {
	cfg      *config.Config
	logger   *utils.Logger
	services *service.Services
}

// NewWebHandler 创建Web处理器
func NewWebHandler(cfg *config.Config, logger *utils.Logger, services *service.Services) *WebHandler {
	return &WebHandler{
		cfg:      cfg,
		logger:   logger,
		services: services,
	}
}

// Index 首页
func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "智能选股系统",
	})
}

// Analysis 分析页面
func (h *WebHandler) Analysis(c *gin.Context) {
	c.HTML(http.StatusOK, "analysis.html", gin.H{
		"title": "技术分析",
	})
}

// Strategy 策略页面
func (h *WebHandler) Strategy(c *gin.Context) {
	c.HTML(http.StatusOK, "strategy.html", gin.H{
		"title": "选股策略",
	})
}
