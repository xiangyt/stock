package model

import (
	"time"

	"gorm.io/gorm"
)

// Strategy 选股策略模型
type Strategy struct {
	ID          uint           `json:"id" gorm:"primaryKey"`          // 主键ID
	Name        string         `json:"name" gorm:"size:100;not null"` // 策略名称
	Description string         `json:"description" gorm:"type:text"`  // 策略描述
	Type        string         `json:"type" gorm:"size:50;not null"`  // 策略类型：technical, fundamental, quantitative
	Parameters  string         `json:"parameters" gorm:"type:json"`   // 策略参数JSON
	IsActive    bool           `json:"is_active" gorm:"default:true"` // 是否启用
	CreatedBy   string         `json:"created_by" gorm:"size:50"`     // 创建者
	CreatedAt   time.Time      `json:"created_at"`                    // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                    // 更新时间
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`       // 软删除时间
}

// StrategyResult 策略执行结果模型
type StrategyResult struct {
	ID         uint      `json:"id" gorm:"primaryKey"`                                                                           // 主键ID
	StrategyID uint      `json:"strategy_id" gorm:"not null;index"`                                                              // 策略ID
	Strategy   Strategy  `json:"strategy" gorm:"foreignKey:StrategyID"`                                                          // 关联策略
	TsCode     string    `json:"ts_code" gorm:"size:20;not null;index"`                                                          // 股票代码
	Stock      Stock     `json:"stock" gorm:"foreignKey:TsCode;references:TsCode;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // 关联股票
	Score      float64   `json:"score"`                                                                                          // 评分
	Rank       int       `json:"rank"`                                                                                           // 排名
	Signal     string    `json:"signal" gorm:"size:20"`                                                                          // 信号：buy, sell, hold
	Reason     string    `json:"reason" gorm:"type:text"`                                                                        // 选股原因
	RunDate    time.Time `json:"run_date" gorm:"not null;index"`                                                                 // 运行日期
	CreatedAt  time.Time `json:"created_at"`                                                                                     // 创建时间
}

// Portfolio 投资组合模型
type Portfolio struct {
	ID          uint             `json:"id" gorm:"primaryKey"`                 // 主键ID
	Name        string           `json:"name" gorm:"size:100;not null"`        // 组合名称
	Description string           `json:"description" gorm:"type:text"`         // 组合描述
	TotalValue  float64          `json:"total_value" gorm:"default:0"`         // 总价值
	Cash        float64          `json:"cash" gorm:"default:0"`                // 现金
	IsActive    bool             `json:"is_active" gorm:"default:true"`        // 是否活跃
	CreatedBy   string           `json:"created_by" gorm:"size:50"`            // 创建者
	Stocks      []PortfolioStock `json:"stocks" gorm:"foreignKey:PortfolioID"` // 持仓股票
	CreatedAt   time.Time        `json:"created_at"`                           // 创建时间
	UpdatedAt   time.Time        `json:"updated_at"`                           // 更新时间
	DeletedAt   gorm.DeletedAt   `json:"deleted_at" gorm:"index"`              // 软删除时间
}

// PortfolioStock 投资组合持仓模型
type PortfolioStock struct {
	ID           uint      `json:"id" gorm:"primaryKey"`                                                                           // 主键ID
	PortfolioID  uint      `json:"portfolio_id" gorm:"not null;index"`                                                             // 组合ID
	Portfolio    Portfolio `json:"portfolio" gorm:"foreignKey:PortfolioID"`                                                        // 关联组合
	TsCode       string    `json:"ts_code" gorm:"size:20;not null;index"`                                                          // 股票代码
	Stock        Stock     `json:"stock" gorm:"foreignKey:TsCode;references:TsCode;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // 关联股票
	Shares       int64     `json:"shares" gorm:"not null"`                                                                         // 持股数量
	AvgPrice     float64   `json:"avg_price" gorm:"not null"`                                                                      // 平均成本价
	CurrentPrice float64   `json:"current_price"`                                                                                  // 当前价格
	MarketValue  float64   `json:"market_value"`                                                                                   // 市值
	PnL          float64   `json:"pnl"`                                                                                            // 盈亏
	PnLPercent   float64   `json:"pnl_percent"`                                                                                    // 盈亏百分比
	Weight       float64   `json:"weight"`                                                                                         // 权重
	CreatedAt    time.Time `json:"created_at"`                                                                                     // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`                                                                                     // 更新时间
}
