package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap 自定义JSON类型，用于存储map[string]interface{}
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待中
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSyncAllStocks   TaskType = "sync_all_stocks"   // 同步全量股票
	TaskTypeSyncSingleStock TaskType = "sync_single_stock" // 刷新单只股票日K数据
)

// Task 异步任务
type Task struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Type        TaskType   `json:"type" gorm:"not null"`
	Status      TaskStatus `json:"status" gorm:"not null;default:pending"`
	Progress    int        `json:"progress" gorm:"default:0"`   // 进度百分比 0-100
	Message     string     `json:"message"`                     // 状态消息
	Parameters  JSONMap    `json:"parameters" gorm:"type:json"` // 任务参数
	Result      JSONMap    `json:"result" gorm:"type:json"`     // 任务结果
	Error       string     `json:"error"`                       // 错误信息
	StartedAt   *time.Time `json:"started_at"`                  // 开始时间
	CompletedAt *time.Time `json:"completed_at"`                // 完成时间
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TaskSummary 任务摘要（用于列表显示）
type TaskSummary struct {
	ID        string     `json:"id"`
	Type      TaskType   `json:"type"`
	Status    TaskStatus `json:"status"`
	Progress  int        `json:"progress"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
}
