package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"stock/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// generateTaskID 生成任务ID
func generateTaskID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// TaskService 任务服务
type TaskService struct {
	db     *gorm.DB
	logger *logrus.Logger

	// 运行中的任务
	runningTasks sync.Map // map[string]*TaskRunner
}

// TaskRunner 任务执行器
type TaskRunner struct {
	Task   *model.Task
	Cancel context.CancelFunc
	Done   chan struct{}
}

// NewTaskService 创建任务服务
func NewTaskService(db *gorm.DB, logger *logrus.Logger) *TaskService {
	return &TaskService{
		db:     db,
		logger: logger,
	}
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(taskType model.TaskType, parameters map[string]interface{}) (*model.Task, error) {
	task := &model.Task{
		ID:        generateTaskID(),
		Type:      taskType,
		Status:    model.TaskStatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 初始化Parameters和Result为空map，避免nil值
	if parameters == nil {
		task.Parameters = make(model.JSONMap)
	} else {
		task.Parameters = model.JSONMap(parameters)
	}
	task.Result = make(model.JSONMap)

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.logger.Infof("Created task %s of type %s", task.ID, task.Type)
	return task, nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(taskID string) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(limit, offset int, status model.TaskStatus) ([]model.TaskSummary, int64, error) {
	var tasks []model.Task
	var total int64

	query := s.db.Model(&model.Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取任务列表
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	// 转换为摘要格式
	summaries := make([]model.TaskSummary, len(tasks))
	for i, task := range tasks {
		summaries[i] = model.TaskSummary{
			ID:        task.ID,
			Type:      task.Type,
			Status:    task.Status,
			Progress:  task.Progress,
			Message:   task.Message,
			CreatedAt: task.CreatedAt,
		}
	}

	return summaries, total, nil
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(taskID string, status model.TaskStatus, progress int, message string) error {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"message":    message,
		"updated_at": time.Now(),
	}

	if status == model.TaskStatusRunning && progress == 0 {
		updates["started_at"] = time.Now()
	}

	if status == model.TaskStatusCompleted || status == model.TaskStatusFailed {
		updates["completed_at"] = time.Now()
	}

	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(taskID string, result map[string]interface{}) error {
	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Update("result", result).Error
}

// UpdateTaskError 更新任务错误
func (s *TaskService) UpdateTaskError(taskID string, errorMsg string) error {
	return s.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":       model.TaskStatusFailed,
		"error":        errorMsg,
		"completed_at": time.Now(),
		"updated_at":   time.Now(),
	}).Error
}

// StartTask 启动任务执行
func (s *TaskService) StartTask(taskID string, executor func(ctx context.Context, task *model.Task, updateProgress func(int, string)) error) {
	// 检查任务是否已在运行
	if _, exists := s.runningTasks.Load(taskID); exists {
		s.logger.Warnf("Task %s is already running", taskID)
		return
	}

	// 获取任务详情
	task, err := s.GetTask(taskID)
	if err != nil {
		s.logger.Errorf("Failed to get task %s: %v", taskID, err)
		return
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	runner := &TaskRunner{
		Task:   task,
		Cancel: cancel,
		Done:   make(chan struct{}),
	}

	// 存储运行中的任务
	s.runningTasks.Store(taskID, runner)

	// 启动goroutine执行任务
	go func() {
		defer func() {
			close(runner.Done)
			s.runningTasks.Delete(taskID)
		}()

		// 更新任务状态为运行中
		if err := s.UpdateTaskStatus(taskID, model.TaskStatusRunning, 0, "任务开始执行"); err != nil {
			s.logger.Errorf("Failed to update task status: %v", err)
			return
		}

		// 进度更新函数
		updateProgress := func(progress int, message string) {
			if err := s.UpdateTaskStatus(taskID, model.TaskStatusRunning, progress, message); err != nil {
				s.logger.Errorf("Failed to update task progress: %v", err)
			}
		}

		// 执行任务
		if err := executor(ctx, task, updateProgress); err != nil {
			s.logger.Errorf("Task %s failed: %v", taskID, err)
			s.UpdateTaskError(taskID, err.Error())
		} else {
			s.logger.Infof("Task %s completed successfully", taskID)
			s.UpdateTaskStatus(taskID, model.TaskStatusCompleted, 100, "任务执行完成")
		}
	}()
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(taskID string) error {
	if runnerInterface, exists := s.runningTasks.Load(taskID); exists {
		runner := runnerInterface.(*TaskRunner)
		runner.Cancel()

		// 等待任务结束
		select {
		case <-runner.Done:
		case <-time.After(5 * time.Second):
			s.logger.Warnf("Task %s cancellation timeout", taskID)
		}

		return s.UpdateTaskStatus(taskID, model.TaskStatusFailed, 0, "任务已取消")
	}

	return fmt.Errorf("task %s is not running", taskID)
}
