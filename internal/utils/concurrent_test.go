package utils

import (
	"context"
	"errors"
	"stock/internal/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTask 模拟任务
type MockTask struct {
	ID          string
	Description string
	Duration    time.Duration
	ShouldFail  bool
	FailError   error
}

func (t *MockTask) Execute(ctx context.Context) error {
	if t.Duration > 0 {
		select {
		case <-time.After(t.Duration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if t.ShouldFail {
		if t.FailError != nil {
			return t.FailError
		}
		return errors.New("模拟任务失败")
	}

	return nil
}

func (t *MockTask) GetID() string {
	return t.ID
}

func (t *MockTask) GetDescription() string {
	return t.Description
}

func TestConcurrentExecutor_Execute(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(2, logger, 5*time.Second)
	defer executor.Close()

	// 测试成功任务
	task := &MockTask{
		ID:          "test-1",
		Description: "测试任务1",
		Duration:    100 * time.Millisecond,
		ShouldFail:  false,
	}

	ctx := context.Background()
	result := executor.Execute(ctx, task)

	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.Equal(t, "test-1", result.TaskID)
	assert.True(t, result.Duration >= 100*time.Millisecond)
}

func TestConcurrentExecutor_ExecuteWithFailure(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(2, logger, 5*time.Second)
	defer executor.Close()

	// 测试失败任务
	task := &MockTask{
		ID:          "test-fail",
		Description: "失败任务",
		Duration:    50 * time.Millisecond,
		ShouldFail:  true,
		FailError:   errors.New("自定义错误"),
	}

	ctx := context.Background()
	result := executor.Execute(ctx, task)

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Equal(t, "自定义错误", result.Error.Error())
}

func TestConcurrentExecutor_ExecuteBatch(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(3, logger, 5*time.Second)
	defer executor.Close()

	// 创建批量任务
	tasks := []Task{
		&MockTask{ID: "batch-1", Description: "批量任务1", Duration: 100 * time.Millisecond},
		&MockTask{ID: "batch-2", Description: "批量任务2", Duration: 150 * time.Millisecond},
		&MockTask{ID: "batch-3", Description: "批量任务3", Duration: 200 * time.Millisecond, ShouldFail: true},
		&MockTask{ID: "batch-4", Description: "批量任务4", Duration: 50 * time.Millisecond},
	}

	ctx := context.Background()
	results, stats := executor.ExecuteBatch(ctx, tasks)

	require.Len(t, results, 4)
	assert.Equal(t, 4, stats.TotalTasks)
	assert.Equal(t, 3, stats.SuccessTasks)
	assert.Equal(t, 1, stats.FailedTasks)
	assert.True(t, stats.TotalDuration > 0)
	assert.True(t, stats.AverageDuration > 0)

	// 验证每个任务的结果
	taskResults := make(map[string]*TaskResult)
	for _, result := range results {
		taskResults[result.TaskID] = result
	}

	assert.True(t, taskResults["batch-1"].Success)
	assert.True(t, taskResults["batch-2"].Success)
	assert.False(t, taskResults["batch-3"].Success)
	assert.True(t, taskResults["batch-4"].Success)
}

func TestConcurrentExecutor_ExecuteWithRetry(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(1, logger, 5*time.Second)
	defer executor.Close()

	// 测试重试机制
	task := &MockTask{
		ID:          "retry-task",
		Description: "重试任务",
		Duration:    50 * time.Millisecond,
		ShouldFail:  true,
	}

	ctx := context.Background()
	result := executor.ExecuteWithRetry(ctx, task, 2, 100*time.Millisecond)

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
}

func TestConcurrentExecutor_Timeout(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(1, logger, 100*time.Millisecond)
	defer executor.Close()

	// 创建一个会超时的任务
	task := &MockTask{
		ID:          "timeout-task",
		Description: "超时任务",
		Duration:    500 * time.Millisecond, // 超过100ms的超时时间
	}

	ctx := context.Background()
	result := executor.Execute(ctx, task)

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "context deadline exceeded")
}

func TestWorkerPool(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	pool := NewWorkerPool(2, 10, logger, 5*time.Second)
	defer pool.Close()

	// 提交任务
	tasks := []*MockTask{
		{ID: "pool-1", Description: "池任务1", Duration: 100 * time.Millisecond},
		{ID: "pool-2", Description: "池任务2", Duration: 150 * time.Millisecond},
		{ID: "pool-3", Description: "池任务3", Duration: 50 * time.Millisecond},
	}

	// 提交所有任务
	for _, task := range tasks {
		err := pool.Submit(task)
		assert.NoError(t, err)
	}

	// 收集结果
	results := make([]*TaskResult, 0, len(tasks))
	for i := 0; i < len(tasks); i++ {
		result := pool.GetResult()
		require.NotNil(t, result)
		results = append(results, result)
	}

	assert.Len(t, results, 3)

	// 验证所有任务都成功
	for _, result := range results {
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
	}
}

func TestSimpleTask(t *testing.T) {
	executed := false
	task := &SimpleTask{
		ID:          "simple-1",
		Description: "简单任务",
		Func: func(ctx context.Context) error {
			executed = true
			return nil
		},
	}

	ctx := context.Background()
	err := task.Execute(ctx)

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Equal(t, "simple-1", task.GetID())
	assert.Equal(t, "简单任务", task.GetDescription())
}

func TestConcurrentExecutor_ConcurrencyLimit(t *testing.T) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	maxConcurrency := 2
	executor := NewConcurrentExecutor(maxConcurrency, logger, 5*time.Second)
	defer executor.Close()

	// 创建多个长时间运行的任务
	tasks := []Task{
		&MockTask{ID: "concurrent-1", Duration: 300 * time.Millisecond},
		&MockTask{ID: "concurrent-2", Duration: 300 * time.Millisecond},
		&MockTask{ID: "concurrent-3", Duration: 300 * time.Millisecond},
		&MockTask{ID: "concurrent-4", Duration: 300 * time.Millisecond},
	}

	start := time.Now()
	ctx := context.Background()
	results, stats := executor.ExecuteBatch(ctx, tasks)
	totalTime := time.Since(start)

	// 验证结果
	assert.Len(t, results, 4)
	assert.Equal(t, 4, stats.SuccessTasks)

	// 由于并发限制为2，4个300ms的任务应该至少需要600ms
	// (第一批2个并行执行300ms，第二批2个并行执行300ms)
	assert.True(t, totalTime >= 600*time.Millisecond,
		"总时间 %v 应该至少为 600ms，说明并发限制生效", totalTime)
}

// BenchmarkConcurrentExecutor 性能测试
func BenchmarkConcurrentExecutor(b *testing.B) {
	logger := NewLogger(config.LogConfig{Level: "info"})
	executor := NewConcurrentExecutor(4, logger, 5*time.Second)
	defer executor.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := &MockTask{
			ID:          "bench-task",
			Description: "性能测试任务",
			Duration:    1 * time.Millisecond,
		}

		ctx := context.Background()
		result := executor.Execute(ctx, task)

		if !result.Success {
			b.Fatalf("任务执行失败: %v", result.Error)
		}
	}
}
