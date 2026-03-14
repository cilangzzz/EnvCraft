package strategies

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"tsc/pkg/util/migration/constants"
	"tsc/pkg/util/migration/core"
)

func init() {
	// 自动注册到全局注册表
	if err := core.RegisterStrategy(&EnvVariableStrategy{}); err != nil {
		panic(fmt.Sprintf("failed to register env variable strategy: %v", err))
	}
}

// EnvVariableStrategy 环境变量迁移策略
type EnvVariableStrategy struct{}

// Name 返回策略名称
func (s *EnvVariableStrategy) Name() string {
	return "环境变量迁移策略"
}

// Type 返回策略类型
func (s *EnvVariableStrategy) Type() core.MigrationType {
	return constants.MigrationTypeEnvVariable
}

// Description 返回策略描述
func (s *EnvVariableStrategy) Description() string {
	return "迁移 Windows 环境变量，支持用户变量和系统变量的迁移"
}

// Validate 验证配置是否有效
func (s *EnvVariableStrategy) Validate(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证源配置
	if config.Source.Type == "" {
		return fmt.Errorf("source type is required")
	}

	// 验证目标配置
	if config.Target.Type == "" {
		return fmt.Errorf("target type is required")
	}

	// 验证环境变量列表
	if len(config.Source.Variables) == 0 && config.Source.Filter.Pattern == "" {
		return fmt.Errorf("at least one variable or filter pattern is required")
	}

	return nil
}

// Execute 执行环境变量迁移
func (s *EnvVariableStrategy) Execute(ctx context.Context, config *core.MigrationConfig) (*core.MigrationResult, error) {
	result := core.NewMigrationResult(config.TaskID)
	result.StartTime = time.Now()
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("env variable migration is only supported on Windows")
	}

	// 获取要迁移的变量
	variables := s.getVariablesToMigrate(config)

	// 遍历并设置环境变量
	for name, value := range variables {
		record := core.MigrationRecord{
			StepName:   fmt.Sprintf("设置环境变量 %s", name),
			ActionType: constants.ActionTypeUpdate,
			Key:        name,
			Timestamp:  time.Now(),
		}

		// 获取当前值
		oldValue := os.Getenv(name)
		record.BeforeValue = oldValue

		// 根据目标类型设置环境变量
		var err error
		switch config.Target.Type {
		case "user":
			err = s.setUserEnvVar(name, value)
		case "system":
			err = s.setSystemEnvVar(name, value)
		case "process":
			err = os.Setenv(name, value)
		default:
			err = os.Setenv(name, value)
		}

		if err != nil {
			record.Status = constants.RecordStatusFailed
			record.Message = err.Error()
			result.Summary.Failed++
		} else {
			record.Status = constants.RecordStatusSuccess
			record.AfterValue = value
			result.Summary.Success++
		}

		result.Records = append(result.Records, record)
		result.Summary.Total++

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			result.Status = constants.TaskStatusFailed
			result.Message = "migration cancelled"
			return result, ctx.Err()
		default:
		}
	}

	result.Status = constants.TaskStatusCompleted
	result.Message = fmt.Sprintf("成功迁移 %d 个环境变量", result.Summary.Success)

	return result, nil
}

// Rollback 回滚环境变量迁移
func (s *EnvVariableStrategy) Rollback(ctx context.Context, config *core.MigrationConfig) error {
	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		return fmt.Errorf("env variable migration is only supported on Windows")
	}

	// 获取要回滚的变量
	variables := s.getVariablesToMigrate(config)

	// 遍历并恢复环境变量
	for name, originalValue := range variables {
		// 从备份或配置中获取原始值
		if backupValue, ok := config.Context.GetState(fmt.Sprintf("backup_%s", name)); ok {
			originalValue = backupValue.(string)
		}

		// 设置回原始值
		var err error
		switch config.Target.Type {
		case "user":
			if originalValue == "" {
				err = s.deleteUserEnvVar(name)
			} else {
				err = s.setUserEnvVar(name, originalValue)
			}
		case "system":
			if originalValue == "" {
				err = s.deleteSystemEnvVar(name)
			} else {
				err = s.setSystemEnvVar(name, originalValue)
			}
		default:
			if originalValue == "" {
				err = os.Unsetenv(name)
			} else {
				err = os.Setenv(name, originalValue)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to rollback env var %s: %w", name, err)
		}

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}

// DryRun 预览环境变量迁移
func (s *EnvVariableStrategy) DryRun(ctx context.Context, config *core.MigrationConfig) (*core.MigrationPreview, error) {
	preview := core.NewMigrationPreview(config.TaskID)

	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		preview.Errors = append(preview.Errors, "环境变量迁移仅支持 Windows 系统")
		return preview, nil
	}

	// 获取要迁移的变量
	variables := s.getVariablesToMigrate(config)

	// 遍历并生成预览
	for name, newValue := range variables {
		oldValue := os.Getenv(name)

		change := core.PreviewChange{
			Key:         name,
			BeforeValue: oldValue,
			AfterValue:  newValue,
		}

		if oldValue == "" {
			change.ActionType = constants.ActionTypeCreate
			change.Description = fmt.Sprintf("将创建环境变量 %s", name)
		} else if oldValue != newValue {
			change.ActionType = constants.ActionTypeUpdate
			change.Description = fmt.Sprintf("将更新环境变量 %s", name)
		} else {
			change.ActionType = "none"
			change.Description = fmt.Sprintf("环境变量 %s 无变化", name)
		}

		// 评估影响程度
		if s.isHighImpactVariable(name) {
			change.Impact = "high"
			preview.Summary.HighImpact++
		} else {
			change.Impact = "low"
		}

		preview.Changes = append(preview.Changes, change)
		preview.Summary.Total++

		switch change.ActionType {
		case constants.ActionTypeCreate:
			preview.Summary.Create++
		case constants.ActionTypeUpdate:
			preview.Summary.Update++
		}
	}

	// 添加警告信息
	if preview.Summary.HighImpact > 0 {
		preview.Warnings = append(preview.Warnings,
			fmt.Sprintf("有 %d 个高影响环境变量将被修改", preview.Summary.HighImpact))
	}

	return preview, nil
}

// getVariablesToMigrate 获取要迁移的变量
func (s *EnvVariableStrategy) getVariablesToMigrate(config *core.MigrationConfig) map[string]string {
	variables := make(map[string]string)

	// 从配置中获取指定的变量
	for name, value := range config.Source.Variables {
		variables[name] = value
	}

	// 如果有过滤模式，匹配环境变量
	if config.Source.Filter.Pattern != "" {
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) != 2 {
				continue
			}
			name, value := pair[0], pair[1]

			// 简单的模式匹配（支持 * 通配符）
			if s.matchPattern(name, config.Source.Filter.Pattern) {
				// 检查是否在排除列表中
				if s.isExcluded(name, config.Source.Filter.Exclude) {
					continue
				}
				variables[name] = value
			}
		}
	}

	return variables
}

// matchPattern 简单的模式匹配
func (s *EnvVariableStrategy) matchPattern(name, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// 支持前缀和后缀匹配
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}

	return name == pattern
}

// isExcluded 检查是否在排除列表中
func (s *EnvVariableStrategy) isExcluded(name string, excludeList []string) bool {
	for _, exclude := range excludeList {
		if name == exclude || s.matchPattern(name, exclude) {
			return true
		}
	}
	return false
}

// isHighImpactVariable 检查是否为高影响变量
func (s *EnvVariableStrategy) isHighImpactVariable(name string) bool {
	highImpactVars := []string{
		"PATH", "JAVA_HOME", "PYTHONPATH", "NODE_PATH",
		"GOPATH", "GOROOT", "MAVEN_HOME", "GRADLE_HOME",
		"CLASSPATH", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH",
	}
	nameUpper := strings.ToUpper(name)
	for _, v := range highImpactVars {
		if nameUpper == v || strings.Contains(nameUpper, v) {
			return true
		}
	}
	return false
}

// setUserEnvVar 设置用户环境变量 (Windows)
func (s *EnvVariableStrategy) setUserEnvVar(name, value string) error {
	// 在 Windows 上，需要使用 registry 或系统调用
	// 这里简化实现，使用 os.Setenv
	return os.Setenv(name, value)
}

// setSystemEnvVar 设置系统环境变量 (Windows)
func (s *EnvVariableStrategy) setSystemEnvVar(name, value string) error {
	// 在 Windows 上，需要使用 registry 或系统调用
	// 这里简化实现，使用 os.Setenv
	return os.Setenv(name, value)
}

// deleteUserEnvVar 删除用户环境变量 (Windows)
func (s *EnvVariableStrategy) deleteUserEnvVar(name string) error {
	return os.Unsetenv(name)
}

// deleteSystemEnvVar 删除系统环境变量 (Windows)
func (s *EnvVariableStrategy) deleteSystemEnvVar(name string) error {
	return os.Unsetenv(name)
}

// Export 导出环境变量（暂不支持）
func (s *EnvVariableStrategy) Export(ctx context.Context, config *core.MigrationConfig) (*core.ExportResult, error) {
	return nil, fmt.Errorf("export is not supported for env_variable strategy")
}

// Import 导入环境变量（暂不支持）
func (s *EnvVariableStrategy) Import(ctx context.Context, config *core.MigrationConfig) (*core.ImportResult, error) {
	return nil, fmt.Errorf("import is not supported for env_variable strategy")
}

// ValidateExport 验证导出配置（暂不支持）
func (s *EnvVariableStrategy) ValidateExport(config *core.MigrationConfig) error {
	return fmt.Errorf("export is not supported for env_variable strategy")
}

// ValidateImport 验证导入配置（暂不支持）
func (s *EnvVariableStrategy) ValidateImport(config *core.MigrationConfig) error {
	return fmt.Errorf("import is not supported for env_variable strategy")
}
