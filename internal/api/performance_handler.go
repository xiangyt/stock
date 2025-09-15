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

// PerformanceHandler 业绩报表API处理器
type PerformanceHandler struct {
	service *service.PerformanceService
}

// NewPerformanceHandler 创建业绩报表处理器
func NewPerformanceHandler(service *service.PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{
		service: service,
	}
}

// GetPerformanceReports 获取业绩报表数据
// @Summary 获取业绩报表数据
// @Description 根据股票代码获取业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param code path string true "股票代码"
// @Success 200 {object} Response{data=[]model.PerformanceReport}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{code}/performance [get]
func (h *PerformanceHandler) GetPerformanceReports(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := utils.ConvertToTsCode(code)

	reports, err := h.service.GetPerformanceReports(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取业绩报表数据失败")
		return
	}

	Success(c, reports)
}

// GetLatestPerformanceReport 获取最新业绩报表数据
// @Summary 获取最新业绩报表数据
// @Description 根据股票代码获取最新业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param code path string true "股票代码"
// @Success 200 {object} Response{data=model.PerformanceReport}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{code}/performance/latest [get]
func (h *PerformanceHandler) GetLatestPerformanceReport(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := utils.ConvertToTsCode(code)

	report, err := h.service.GetLatestPerformanceReport(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取最新业绩报表数据失败")
		return
	}

	if report == nil {
		Error(c, http.StatusNotFound, "未找到业绩报表数据")
		return
	}

	Success(c, report)
}

// SyncPerformanceReports 同步业绩报表数据
// @Summary 同步业绩报表数据
// @Description 从数据源同步指定股票的业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param code path string true "股票代码"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{code}/performance/sync [post]
func (h *PerformanceHandler) SyncPerformanceReports(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := utils.ConvertToTsCode(code)

	err := h.service.SyncPerformanceReports(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "同步业绩报表数据失败")
		return
	}

	Success(c, gin.H{"message": "同步成功"})
}

// SyncAllPerformanceReports 同步所有股票的业绩报表数据
// @Summary 同步所有股票的业绩报表数据
// @Description 从数据源同步所有股票的业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/performance/sync-all [post]
func (h *PerformanceHandler) SyncAllPerformanceReports(c *gin.Context) {
	err := h.service.SyncAllStocksPerformanceReports(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "同步所有业绩报表数据失败")
		return
	}

	Success(c, gin.H{"message": "同步成功"})
}

// GetPerformanceReportsByDateRange 根据日期范围获取业绩报表
// @Summary 根据日期范围获取业绩报表
// @Description 根据股票代码和日期范围获取业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param code path string true "股票代码"
// @Param start_date query string true "开始日期 (YYYY-MM-DD)"
// @Param end_date query string true "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} Response{data=[]model.PerformanceReport}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{code}/performance/range [get]
func (h *PerformanceHandler) GetPerformanceReportsByDateRange(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
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
	tsCode := utils.ConvertToTsCode(code)

	reports, err := h.service.GetPerformanceReportsByDateRange(c.Request.Context(), tsCode, startDate, endDate)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取业绩报表数据失败")
		return
	}

	Success(c, reports)
}

// GetTopPerformers 获取业绩表现最好的股票
// @Summary 获取业绩表现最好的股票
// @Description 根据指定指标获取业绩表现最好的股票
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制" default(10)
// @Param order_by query string false "排序字段" default(eps) Enums(eps,roe,roa,gross_margin,dividend_yield,revenue,net_profit)
// @Success 200 {object} Response{data=[]model.PerformanceReport}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/performance/top-performers [get]
func (h *PerformanceHandler) GetTopPerformers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	orderBy := c.DefaultQuery("order_by", "eps")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100 // 限制最大返回数量
	}

	reports, err := h.service.GetTopPerformers(c.Request.Context(), limit, orderBy)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取业绩排行数据失败")
		return
	}

	Success(c, reports)
}

// GetPerformanceStatistics 获取业绩报表统计信息
// @Summary 获取业绩报表统计信息
// @Description 获取业绩报表的统计信息
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=map[string]interface{}}
// @Failure 500 {object} Response
// @Router /api/v1/performance/statistics [get]
func (h *PerformanceHandler) GetPerformanceStatistics(c *gin.Context) {
	stats, err := h.service.GetStatistics(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取统计信息失败")
		return
	}

	Success(c, stats)
}

// CreatePerformanceReport 创建业绩报表记录
// @Summary 创建业绩报表记录
// @Description 手动创建业绩报表记录
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param report body model.PerformanceReport true "业绩报表数据"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/performance [post]
func (h *PerformanceHandler) CreatePerformanceReport(c *gin.Context) {
	var report model.PerformanceReport
	if err := c.ShouldBindJSON(&report); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	err := h.service.CreatePerformanceReport(c.Request.Context(), &report)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建业绩报表记录失败")
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "创建成功",
		Data:    report,
	})
}

// UpdatePerformanceReport 更新业绩报表记录
// @Summary 更新业绩报表记录
// @Description 更新业绩报表记录
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param id path int true "记录ID"
// @Param report body model.PerformanceReport true "业绩报表数据"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/performance/{id} [put]
func (h *PerformanceHandler) UpdatePerformanceReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	var report model.PerformanceReport
	if err := c.ShouldBindJSON(&report); err != nil {
		Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	report.ID = uint(id)
	err = h.service.UpdatePerformanceReport(c.Request.Context(), &report)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新业绩报表记录失败")
		return
	}

	Success(c, report)
}

// DeletePerformanceReports 删除指定股票的业绩报表数据
// @Summary 删除业绩报表数据
// @Description 删除指定股票的所有业绩报表数据
// @Tags 业绩报表
// @Accept json
// @Produce json
// @Param code path string true "股票代码"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{code}/performance [delete]
func (h *PerformanceHandler) DeletePerformanceReports(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, "股票代码不能为空")
		return
	}

	// 转换股票代码格式
	tsCode := utils.ConvertToTsCode(code)

	err := h.service.DeletePerformanceReports(c.Request.Context(), tsCode)
	if err != nil {
		Error(c, http.StatusInternalServerError, "删除业绩报表数据失败")
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}
