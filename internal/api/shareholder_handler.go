package api

import (
	"net/http"
	"strconv"
	"time"

	"stock/internal/model"
	"stock/internal/service"
	"stock/internal/utils"

	"github.com/gin-gonic/gin"
)

// ShareholderHandler 股东户数API处理器
type ShareholderHandler struct {
	service *service.ShareholderService
}

// NewShareholderHandler 创建股东户数处理器
func NewShareholderHandler(service *service.ShareholderService) *ShareholderHandler {
	return &ShareholderHandler{
		service: service,
	}
}

// GetShareholderCounts 获取股东户数数据
// @Summary 获取股东户数数据
// @Description 根据股票代码获取股东户数历史数据
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param ts_code path string true "股票代码，如：000001.SZ"
// @Success 200 {object} Response{data=[]model.ShareholderCount}
// @Router /api/v1/shareholder/{ts_code} [get]
func (h *ShareholderHandler) GetShareholderCounts(c *gin.Context) {
	tsCode := c.Param("ts_code")
	if tsCode == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode = utils.ConvertToTsCode(tsCode)

	counts, err := h.service.GetShareholderCounts(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取股东户数数据失败")
		return
	}

	Success(c, counts)
}

// GetLatestShareholderCount 获取最新股东户数数据
// @Summary 获取最新股东户数数据
// @Description 根据股票代码获取最新的股东户数数据
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param ts_code path string true "股票代码，如：000001.SZ"
// @Success 200 {object} Response{data=model.ShareholderCount}
// @Router /api/v1/shareholder/{ts_code}/latest [get]
func (h *ShareholderHandler) GetLatestShareholderCount(c *gin.Context) {
	tsCode := c.Param("ts_code")
	if tsCode == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode = utils.ConvertToTsCode(tsCode)

	count, err := h.service.GetLatestShareholderCount(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取最新股东户数数据失败")
		return
	}

	Success(c, count)
}

// GetShareholderCountsByDateRange 根据日期范围获取股东户数数据
// @Summary 根据日期范围获取股东户数数据
// @Description 根据股票代码和日期范围获取股东户数数据
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param ts_code path string true "股票代码，如：000001.SZ"
// @Param start_date query string true "开始日期，格式：2006-01-02"
// @Param end_date query string true "结束日期，格式：2006-01-02"
// @Success 200 {object} Response{data=[]model.ShareholderCount}
// @Router /api/v1/shareholder/{ts_code}/range [get]
func (h *ShareholderHandler) GetShareholderCountsByDateRange(c *gin.Context) {
	tsCode := c.Param("ts_code")
	if tsCode == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		Error(c, http.StatusBadRequest, "开始日期和结束日期不能为空")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "开始日期格式错误")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "结束日期格式错误")
		return
	}

	// 转换股票代码格式
	tsCode = utils.ConvertToTsCode(tsCode)

	counts, err := h.service.GetShareholderCountsByDateRange(c.Request.Context(), tsCode, startDate, endDate)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取股东户数数据失败")
		return
	}

	Success(c, counts)
}

// SyncShareholderCounts 同步股东户数数据
// @Summary 同步股东户数数据
// @Description 从数据源同步指定股票的股东户数数据
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param ts_code path string true "股票代码，如：000001.SZ"
// @Success 200 {object} Response{data=string}
// @Router /api/v1/shareholder/{ts_code}/sync [post]
func (h *ShareholderHandler) SyncShareholderCounts(c *gin.Context) {
	tsCode := c.Param("ts_code")
	if tsCode == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode = utils.ConvertToTsCode(tsCode)

	err := h.service.SyncShareholderCounts(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "同步股东户数数据失败")
		return
	}

	Success(c, gin.H{"message": "同步成功"})
}

// SyncAllStocksShareholderCounts 同步所有股票的股东户数数据
// @Summary 同步所有股票的股东户数数据
// @Description 从数据源同步所有股票的股东户数数据（耗时较长）
// @Tags 股东户数
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=string}
// @Router /api/v1/shareholder/sync-all [post]
func (h *ShareholderHandler) SyncAllStocksShareholderCounts(c *gin.Context) {
	err := h.service.SyncAllStocksShareholderCounts(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "同步所有股票股东户数数据失败")
		return
	}

	Success(c, gin.H{"message": "同步成功"})
}

// GetStatistics 获取股东户数统计信息
// @Summary 获取股东户数统计信息
// @Description 获取股东户数数据的统计信息
// @Tags 股东户数
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=map[string]interface{}}
// @Router /api/v1/shareholder/statistics [get]
func (h *ShareholderHandler) GetStatistics(c *gin.Context) {
	stats, err := h.service.GetStatistics(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取统计信息失败")
		return
	}

	Success(c, stats)
}

// GetTopByHolderNum 获取股东户数排行榜
// @Summary 获取股东户数排行榜
// @Description 获取股东户数排行榜（最多或最少）
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制，默认10"
// @Param order query string false "排序方式：asc(升序)或desc(降序)，默认desc"
// @Success 200 {object} Response{data=[]model.ShareholderCount}
// @Router /api/v1/shareholder/top/holder-num [get]
func (h *ShareholderHandler) GetTopByHolderNum(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	order := c.DefaultQuery("order", "desc")
	ascending := order == "asc"

	counts, err := h.service.GetTopByHolderNum(c.Request.Context(), limit, ascending)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取股东户数排行榜失败")
		return
	}

	Success(c, counts)
}

// GetTopByAvgMarketCap 获取户均市值排行榜
// @Summary 获取户均市值排行榜
// @Description 获取户均市值排行榜（最高或最低）
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制，默认10"
// @Param order query string false "排序方式：asc(升序)或desc(降序)，默认desc"
// @Success 200 {object} Response{data=[]model.ShareholderCount}
// @Router /api/v1/shareholder/top/avg-market-cap [get]
func (h *ShareholderHandler) GetTopByAvgMarketCap(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	order := c.DefaultQuery("order", "desc")
	ascending := order == "asc"

	counts, err := h.service.GetTopByAvgMarketCap(c.Request.Context(), limit, ascending)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取户均市值排行榜失败")
		return
	}

	Success(c, counts)
}

// GetRecentChanges 获取股东户数变化较大的股票
// @Summary 获取股东户数变化较大的股票
// @Description 获取最近股东户数变化较大的股票
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制，默认10"
// @Param days query int false "时间范围（天），默认30"
// @Success 200 {object} Response{data=[]model.ShareholderCount}
// @Router /api/v1/shareholder/recent-changes [get]
func (h *ShareholderHandler) GetRecentChanges(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	counts, err := h.service.GetRecentChanges(c.Request.Context(), limit, days)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取股东户数变化数据失败")
		return
	}

	Success(c, counts)
}

// GetShareholderCountsWithPagination 分页获取股东户数数据
// @Summary 分页获取股东户数数据
// @Description 分页获取指定日期范围内的股东户数数据
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：2006-01-02"
// @Param end_date query string true "结束日期，格式：2006-01-02"
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认20"
// @Success 200 {object} Response{data=map[string]interface{}}
// @Router /api/v1/shareholder/list [get]
func (h *ShareholderHandler) GetShareholderCountsWithPagination(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		Error(c, http.StatusBadRequest, "开始日期和结束日期不能为空")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "开始日期格式错误")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "结束日期格式错误")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 20
	}

	counts, total, err := h.service.GetShareholderCountsWithPagination(c.Request.Context(), startDate, endDate, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取股东户数数据失败")
		return
	}

	result := gin.H{
		"data":       counts,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	Success(c, result)
}

// CreateShareholderCount 创建股东户数记录
// @Summary 创建股东户数记录
// @Description 创建新的股东户数记录
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param shareholder_count body model.ShareholderCount true "股东户数数据"
// @Success 200 {object} Response{data=string}
// @Router /api/v1/shareholder [post]
func (h *ShareholderHandler) CreateShareholderCount(c *gin.Context) {
	var count model.ShareholderCount
	if err := c.ShouldBindJSON(&count); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	err := h.service.CreateShareholderCount(c.Request.Context(), &count)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建股东户数记录失败")
		return
	}

	Success(c, gin.H{"message": "创建成功"})
}

// UpdateShareholderCount 更新股东户数记录
// @Summary 更新股东户数记录
// @Description 更新股东户数记录
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param id path int true "记录ID"
// @Param shareholder_count body model.ShareholderCount true "股东户数数据"
// @Success 200 {object} Response{data=string}
// @Router /api/v1/shareholder/{id} [put]
func (h *ShareholderHandler) UpdateShareholderCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	var count model.ShareholderCount
	if err := c.ShouldBindJSON(&count); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	count.ID = uint(id)
	err = h.service.UpdateShareholderCount(c.Request.Context(), &count)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新股东户数记录失败")
		return
	}

	Success(c, gin.H{"message": "更新成功"})
}

// DeleteShareholderCount 删除股东户数记录
// @Summary 删除股东户数记录
// @Description 删除股东户数记录
// @Tags 股东户数
// @Accept json
// @Produce json
// @Param id path int true "记录ID"
// @Success 200 {object} Response{data=string}
// @Router /api/v1/shareholder/{id} [delete]
func (h *ShareholderHandler) DeleteShareholderCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	err = h.service.DeleteShareholderCount(c.Request.Context(), uint(id))
	if err != nil {
		Error(c, http.StatusInternalServerError, "删除股东户数记录失败")
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}
