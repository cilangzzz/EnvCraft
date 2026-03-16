package migration

import (
	"time"
	"tsc/cmd/backend_service/model/common"
)

// MigrationTask 迁移任务模型
type MigrationTask struct {
	common.BaseEntity
	TaskID       string     `json:"task_id" gorm:"uniqueIndex;size:64;comment:任务ID"`
	Name         string     `json:"name" gorm:"size:255;comment:任务名称"`
	Type         string     `json:"type" gorm:"size:64;comment:迁移类型"`
	Status       string     `json:"status" gorm:"size:32;default:pending;comment:任务状态"`
	SourceConfig string     `json:"source_config" gorm:"type:text;comment:源配置(JSON)"`
	TargetConfig string     `json:"target_config" gorm:"type:text;comment:目标配置(JSON)"`
	Options      string     `json:"options" gorm:"type:text;comment:迁移选项(JSON)"`
	Result       string     `json:"result" gorm:"type:text;comment:执行结果(JSON)"`
	StartTime    *time.Time `json:"start_time" gorm:"comment:开始时间"`
	EndTime      *time.Time `json:"end_time" gorm:"comment:结束时间"`
	Duration     int64      `json:"duration" gorm:"comment:执行时长(毫秒)"`
	ErrorMsg     string     `json:"error_msg" gorm:"type:text;comment:错误信息"`
}

// TableName 指定表名
func (MigrationTask) TableName() string {
	return "migration_task"
}

// BeforeCreate GORM钩子 - 创建前
func (t *MigrationTask) BeforeCreate() error {
	if t.Status == "" {
		t.Status = "pending"
	}
	return nil
}

// IsRunning 检查任务是否正在运行
func (t *MigrationTask) IsRunning() bool {
	return t.Status == "running"
}

// IsCompleted 检查任务是否已完成
func (t *MigrationTask) IsCompleted() bool {
	return t.Status == "completed"
}

// IsFailed 检查任务是否失败
func (t *MigrationTask) IsFailed() bool {
	return t.Status == "failed"
}

// CanRollback 检查任务是否可以回滚
func (t *MigrationTask) CanRollback() bool {
	return t.Status == "completed" || t.Status == "failed"
}

// SetRunning 设置任务为运行状态
func (t *MigrationTask) SetRunning() {
	now := time.Now()
	t.Status = "running"
	t.StartTime = &now
	t.ErrorMsg = ""
}

// SetCompleted 设置任务为完成状态
func (t *MigrationTask) SetCompleted(result string) {
	now := time.Now()
	t.Status = "completed"
	t.EndTime = &now
	t.Result = result
	if t.StartTime != nil {
		t.Duration = now.Sub(*t.StartTime).Milliseconds()
	}
}

// SetFailed 设置任务为失败状态
func (t *MigrationTask) SetFailed(errMsg string) {
	now := time.Now()
	t.Status = "failed"
	t.EndTime = &now
	t.ErrorMsg = errMsg
	if t.StartTime != nil {
		t.Duration = now.Sub(*t.StartTime).Milliseconds()
	}
}

// SetRollback 设置任务为回滚状态
func (t *MigrationTask) SetRollback() {
	t.Status = "rollback"
}

// NewMigrationTask 创建迁移任务实例
func NewMigrationTask(taskID, name, migrationType string) *MigrationTask {
	return &MigrationTask{
		TaskID: taskID,
		Name:   name,
		Type:   migrationType,
		Status: "pending",
	}
}
