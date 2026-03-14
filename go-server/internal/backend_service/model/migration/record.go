package migration

import (
	"time"

	"tsc/internal/backend_service/model/common"
)

// MigrationRecord 迁移记录模型
type MigrationRecord struct {
	common.BaseEntity
	TaskID      string    `json:"task_id" gorm:"index;size:64;comment:任务ID"`
	StepName    string    `json:"step_name" gorm:"size:255;comment:步骤名称"`
	ActionType  string    `json:"action_type" gorm:"size:64;comment:操作类型"`
	Key         string    `json:"key" gorm:"size:512;comment:操作键"`
	BeforeValue string    `json:"before_value" gorm:"type:text;comment:变更前值"`
	AfterValue  string    `json:"after_value" gorm:"type:text;comment:变更后值"`
	Status      string    `json:"status" gorm:"size:32;default:success;comment:记录状态"`
	Message     string    `json:"message" gorm:"type:text;comment:记录消息"`
	Timestamp   time.Time `json:"timestamp" gorm:"autoCreateTime;comment:时间戳"`
}

// TableName 指定表名
func (MigrationRecord) TableName() string {
	return "migration_record"
}

// BeforeCreate GORM钩子 - 创建前
func (r *MigrationRecord) BeforeCreate() error {
	if r.Status == "" {
		r.Status = "success"
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	return nil
}

// IsSuccess 检查记录是否成功
func (r *MigrationRecord) IsSuccess() bool {
	return r.Status == "success"
}

// IsFailed 检查记录是否失败
func (r *MigrationRecord) IsFailed() bool {
	return r.Status == "failed"
}

// IsSkipped 检查记录是否跳过
func (r *MigrationRecord) IsSkipped() bool {
	return r.Status == "skipped"
}

// IsRolledBack 检查记录是否已回滚
func (r *MigrationRecord) IsRolledBack() bool {
	return r.Status == "rolled_back"
}

// SetSuccess 设置记录为成功状态
func (r *MigrationRecord) SetSuccess() {
	r.Status = "success"
}

// SetFailed 设置记录为失败状态
func (r *MigrationRecord) SetFailed(message string) {
	r.Status = "failed"
	r.Message = message
}

// SetSkipped 设置记录为跳过状态
func (r *MigrationRecord) SetSkipped(message string) {
	r.Status = "skipped"
	r.Message = message
}

// SetRolledBack 设置记录为回滚状态
func (r *MigrationRecord) SetRolledBack() {
	r.Status = "rolled_back"
}

// NewMigrationRecord 创建迁移记录实例
func NewMigrationRecord(taskID, stepName, actionType, key string) *MigrationRecord {
	return &MigrationRecord{
		TaskID:     taskID,
		StepName:   stepName,
		ActionType: actionType,
		Key:        key,
		Status:     "success",
		Timestamp:  time.Now(),
	}
}

// NewMigrationRecordWithValues 创建带值的迁移记录实例
func NewMigrationRecordWithValues(taskID, stepName, actionType, key, beforeValue, afterValue string) *MigrationRecord {
	return &MigrationRecord{
		TaskID:      taskID,
		StepName:    stepName,
		ActionType:  actionType,
		Key:         key,
		BeforeValue: beforeValue,
		AfterValue:  afterValue,
		Status:      "success",
		Timestamp:   time.Now(),
	}
}
