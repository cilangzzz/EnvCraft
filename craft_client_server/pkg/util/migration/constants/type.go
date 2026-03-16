package constants

import "tsc/pkg/util/migration/core"

// 迁移类型常量
const (
	MigrationTypeEnvVariable core.MigrationType = "env_variable" // 环境变量迁移
	MigrationTypeConfigFile  core.MigrationType = "config_file"  // 配置文件迁移
	MigrationTypeSoftware    core.MigrationType = "software"     // 软件配置迁移
	MigrationTypeRegistry    core.MigrationType = "registry"     // 注册表迁移
)

// 任务状态常量
const (
	TaskStatusPending   string = "pending"   // 待执行
	TaskStatusRunning   string = "running"   // 执行中
	TaskStatusCompleted string = "completed" // 已完成
	TaskStatusFailed    string = "failed"    // 失败
	TaskStatusRollback  string = "rollback"  // 已回滚
)

// 记录状态常量
const (
	RecordStatusSuccess    string = "success"     // 成功
	RecordStatusFailed     string = "failed"      // 失败
	RecordStatusSkipped    string = "skipped"     // 跳过
	RecordStatusRolledBack string = "rolled_back" // 已回滚
)

// 操作类型常量
const (
	ActionTypeCreate string = "create" // 创建
	ActionTypeUpdate string = "update" // 更新
	ActionTypeDelete string = "delete" // 删除
	ActionTypeCopy   string = "copy"   // 复制
	ActionTypeMerge  string = "merge"  // 合并
	ActionTypeExport string = "export" // 导出
	ActionTypeImport string = "import" // 导入
)
