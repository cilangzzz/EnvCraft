package migration

import (
	"fmt"
	"tsc/pkg/util/migration/constants"
	"tsc/pkg/util/migration/core"
)

// MigrationType 迁移类型常量
var MigrationType = struct {
	EnvVariable core.MigrationType
	ConfigFile  core.MigrationType
	Software    core.MigrationType
	Registry    core.MigrationType
}{
	EnvVariable: constants.MigrationTypeEnvVariable,
	ConfigFile:  constants.MigrationTypeConfigFile,
	Software:    constants.MigrationTypeSoftware,
	Registry:    constants.MigrationTypeRegistry,
}

// TaskStatus 任务状态常量
var TaskStatus = struct {
	Pending   string
	Running   string
	Completed string
	Failed    string
	Rollback  string
}{
	Pending:   constants.TaskStatusPending,
	Running:   constants.TaskStatusRunning,
	Completed: constants.TaskStatusCompleted,
	Failed:    constants.TaskStatusFailed,
	Rollback:  constants.TaskStatusRollback,
}

// RecordStatus 记录状态常量
var RecordStatus = struct {
	Success    string
	Failed     string
	Skipped    string
	RolledBack string
}{
	Success:    constants.RecordStatusSuccess,
	Failed:     constants.RecordStatusFailed,
	Skipped:    constants.RecordStatusSkipped,
	RolledBack: constants.RecordStatusRolledBack,
}

// ActionType 操作类型常量
var ActionType = struct {
	Create string
	Update string
	Delete string
	Copy   string
	Merge  string
	Export string
	Import string
}{
	Create: constants.ActionTypeCreate,
	Update: constants.ActionTypeUpdate,
	Delete: constants.ActionTypeDelete,
	Copy:   constants.ActionTypeCopy,
	Merge:  constants.ActionTypeMerge,
	Export: constants.ActionTypeExport,
	Import: constants.ActionTypeImport,
}

// New 创建迁移策略实例
// migrationType: 迁移类型
// 返回对应的迁移策略实例
func New(migrationType core.MigrationType) (core.MigrationStrategy, error) {
	strategy, err := core.GetStrategy(migrationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration strategy: %w", err)
	}
	return strategy, nil
}

// NewConfig 创建迁移配置实例
func NewConfig() *core.MigrationConfig {
	return core.NewMigrationConfig()
}

// NewContext 创建迁移上下文实例
func NewContext(taskID string) *core.MigrationContext {
	return core.NewMigrationContext(taskID)
}

// RegisterStrategy 注册迁移策略
func RegisterStrategy(strategy core.MigrationStrategy) error {
	return core.RegisterStrategy(strategy)
}

// GetStrategy 获取迁移策略
func GetStrategy(migrationType core.MigrationType) (core.MigrationStrategy, error) {
	return core.GetStrategy(migrationType)
}

// ListStrategies 列出所有已注册的策略
func ListStrategies() []core.MigrationStrategy {
	return core.ListStrategies()
}

// Execute 执行迁移任务
func Execute(config *core.MigrationConfig) (*core.MigrationResult, error) {
	strategy, err := core.GetStrategy(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration strategy: %w", err)
	}

	// 设置上下文
	if config.Context == nil {
		config.Context = core.NewMigrationContext(config.TaskID)
	}

	// 验证配置
	if err := strategy.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// 执行迁移
	return strategy.Execute(config.Context.Context, config)
}

// DryRun 预览迁移任务
func DryRun(config *core.MigrationConfig) (*core.MigrationPreview, error) {
	strategy, err := core.GetStrategy(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration strategy: %w", err)
	}

	// 设置上下文
	if config.Context == nil {
		config.Context = core.NewMigrationContext(config.TaskID)
	}

	// 验证配置
	if err := strategy.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// 执行预览
	return strategy.DryRun(config.Context.Context, config)
}

// Rollback 回滚迁移任务
func Rollback(config *core.MigrationConfig) error {
	strategy, err := core.GetStrategy(config.Type)
	if err != nil {
		return fmt.Errorf("failed to get migration strategy: %w", err)
	}

	// 设置上下文
	if config.Context == nil {
		config.Context = core.NewMigrationContext(config.TaskID)
	}

	// 执行回滚
	return strategy.Rollback(config.Context.Context, config)
}
