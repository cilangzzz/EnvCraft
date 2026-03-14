package core

import (
	"context"
	"sync"
	"time"
)

// MigrationContext 迁移上下文
type MigrationContext struct {
	// TaskID 任务ID
	TaskID string `json:"task_id"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`

	// Context Go上下文
	Context context.Context `json:"-"`

	// Cancel 取消函数
	Cancel context.CancelFunc `json:"-"`

	// Logger 日志记录器
	Logger MigrationLogger `json:"-"`

	// State 状态存储
	State map[string]interface{} `json:"state"`

	// Records 迁移记录
	Records []MigrationRecord `json:"records"`

	// mu 互斥锁
	mu sync.RWMutex `json:"-"`
}

// MigrationLogger 迁移日志接口
type MigrationLogger interface {
	// Debug 调试日志
	Debug(format string, args ...interface{})

	// Info 信息日志
	Info(format string, args ...interface{})

	// Warn 警告日志
	Warn(format string, args ...interface{})

	// Error 错误日志
	Error(format string, args ...interface{})
}

// NewMigrationContext 创建迁移上下文
func NewMigrationContext(taskID string) *MigrationContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &MigrationContext{
		TaskID:    taskID,
		StartTime: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
		State:     make(map[string]interface{}),
		Records:   make([]MigrationRecord, 0),
	}
}

// WithContext 使用已有上下文创建迁移上下文
func NewMigrationContextWithContext(taskID string, parentCtx context.Context) *MigrationContext {
	ctx, cancel := context.WithCancel(parentCtx)
	return &MigrationContext{
		TaskID:    taskID,
		StartTime: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
		State:     make(map[string]interface{}),
		Records:   make([]MigrationRecord, 0),
	}
}

// SetState 设置状态
func (c *MigrationContext) SetState(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.State[key] = value
}

// GetState 获取状态
func (c *MigrationContext) GetState(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.State[key]
	return value, ok
}

// DeleteState 删除状态
func (c *MigrationContext) DeleteState(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.State, key)
}

// AddRecord 添加迁移记录
func (c *MigrationContext) AddRecord(record MigrationRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Records = append(c.Records, record)
}

// GetRecords 获取所有记录
func (c *MigrationContext) GetRecords() []MigrationRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	records := make([]MigrationRecord, len(c.Records))
	copy(records, c.Records)
	return records
}

// RecordCount 获取记录数量
func (c *MigrationContext) RecordCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Records)
}

// IsCancelled 检查是否已取消
func (c *MigrationContext) IsCancelled() bool {
	select {
	case <-c.Context.Done():
		return true
	default:
		return false
	}
}

// Cancel 取消迁移
func (c *MigrationContext) CancelMigration() {
	if c.Cancel != nil {
		c.Cancel()
	}
}

// ElapsedTime 获取已用时间
func (c *MigrationContext) ElapsedTime() time.Duration {
	return time.Since(c.StartTime)
}

// SetLogger 设置日志记录器
func (c *MigrationContext) SetLogger(logger MigrationLogger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Logger = logger
}

// LogDebug 记录调试日志
func (c *MigrationContext) LogDebug(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Debug(format, args...)
	}
}

// LogInfo 记录信息日志
func (c *MigrationContext) LogInfo(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Info(format, args...)
	}
}

// LogWarn 记录警告日志
func (c *MigrationContext) LogWarn(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Warn(format, args...)
	}
}

// LogError 记录错误日志
func (c *MigrationContext) LogError(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Error(format, args...)
	}
}

// DefaultLogger 默认日志实现（空操作）
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(format string, args ...interface{}) {}
func (l *DefaultLogger) Info(format string, args ...interface{})  {}
func (l *DefaultLogger) Warn(format string, args ...interface{})  {}
func (l *DefaultLogger) Error(format string, args ...interface{}) {}
