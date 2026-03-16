package core

import (
	"context"
	"time"
)

// MigrationType 迁移类型
type MigrationType string

// MigrationStrategy 迁移策略接口
type MigrationStrategy interface {
	// Name 返回策略名称
	Name() string

	// Type 返回策略类型
	Type() MigrationType

	// Description 返回策略描述
	Description() string

	// Validate 验证配置是否有效
	Validate(config *MigrationConfig) error

	// Execute 执行迁移
	Execute(ctx context.Context, config *MigrationConfig) (*MigrationResult, error)

	// Rollback 回滚迁移
	Rollback(ctx context.Context, config *MigrationConfig) error

	// DryRun 预览迁移（模拟执行）
	DryRun(ctx context.Context, config *MigrationConfig) (*MigrationPreview, error)

	// Export 导出配置
	Export(ctx context.Context, config *MigrationConfig) (*ExportResult, error)

	// Import 导入配置
	Import(ctx context.Context, config *MigrationConfig) (*ImportResult, error)

	// ValidateExport 验证导出配置
	ValidateExport(config *MigrationConfig) error

	// ValidateImport 验证导入配置
	ValidateImport(config *MigrationConfig) error
}

// MigrationConfig 迁移配置
type MigrationConfig struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" gorm:"size:64;comment:任务ID"`

	// Name 任务名称
	Name string `json:"name" gorm:"size:255;comment:任务名称"`

	// Type 迁移类型
	Type MigrationType `json:"type" gorm:"size:64;comment:迁移类型"`

	// Source 源配置
	Source MigrationSource `json:"source" gorm:"type:json;comment:源配置"`

	// Target 目标配置
	Target MigrationTarget `json:"target" gorm:"type:json;comment:目标配置"`

	// Options 迁移选项
	Options MigrationOptions `json:"options" gorm:"type:json;comment:迁移选项"`

	// Context 迁移上下文
	Context *MigrationContext `json:"context,omitempty" gorm:"-"`
}

// MigrationSource 迁移源配置
type MigrationSource struct {
	// Type 源类型 (local, remote, file, env, registry)
	Type string `json:"type" gorm:"size:32;comment:源类型"`

	// Path 源路径（文件路径或注册表路径）
	Path string `json:"path" gorm:"size:512;comment:源路径"`

	// Variables 环境变量列表
	Variables map[string]string `json:"variables" gorm:"type:json;comment:变量列表"`

	// Filter 过滤条件
	Filter SourceFilter `json:"filter" gorm:"type:json;comment:过滤条件"`

	// Encoding 文件编码
	Encoding string `json:"encoding" gorm:"size:32;comment:文件编码"`

	// Format 配置文件格式 (json, yaml, ini, toml)
	Format string `json:"format" gorm:"size:16;comment:文件格式"`
}

// SourceFilter 源过滤条件
type SourceFilter struct {
	// Include 包含的键
	Include []string `json:"include" gorm:"type:json;comment:包含的键"`

	// Exclude 排除的键
	Exclude []string `json:"exclude" gorm:"type:json;comment:排除的键"`

	// Pattern 匹配模式（正则表达式）
	Pattern string `json:"pattern" gorm:"size:256;comment:匹配模式"`
}

// MigrationTarget 目标配置
type MigrationTarget struct {
	// Type 目标类型 (local, remote, env, registry)
	Type string `json:"type" gorm:"size:32;comment:目标类型"`

	// Path 目标路径
	Path string `json:"path" gorm:"size:512;comment:目标路径"`

	// MergeMode 合并模式 (overwrite, merge, skip)
	MergeMode string `json:"merge_mode" gorm:"size:32;comment:合并模式"`

	// Backup 是否备份
	Backup bool `json:"backup" gorm:"comment:是否备份"`

	// BackupPath 备份路径
	BackupPath string `json:"backup_path" gorm:"size:512;comment:备份路径"`

	// CreateIfNotExists 不存在时是否创建
	CreateIfNotExists bool `json:"create_if_not_exists" gorm:"comment:不存在时是否创建"`

	// Encoding 文件编码
	Encoding string `json:"encoding" gorm:"size:32;comment:文件编码"`

	// Format 配置文件格式
	Format string `json:"format" gorm:"size:16;comment:文件格式"`
}

// MigrationOptions 迁移选项
type MigrationOptions struct {
	// DryRun 是否为预览模式
	DryRun bool `json:"dry_run" gorm:"comment:是否预览模式"`

	// Force 是否强制执行（忽略警告）
	Force bool `json:"force" gorm:"comment:是否强制执行"`

	// Timeout 超时时间（秒）
	Timeout int `json:"timeout" gorm:"comment:超时时间(秒)"`

	// RetryCount 重试次数
	RetryCount int `json:"retry_count" gorm:"comment:重试次数"`

	// RetryDelay 重试延迟（毫秒）
	RetryDelay int `json:"retry_delay" gorm:"comment:重试延迟(毫秒)"`

	// Verbose 是否输出详细日志
	Verbose bool `json:"verbose" gorm:"comment:是否详细日志"`

	// StopOnError 遇到错误时是否停止
	StopOnError bool `json:"stop_on_error" gorm:"comment:错误时是否停止"`

	// SkipValidation 是否跳过验证
	SkipValidation bool `json:"skip_validation" gorm:"comment:是否跳过验证"`

	// OperationMode 操作模式 (migrate, export, import)
	OperationMode string `json:"operation_mode" gorm:"size:32;comment:操作模式"`

	// ExportPath 导出文件路径 (导出模式使用)
	ExportPath string `json:"export_path" gorm:"size:512;comment:导出文件路径"`

	// ImportPath 导入文件路径 (导入模式使用)
	ImportPath string `json:"import_path" gorm:"size:512;comment:导入文件路径"`

	// IncludeRawContent 是否包含原始内容
	IncludeRawContent bool `json:"include_raw_content" gorm:"comment:是否包含原始内容"`

	// PreserveFormat 是否保持原始格式 (导入时)
	PreserveFormat bool `json:"preserve_format" gorm:"comment:是否保持原始格式"`
}

// MigrationResult 迁移结果
type MigrationResult struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" gorm:"size:64;comment:任务ID"`

	// Status 执行状态
	Status string `json:"status" gorm:"size:32;comment:执行状态"`

	// Message 结果消息
	Message string `json:"message" gorm:"type:text;comment:结果消息"`

	// Warnings 警告信息
	Warnings []string `json:"warnings" gorm:"type:json;comment:警告信息"`

	// Records 迁移记录
	Records []MigrationRecord `json:"records" gorm:"type:json;comment:迁移记录"`

	// Summary 汇总信息
	Summary MigrationSummary `json:"summary" gorm:"type:json;comment:汇总信息"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time" gorm:"comment:开始时间"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time" gorm:"comment:结束时间"`

	// Duration 执行时长（毫秒）
	Duration int64 `json:"duration" gorm:"comment:执行时长(毫秒)"`
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	// StepName 步骤名称
	StepName string `json:"step_name" gorm:"size:255;comment:步骤名称"`

	// ActionType 操作类型
	ActionType string `json:"action_type" gorm:"size:64;comment:操作类型"`

	// Key 操作键
	Key string `json:"key" gorm:"size:512;comment:操作键"`

	// BeforeValue 变更前值
	BeforeValue string `json:"before_value" gorm:"type:text;comment:变更前值"`

	// AfterValue 变更后值
	AfterValue string `json:"after_value" gorm:"type:text;comment:变更后值"`

	// Status 记录状态
	Status string `json:"status" gorm:"size:32;comment:记录状态"`

	// Message 记录消息
	Message string `json:"message" gorm:"type:text;comment:记录消息"`

	// Timestamp 时间戳
	Timestamp time.Time `json:"timestamp" gorm:"comment:时间戳"`
}

// MigrationSummary 迁移汇总
type MigrationSummary struct {
	// Total 总记录数
	Total int `json:"total" gorm:"comment:总记录数"`

	// Success 成功数
	Success int `json:"success" gorm:"comment:成功数"`

	// Failed 失败数
	Failed int `json:"failed" gorm:"comment:失败数"`

	// Skipped 跳过数
	Skipped int `json:"skipped" gorm:"comment:跳过数"`

	// RolledBack 回滚数
	RolledBack int `json:"rolled_back" gorm:"comment:回滚数"`
}

// MigrationPreview 迁移预览
type MigrationPreview struct {
	// TaskID 任务ID
	TaskID string `json:"task_id" gorm:"size:64;comment:任务ID"`

	// Changes 预览的变更列表
	Changes []PreviewChange `json:"changes" gorm:"type:json;comment:变更列表"`

	// Warnings 警告信息
	Warnings []string `json:"warnings" gorm:"type:json;comment:警告信息"`

	// Errors 错误信息
	Errors []string `json:"errors" gorm:"type:json;comment:错误信息"`

	// Summary 汇总信息
	Summary PreviewSummary `json:"summary" gorm:"type:json;comment:汇总信息"`
}

// PreviewChange 预览变更
type PreviewChange struct {
	// ActionType 操作类型
	ActionType string `json:"action_type" gorm:"size:64;comment:操作类型"`

	// Key 操作键
	Key string `json:"key" gorm:"size:512;comment:操作键"`

	// BeforeValue 变更前值
	BeforeValue string `json:"before_value" gorm:"type:text;comment:变更前值"`

	// AfterValue 变更后值
	AfterValue string `json:"after_value" gorm:"type:text;comment:变更后值"`

	// Impact 影响程度 (high, medium, low)
	Impact string `json:"impact" gorm:"size:16;comment:影响程度"`

	// Description 变更描述
	Description string `json:"description" gorm:"type:text;comment:变更描述"`
}

// PreviewSummary 预览汇总
type PreviewSummary struct {
	// Total 总变更数
	Total int `json:"total" gorm:"comment:总变更数"`

	// Create 创建数
	Create int `json:"create" gorm:"comment:创建数"`

	// Update 更新数
	Update int `json:"update" gorm:"comment:更新数"`

	// Delete 删除数
	Delete int `json:"delete" gorm:"comment:删除数"`

	// HighImpact 高影响变更数
	HighImpact int `json:"high_impact" gorm:"comment:高影响变更数"`
}

// NewMigrationConfig 创建迁移配置实例
func NewMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		Source: MigrationSource{
			Variables: make(map[string]string),
			Filter: SourceFilter{
				Include: make([]string, 0),
				Exclude: make([]string, 0),
			},
		},
		Target: MigrationTarget{
			MergeMode: "overwrite",
		},
		Options: MigrationOptions{
			Timeout:     300,
			RetryCount:  3,
			RetryDelay:  1000,
			StopOnError: true,
		},
	}
}

// NewMigrationResult 创建迁移结果实例
func NewMigrationResult(taskID string) *MigrationResult {
	return &MigrationResult{
		TaskID:   taskID,
		Records:  make([]MigrationRecord, 0),
		Warnings: make([]string, 0),
		Summary:  MigrationSummary{},
	}
}

// NewMigrationPreview 创建迁移预览实例
func NewMigrationPreview(taskID string) *MigrationPreview {
	return &MigrationPreview{
		TaskID:   taskID,
		Changes:  make([]PreviewChange, 0),
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
		Summary:  PreviewSummary{},
	}
}
