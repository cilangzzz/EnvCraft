package core

import (
	"time"
)

// ExportPackage 导出包结构 - 导出文件的标准格式
type ExportPackage struct {
	// Metadata 元数据
	Metadata ExportMetadata `json:"metadata"`

	// Content 配置内容
	Content ExportContent `json:"content"`
}

// ExportMetadata 导出元数据
type ExportMetadata struct {
	// Version 导出版本
	Version string `json:"version" example:"1.0"`

	// ExportID 导出唯一标识
	ExportID string `json:"export_id" example:"export_123"`

	// ExportTime 导出时间
	ExportTime time.Time `json:"export_time" example:"2024-01-01T12:00:00Z"`

	// SourceType 源类型 (config_file, env_variable, registry, software)
	SourceType string `json:"source_type" example:"config_file"`

	// OriginalFormat 原始格式
	OriginalFormat string `json:"original_format" example:"json"`

	// OriginalPath 原始路径
	OriginalPath string `json:"original_path" example:"C:\\Users\\...\\.idea\\config.xml"`

	// OriginalEncoding 原始编码
	OriginalEncoding string `json:"original_encoding" example:"utf-8"`

	// AppInfo 应用信息 (可选，如 IDEA, VS Code 等)
	AppInfo *AppInfo `json:"app_info,omitempty"`

	// Checksum 内容校验和
	Checksum string `json:"checksum" example:"sha256:abc123..."`

	// Tags 标签
	Tags []string `json:"tags" example:"[\"ide\",\"java\"]"`

	// Description 描述
	Description string `json:"description" example:"IDEA 配置文件导出"`
}

// AppInfo 应用信息
type AppInfo struct {
	// Name 应用名称
	Name string `json:"name" example:"IntelliJ IDEA"`

	// Version 应用版本
	Version string `json:"version" example:"2024.1"`

	// Category 应用类别
	Category string `json:"category" example:"IDE"`
}

// ExportContent 导出内容
type ExportContent struct {
	// Data 配置数据 (已转换为通用格式)
	Data map[string]interface{} `json:"data"`

	// RawContent 原始内容 (Base64 编码，可选)
	RawContent string `json:"raw_content,omitempty"`

	// FormatSpecificData 格式特定数据 (如 XML 属性、注释等)
	FormatSpecificData map[string]interface{} `json:"format_specific_data,omitempty"`
}

// NewExportPackage 创建导出包实例
func NewExportPackage() *ExportPackage {
	return &ExportPackage{
		Metadata: ExportMetadata{
			Version:    "1.0",
			ExportTime: time.Now(),
			Tags:       make([]string, 0),
		},
		Content: ExportContent{
			Data:               make(map[string]interface{}),
			FormatSpecificData: make(map[string]interface{}),
		},
	}
}

// ExportResult 导出结果
type ExportResult struct {
	// TaskID 任务ID
	TaskID string `json:"task_id"`

	// ExportID 导出ID
	ExportID string `json:"export_id"`

	// Status 执行状态
	Status string `json:"status"`

	// Message 结果消息
	Message string `json:"message"`

	// ExportPath 导出文件路径
	ExportPath string `json:"export_path"`

	// Package 导出包
	Package *ExportPackage `json:"package,omitempty"`

	// Records 导出记录
	Records []MigrationRecord `json:"records"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time"`

	// Duration 执行时长（毫秒）
	Duration int64 `json:"duration"`
}

// NewExportResult 创建导出结果实例
func NewExportResult(taskID string) *ExportResult {
	return &ExportResult{
		TaskID:    taskID,
		Records:   make([]MigrationRecord, 0),
		StartTime: time.Now(),
	}
}

// ImportResult 导入结果
type ImportResult struct {
	// TaskID 任务ID
	TaskID string `json:"task_id"`

	// Status 执行状态
	Status string `json:"status"`

	// Message 结果消息
	Message string `json:"message"`

	// SourcePackage 源导入包信息
	SourcePackage *ExportPackage `json:"source_package,omitempty"`

	// Records 导入记录
	Records []MigrationRecord `json:"records"`

	// Summary 汇总信息
	Summary MigrationSummary `json:"summary"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time"`

	// Duration 执行时长（毫秒）
	Duration int64 `json:"duration"`
}

// NewImportResult 创建导入结果实例
func NewImportResult(taskID string) *ImportResult {
	return &ImportResult{
		TaskID:    taskID,
		Records:   make([]MigrationRecord, 0),
		Summary:   MigrationSummary{},
		StartTime: time.Now(),
	}
}
