package service

import (
	"sync"
	"time"

	"stock/internal/indicator"
	"stock/internal/logger"
	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/utils"

	"gorm.io/gorm"
)

// IndicatorService 业绩报表服务
type IndicatorService struct {
	indicatorRepo *repository.TechnicalIndicatorRepository
	stockRepo     *repository.Stock
	dailyDataRepo *repository.DailyData
	weeklyRepo    *repository.WeeklyData
	monthlyRepo   *repository.MonthlyData
	yearlyRepo    *repository.YearlyData
}

var (
	indicatorServiceInstance *IndicatorService
	indicatorServiceOnce     sync.Once
)

// GetIndicatorService 获取技术指标服务单例
func GetIndicatorService(db *gorm.DB) *IndicatorService {
	indicatorServiceOnce.Do(func() {
		indicatorServiceInstance = &IndicatorService{
			indicatorRepo: repository.NewTechnicalIndicatorRepository(db),
			stockRepo:     repository.NewStock(db),
			dailyDataRepo: repository.NewDailyData(db),
			weeklyRepo:    repository.NewWeeklyData(db),
			monthlyRepo:   repository.NewMonthlyData(db),
			yearlyRepo:    repository.NewYearlyData(db),
		}
	})
	return indicatorServiceInstance
}

// CalculateKDJ 计算kdj指标
func (s *IndicatorService) CalculateKDJ(stock model.Stock) error {
	logger.Infof("开始计算股票 %s 的KDJ指标", stock.TsCode)
	if err := s.CalculateKDJByPeriod(stock, model.TechnicalIndicatorPeriodYearly); err != nil {
		logger.Errorf("股票 %s 的年KDJ指标计算失败, err:%s", stock.TsCode, err.Error())
		return err
	}
	if err := s.CalculateKDJByPeriod(stock, model.TechnicalIndicatorPeriodMonthly); err != nil {
		logger.Errorf("股票 %s 的月KDJ指标计算失败, err:%s", stock.TsCode, err.Error())
		return err
	}
	if err := s.CalculateKDJByPeriod(stock, model.TechnicalIndicatorPeriodWeekly); err != nil {
		logger.Errorf("股票 %s 的周KDJ指标计算失败, err:%s", stock.TsCode, err.Error())
		return err
	}
	if err := s.CalculateKDJByPeriod(stock, model.TechnicalIndicatorPeriodDaily); err != nil {
		logger.Errorf("股票 %s 的日KDJ指标计算失败, err:%s", stock.TsCode, err.Error())
		return err
	}
	logger.Infof("成功计算股票 %s 的KDJ指标", stock.TsCode)
	return nil
}

// CalculateKDJByPeriod 计算不同周期的kdj指标
func (s *IndicatorService) CalculateKDJByPeriod(stock model.Stock, period model.TechnicalIndicatorPeriod) error {
	logger.Infof("开始计算股票 %s 的KDJ指标(%s)", stock.TsCode, period)
	inds, err := s.indicatorRepo.GetBySymbol(stock.Symbol, period)
	if err != nil {
		return err
	}
	if len(inds) < 9 {
		inds = nil
	}
	stocks, err := s.getIndStockList(stock, inds, period)
	if err != nil {
		return err
	}

	var exist bool
	if len(inds) == 0 {
		inds = indicator.KDJ(stocks)
		if len(inds) > 60 {
			inds = inds[len(inds)-60:]
		}
		exist = true
	} else {
		end := inds[len(inds)-2].TradeDate // 从倒数第二条开始更新
		for i, stock := range stocks {
			if stock.GetTradeDate() == end {
				inds = indicator.KDJ(stocks, i+1)
				exist = true
				break
			}
		}
	}
	for _, ind := range inds {
		ind.Period = period
	}
	if exist {
		err = s.indicatorRepo.UpsertKdj(inds)
	}

	logger.Infof("成功计算股票 %s 的KDJ指标(%s)，共 %d 条记录", stock.TsCode, period, len(inds))
	return err
}

func (s *IndicatorService) getIndStockList(stock model.Stock, inds []*model.TechnicalIndicator,
	period model.TechnicalIndicatorPeriod) ([]indicator.KDJStock, error) {
	var stocks []indicator.KDJStock
	var start = time.Time{}
	if len(inds) > 0 {
		start, _ = utils.ParseTradeDate(inds[0].TradeDate)
	}
	indMap := map[int]*model.TechnicalIndicator{}
	for _, ind := range inds {
		indMap[ind.TradeDate] = ind
	}
	switch period {
	case model.TechnicalIndicatorPeriodDaily:
		list, err := s.dailyDataRepo.GetDailyData(stock.TsCode, start, time.Time{}, 0)
		if err != nil {
			return nil, err
		}

		for i := len(list) - 1; i >= 0; i-- {
			v := list[i]
			if ind, ok := indMap[v.TradeDate]; ok {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock:  v,
					Indicator: *ind,
				})
			} else {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock: v,
				})
			}
		}
	case model.TechnicalIndicatorPeriodWeekly:
		list, err := s.weeklyRepo.GetWeeklyData(stock.TsCode, start, time.Time{}, 0)
		if err != nil {
			return nil, err
		}
		for i := len(list) - 1; i >= 0; i-- {
			v := list[i]
			if ind, ok := indMap[v.TradeDate]; ok {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock:  v,
					Indicator: *ind,
				})
			} else {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock: v,
				})
			}
		}
	case model.TechnicalIndicatorPeriodMonthly:
		list, err := s.monthlyRepo.GetMonthlyData(stock.TsCode, start, time.Time{}, 0)
		if err != nil {
			return nil, err
		}
		for i := len(list) - 1; i >= 0; i-- {
			v := list[i]
			if ind, ok := indMap[v.TradeDate]; ok {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock:  v,
					Indicator: *ind,
				})
			} else {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock: v,
				})
			}
		}
	case model.TechnicalIndicatorPeriodYearly:
		list, err := s.yearlyRepo.GetYearlyDataByTsCode(stock.TsCode, start, time.Time{}, 0)
		if err != nil {
			return nil, err
		}
		for i := len(list) - 1; i >= 0; i-- {
			v := list[i]
			if ind, ok := indMap[v.TradeDate]; ok {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock:  v,
					Indicator: *ind,
				})
			} else {
				stocks = append(stocks, &indicator.KDJBase{
					IndStock: v,
				})
			}
		}
	}

	return stocks, nil
}
