package utils

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	logger "stock/internal/logger"
)

// ConcurrentExecutor 并发执行器，支持限制最大并发量
type ConcurrentExecutor struct {
	maxConcurrency int
	semaphore      chan struct{}
	wg             sync.WaitGroup
	logger         *logger.Logger
	timeout        time.Duration
}

// Task 任务接口
type Task interface {
	Execute(ctx context.Context) error
	GetID() string
	GetDescription() string
}

// SimpleTask 简单任务实现
type SimpleTask struct {
	ID          string
	Description string
	Func        func(ctx context.Context) error
}

func (t *SimpleTask) Execute(ctx context.Context) error {
	return t.Func(ctx)
}

func (t *SimpleTask) GetID() string {
	return t.ID
}

func (t *SimpleTask) GetDescription() string {
	return t.Description
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID    string
	Success   bool
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// ExecutionStats 执行统计信息
type ExecutionStats struct {
	TotalTasks      int
	SuccessTasks    int
	FailedTasks     int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MaxDuration     time.Duration
	MinDuration     time.Duration
	StartTime       time.Time
	EndTime         time.Time
}

// NewConcurrentExecutor 创建并发执行器
func NewConcurrentExecutor(maxConcurrency int, timeout time.Duration) *ConcurrentExecutor {
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU()
	}

	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}

	return &ConcurrentExecutor{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
		logger:         logger.GetGlobalLogger(),
		timeout:        timeout,
	}
}

// Execute 执行单个任务
func (ce *ConcurrentExecutor) Execute(ctx context.Context, task Task) *TaskResult {
	result := &TaskResult{
		TaskID:    task.GetID(),
		StartTime: time.Now(),
	}

	// 获取信号量
	select {
	case ce.semaphore <- struct{}{}:
		defer func() { <-ce.semaphore }()
	case <-ctx.Done():
		result.Error = ctx.Err()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// 创建带超时的上下文
	taskCtx, cancel := context.WithTimeout(ctx, ce.timeout)
	defer cancel()

	ce.logger.Infof("开始执行任务: %s - %s", task.GetID(), task.GetDescription())

	// 执行任务
	err := task.Execute(taskCtx)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Error = err
	result.Success = err == nil

	if err != nil {
		ce.logger.Errorf("任务执行失败: %s - %v", task.GetID(), err)
	} else {
		ce.logger.Infof("任务执行成功: %s (耗时: %v)", task.GetID(), result.Duration)
	}

	return result
}

// ExecuteBatch 批量执行任务
func (ce *ConcurrentExecutor) ExecuteBatch(ctx context.Context, tasks []Task) ([]*TaskResult, *ExecutionStats) {
	if len(tasks) == 0 {
		return nil, &ExecutionStats{}
	}

	stats := &ExecutionStats{
		TotalTasks:  len(tasks),
		StartTime:   time.Now(),
		MinDuration: time.Hour, // 初始化为一个大值
	}

	results := make([]*TaskResult, len(tasks))
	resultChan := make(chan struct {
		index  int
		result *TaskResult
	}, len(tasks))

	ce.logger.Infof("开始批量执行 %d 个任务，最大并发数: %d", len(tasks), ce.maxConcurrency)

	// 启动所有任务
	for i, task := range tasks {
		ce.wg.Add(1)
		go func(index int, t Task) {
			defer ce.wg.Done()
			result := ce.Execute(ctx, t)
			resultChan <- struct {
				index  int
				result *TaskResult
			}{index, result}
		}(i, task)
	}

	// 等待所有任务完成
	go func() {
		ce.wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for item := range resultChan {
		results[item.index] = item.result

		// 更新统计信息
		if item.result.Success {
			stats.SuccessTasks++
		} else {
			stats.FailedTasks++
		}

		stats.TotalDuration += item.result.Duration

		if item.result.Duration > stats.MaxDuration {
			stats.MaxDuration = item.result.Duration
		}

		if item.result.Duration < stats.MinDuration {
			stats.MinDuration = item.result.Duration
		}
	}

	stats.EndTime = time.Now()
	if stats.TotalTasks > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalTasks)
	}

	ce.logger.Infof("批量任务执行完成: 总数=%d, 成功=%d, 失败=%d, 总耗时=%v, 平均耗时=%v",
		stats.TotalTasks, stats.SuccessTasks, stats.FailedTasks,
		stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	return results, stats
}

// ExecuteWithRetry 带重试的任务执行
func (ce *ConcurrentExecutor) ExecuteWithRetry(ctx context.Context, task Task, maxRetries int, retryDelay time.Duration) *TaskResult {
	var lastResult *TaskResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			ce.logger.Warnf("任务 %s 第 %d 次重试", task.GetID(), attempt)

			// 重试延迟
			select {
			case <-time.After(retryDelay):
			case <-ctx.Done():
				lastResult.Error = ctx.Err()
				return lastResult
			}
		}

		lastResult = ce.Execute(ctx, task)
		if lastResult.Success {
			if attempt > 0 {
				ce.logger.Infof("任务 %s 重试成功 (第 %d 次尝试)", task.GetID(), attempt+1)
			}
			return lastResult
		}
	}

	ce.logger.Errorf("任务 %s 重试 %d 次后仍然失败", task.GetID(), maxRetries)
	return lastResult
}

// GetStats 获取执行器统计信息
func (ce *ConcurrentExecutor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_concurrency": ce.maxConcurrency,
		"timeout":         ce.timeout.String(),
		"available_slots": len(ce.semaphore),
		"used_slots":      ce.maxConcurrency - len(ce.semaphore),
	}
}

// Close 关闭执行器
func (ce *ConcurrentExecutor) Close() {
	ce.wg.Wait()
	close(ce.semaphore)
}

// WorkerPool 工作池，提供更高级的并发控制
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	resultChan chan *TaskResult
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	logger     *logger.Logger
	timeout    time.Duration
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workers int, queueSize int, log *logger.Logger, timeout time.Duration) *WorkerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	if queueSize <= 0 {
		queueSize = workers * 2
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		workers:    workers,
		taskQueue:  make(chan Task, queueSize),
		resultChan: make(chan *TaskResult, queueSize),
		ctx:        ctx,
		cancel:     cancel,
		logger:     log,
		timeout:    timeout,
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	log.Infof("工作池已启动: %d 个工作协程, 队列大小: %d", workers, queueSize)
	return wp
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debugf("工作协程 %d 已启动", id)

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				wp.logger.Debugf("工作协程 %d 退出: 任务队列已关闭", id)
				return
			}

			result := &TaskResult{
				TaskID:    task.GetID(),
				StartTime: time.Now(),
			}

			// 创建带超时的上下文
			taskCtx, cancel := context.WithTimeout(wp.ctx, wp.timeout)

			wp.logger.Debugf("工作协程 %d 开始执行任务: %s", id, task.GetID())

			// 执行任务
			err := task.Execute(taskCtx)
			cancel()

			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			result.Error = err
			result.Success = err == nil

			if err != nil {
				wp.logger.Errorf("工作协程 %d 任务执行失败: %s - %v", id, task.GetID(), err)
			} else {
				wp.logger.Debugf("工作协程 %d 任务执行成功: %s (耗时: %v)", id, task.GetID(), result.Duration)
			}

			// 发送结果
			select {
			case wp.resultChan <- result:
			case <-wp.ctx.Done():
				return
			}

		case <-wp.ctx.Done():
			wp.logger.Debugf("工作协程 %d 退出: 上下文已取消", id)
			return
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("工作池已关闭")
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// GetResult 获取任务结果
func (wp *WorkerPool) GetResult() *TaskResult {
	select {
	case result := <-wp.resultChan:
		return result
	case <-wp.ctx.Done():
		return nil
	}
}

// Close 关闭工作池
func (wp *WorkerPool) Close() {
	wp.logger.Info("正在关闭工作池...")

	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	close(wp.resultChan)

	wp.logger.Info("工作池已关闭")
}

// GetStats 获取工作池统计信息
func (wp *WorkerPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"workers":         wp.workers,
		"queue_size":      cap(wp.taskQueue),
		"pending_tasks":   len(wp.taskQueue),
		"pending_results": len(wp.resultChan),
		"timeout":         wp.timeout.String(),
	}
}
