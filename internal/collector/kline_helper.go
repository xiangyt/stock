package collector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"stock/internal/model"
)

// KLineParser K线数据解析器
type KLineParser struct{}

// NewKLineParser 创建K线解析器
func NewKLineParser() *KLineParser {
	return &KLineParser{}
}

// ParseToDaily 解析为日K线数据
func (p *KLineParser) ParseToDaily(tsCode, kline string) (*model.DailyData, error) {
	fields := strings.Split(kline, ",")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid kline data format: %s", kline)
	}

	tradeDateInt, err := p.parseTradeDate(fields[0])
	if err != nil {
		return nil, err
	}

	return &model.DailyData{
		TsCode:    tsCode,
		TradeDate: tradeDateInt,
		Open:      p.parseFloat(fields[1]),
		High:      p.parseFloat(fields[3]),
		Low:       p.parseFloat(fields[4]),
		Close:     p.parseFloat(fields[2]),
		Volume:    p.parseInt64(fields[5]),
		Amount:    p.parseFloat(fields[6]),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

// ParseToWeekly 解析为周K线数据
func (p *KLineParser) ParseToWeekly(tsCode, kline string) (*model.WeeklyData, error) {
	fields := strings.Split(kline, ",")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid kline data format: %s", kline)
	}

	tradeDateInt, err := p.parseTradeDate(fields[0])
	if err != nil {
		return nil, err
	}

	return &model.WeeklyData{
		TsCode:    tsCode,
		TradeDate: tradeDateInt,
		Open:      p.parseFloat(fields[1]),
		Close:     p.parseFloat(fields[2]),
		High:      p.parseFloat(fields[3]),
		Low:       p.parseFloat(fields[4]),
		Volume:    p.parseInt64(fields[5]),
		Amount:    p.parseFloat(fields[6]),
		CreatedAt: time.Now().Unix(),
	}, nil
}

// ParseToMonthly 解析为月K线数据
func (p *KLineParser) ParseToMonthly(tsCode, kline string) (*model.MonthlyData, error) {
	fields := strings.Split(kline, ",")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid kline data format: %s", kline)
	}

	tradeDateInt, err := p.parseTradeDate(fields[0])
	if err != nil {
		return nil, err
	}

	return &model.MonthlyData{
		TsCode:    tsCode,
		TradeDate: tradeDateInt,
		Open:      p.parseFloat(fields[1]),
		Close:     p.parseFloat(fields[2]),
		High:      p.parseFloat(fields[3]),
		Low:       p.parseFloat(fields[4]),
		Volume:    p.parseInt64(fields[5]),
		Amount:    p.parseFloat(fields[6]),
		CreatedAt: time.Now().Unix(),
	}, nil
}

// ParseToYearly 解析为年K线数据
func (p *KLineParser) ParseToYearly(tsCode, kline string) (*model.YearlyData, error) {
	fields := strings.Split(kline, ",")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid kline data format: %s", kline)
	}

	tradeDateInt, err := p.parseTradeDate(fields[0])
	if err != nil {
		return nil, err
	}

	return &model.YearlyData{
		TsCode:    tsCode,
		TradeDate: tradeDateInt,
		Open:      p.parseFloat(fields[1]),
		Close:     p.parseFloat(fields[2]),
		High:      p.parseFloat(fields[3]),
		Low:       p.parseFloat(fields[4]),
		Volume:    p.parseInt64(fields[5]),
		Amount:    p.parseFloat(fields[6]),
		CreatedAt: time.Now().Unix(),
	}, nil
}

// parseTradeDate 解析交易日期
func (p *KLineParser) parseTradeDate(dateStr string) (int, error) {
	tradeDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}
	return tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day(), nil
}

// parseFloat 安全解析浮点数
func (p *KLineParser) parseFloat(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// parseInt64 安全解析64位整数
func (p *KLineParser) parseInt64(s string) int64 {
	if s == "" || s == "-" {
		return 0
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
