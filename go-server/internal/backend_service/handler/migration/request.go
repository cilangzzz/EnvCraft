package migration

import (
	"time"

	"tsc/pkg/util/migration/core"
)

// ExecuteRequest 执行迁移请求
type ExecuteRequest struct {
	// Type 迁移类型 (env_variable, config_file, software, registry)
	Type string `json:"type" binding:"required" example:"env_variable"`

	// Name 任务名称
	Name string `json:"name" example:"迁移环境变量"`

	// Source 源配置
	Source SourceConfig `json:"source" binding:"required"`

	// Target 目标配置
	Target TargetConfig `json:"target" binding:"required"`

	// Options 迁移选项
	Options *OptionsConfig `json:"options"`

	// DryRun 是否为预览模式
	DryRun bool `json:"dry_run" example:"false"`
}

// SourceConfig 源配置
type SourceConfig struct {
	// Type 源类型
	Type string `json:"type" example:"env"`

	// Path 源路径
	Path string `json:"path" example:"C:\\config"`

	// Variables 变量列表
	Variables map[string]string `json:"variables" example:"{\"PATH\":\"/usr/bin\"}"`

	// Filter 过滤条件
	Filter *FilterConfig `json:"filter"`

	// Encoding 文件编码
	Encoding string `json:"encoding" example:"utf-8"`

	// Format 配置文件格式
	Format string `json:"format" example:"json"`
}

// TargetConfig 目标配置
type TargetConfig struct {
	// Type 目标类型
	Type string `json:"type" example:"env"`

	// Path 目标路径
	Path string `json:"path" example:"D:\\config"`

	// MergeMode 合并模式 (overwrite, merge, skip)
	MergeMode string `json:"merge_mode" example:"overwrite"`

	// Backup 是否备份
	Backup bool `json:"backup" example:"true"`

	// BackupPath 备份路径
	BackupPath string `json:"backup_path" example:"D:\\backup"`

	// CreateIfNotExists 不存在时是否创建
	CreateIfNotExists bool `json:"create_if_not_exists" example:"true"`

	// Encoding 文件编码
	Encoding string `json:"encoding" example:"utf-8"`

	// Format 配置文件格式
	Format string `json:"format" example:"json"`
}

// FilterConfig 过滤配置
type FilterConfig struct {
	// Include 包含的键
	Include []string `json:"include" example:"[\"PATH\",\"JAVA_HOME\"]"`

	// Exclude 排除的键
	Exclude []string `json:"exclude" example:"[\"TEMP\"]"`

	// Pattern 匹配模式
	Pattern string `json:"pattern" example:"JAVA_*"`
}

// OptionsConfig 迁移选项配置
type OptionsConfig struct {
	// Force 是否强制执行
	Force bool `json:"force" example:"false"`

	// Timeout 超时时间（秒）
	Timeout int `json:"timeout" example:"300"`

	// RetryCount 重试次数
	RetryCount int `json:"retry_count" example:"3"`

	// RetryDelay 重试延迟（毫秒）
	RetryDelay int `json:"retry_delay" example:"1000"`

	// Verbose 是否输出详细日志
	Verbose bool `json:"verbose" example:"true"`

	// StopOnError 遇到错误时是否停止
	StopOnError bool `json:"stop_on_error" example:"true"`

	// SkipValidation 是否跳过验证
	SkipValidation bool `json:"skip_validation" example:"false"`
}

// RollbackRequest 回滚请求
type RollbackRequest struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" binding:"required" example:"task_123"`

	// Type 迁移类型
	Type string `json:"type" binding:"required" example:"env_variable"`
}

// GetTaskRequest 获取任务请求
type GetTaskRequest struct {
	// TaskID 任务ID
	TaskID string `uri:"task_id" binding:"required" example:"task_123"`
}

// ListTasksRequest 列出任务请求
type ListTasksRequest struct {
	// Page 页码
	Page int `form:"page" example:"1"`

	// PageSize 每页数量
	PageSize int `form:"page_size" example:"10"`

	// Status 任务状态
	Status string `form:"status" example:"completed"`

	// Type 迁移类型
	Type string `form:"type" example:"env_variable"`
}

// ToMigrationConfig 将请求转换为迁移配置
func (r *ExecuteRequest) ToMigrationConfig(taskID string) *core.MigrationConfig {
	config := core.NewMigrationConfig()
	config.TaskID = taskID
	config.Name = r.Name
	config.Type = core.MigrationType(r.Type)

	// 源配置
	config.Source.Type = r.Source.Type
	config.Source.Path = r.Source.Path
	config.Source.Variables = r.Source.Variables
	config.Source.Encoding = r.Source.Encoding
	config.Source.Format = r.Source.Format

	if r.Source.Filter != nil {
		config.Source.Filter.Include = r.Source.Filter.Include
		config.Source.Filter.Exclude = r.Source.Filter.Exclude
		config.Source.Filter.Pattern = r.Source.Filter.Pattern
	}

	// 目标配置
	config.Target.Type = r.Target.Type
	config.Target.Path = r.Target.Path
	config.Target.MergeMode = r.Target.MergeMode
	config.Target.Backup = r.Target.Backup
	config.Target.BackupPath = r.Target.BackupPath
	config.Target.CreateIfNotExists = r.Target.CreateIfNotExists
	config.Target.Encoding = r.Target.Encoding
	config.Target.Format = r.Target.Format

	// 选项
	if r.Options != nil {
		config.Options.Force = r.Options.Force
		config.Options.Timeout = r.Options.Timeout
		config.Options.RetryCount = r.Options.RetryCount
		config.Options.RetryDelay = r.Options.RetryDelay
		config.Options.Verbose = r.Options.Verbose
		config.Options.StopOnError = r.Options.StopOnError
		config.Options.SkipValidation = r.Options.SkipValidation
	}

	// DryRun
	config.Options.DryRun = r.DryRun

	return config
}

// ToRollbackConfig 将回滚请求转换为迁移配置
func (r *RollbackRequest) ToRollbackConfig() *core.MigrationConfig {
	config := core.NewMigrationConfig()
	config.TaskID = r.TaskID
	config.Type = core.MigrationType(r.Type)
	return config
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" example:"task_123"`

	// Status 执行状态
	Status string `json:"status" example:"completed"`

	// Message 结果消息
	Message string `json:"message" example:"迁移成功"`

	// Records 迁移记录数
	RecordsCount int `json:"records_count" example:"10"`

	// Summary 汇总信息
	Summary SummaryResponse `json:"summary"`

	// Duration 执行时长（毫秒）
	Duration int64 `json:"duration" example:"1500"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time" example:"2024-01-01T12:00:00Z"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time" example:"2024-01-01T12:00:01Z"`
}

// SummaryResponse 汇总响应
type SummaryResponse struct {
	// Total 总记录数
	Total int `json:"total" example:"10"`

	// Success 成功数
	Success int `json:"success" example:"8"`

	// Failed 失败数
	Failed int `json:"failed" example:"1"`

	// Skipped 跳过数
	Skipped int `json:"skipped" example:"1"`
}

// DryRunResponse 预览响应
type DryRunResponse struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" example:"task_123"`

	// Changes 预览变更列表
	Changes []ChangeResponse `json:"changes"`

	// Warnings 警告信息
	Warnings []string `json:"warnings" example:"[\"目标文件已存在\"]"`

	// Errors 错误信息
	Errors []string `json:"errors" example:"[]"`

	// Summary 汇总信息
	Summary PreviewSummaryResponse `json:"summary"`
}

// ChangeResponse 变更响应
type ChangeResponse struct {
	// ActionType 操作类型
	ActionType string `json:"action_type" example:"update"`

	// Key 操作键
	Key string `json:"key" example:"PATH"`

	// BeforeValue 变更前值
	BeforeValue string `json:"before_value" example:"/usr/bin"`

	// AfterValue 变更后值
	AfterValue string `json:"after_value" example:"/usr/local/bin"`

	// Impact 影响程度
	Impact string `json:"impact" example:"high"`

	// Description 变更描述
	Description string `json:"description" example:"将更新环境变量 PATH"`
}

// PreviewSummaryResponse 预览汇总响应
type PreviewSummaryResponse struct {
	// Total 总变更数
	Total int `json:"total" example:"10"`

	// Create 创建数
	Create int `json:"create" example:"3"`

	// Update 更新数
	Update int `json:"update" example:"5"`

	// Delete 删除数
	Delete int `json:"delete" example:"0"`

	// HighImpact 高影响变更数
	HighImpact int `json:"high_impact" example:"2"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	// ID 数据库ID
	ID uint64 `json:"id" example:"1"`

	// TaskID 任务ID
	TaskID string `json:"task_id" example:"task_123"`

	// Name 任务名称
	Name string `json:"name" example:"迁移环境变量"`

	// Type 迁移类型
	Type string `json:"type" example:"env_variable"`

	// Status 任务状态
	Status string `json:"status" example:"completed"`

	// SourceConfig 源配置
	SourceConfig string `json:"source_config" example:"{\"type\":\"env\"}"`

	// TargetConfig 目标配置
	TargetConfig string `json:"target_config" example:"{\"type\":\"env\"}"`

	// Result 执行结果
	Result string `json:"result" example:"{\"success\":10}"`

	// StartTime 开始时间
	StartTime *time.Time `json:"start_time" example:"2024-01-01T12:00:00Z"`

	// EndTime 结束时间
	EndTime *time.Time `json:"end_time" example:"2024-01-01T12:00:01Z"`

	// Duration 执行时长（毫秒）
	Duration int64 `json:"duration" example:"1500"`

	// ErrorMsg 错误信息
	ErrorMsg string `json:"error_msg" example:""`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T12:00:01Z"`
}

// StrategyResponse 策略响应
type StrategyResponse struct {
	// Type 策略类型
	Type string `json:"type" example:"env_variable"`

	// Name 策略名称
	Name string `json:"name" example:"环境变量迁移策略"`

	// Description 策略描述
	Description string `json:"description" example:"迁移 Windows 环境变量"`
}

// FromMigrationResult 从迁移结果创建响应
func FromMigrationResult(result *core.MigrationResult) *ExecuteResponse {
	return &ExecuteResponse{
		TaskID:       result.TaskID,
		Status:       result.Status,
		Message:      result.Message,
		RecordsCount: len(result.Records),
		Summary: SummaryResponse{
			Total:   result.Summary.Total,
			Success: result.Summary.Success,
			Failed:  result.Summary.Failed,
			Skipped: result.Summary.Skipped,
		},
		Duration:  result.Duration,
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
	}
}

// FromMigrationPreview 从预览结果创建响应
func FromMigrationPreview(preview *core.MigrationPreview) *DryRunResponse {
	changes := make([]ChangeResponse, 0, len(preview.Changes))
	for _, c := range preview.Changes {
		changes = append(changes, ChangeResponse{
			ActionType:  c.ActionType,
			Key:         c.Key,
			BeforeValue: c.BeforeValue,
			AfterValue:  c.AfterValue,
			Impact:      c.Impact,
			Description: c.Description,
		})
	}

	return &DryRunResponse{
		TaskID:   preview.TaskID,
		Changes:  changes,
		Warnings: preview.Warnings,
		Errors:   preview.Errors,
		Summary: PreviewSummaryResponse{
			Total:      preview.Summary.Total,
			Create:     preview.Summary.Create,
			Update:     preview.Summary.Update,
			Delete:     preview.Summary.Delete,
			HighImpact: preview.Summary.HighImpact,
		},
	}
}
