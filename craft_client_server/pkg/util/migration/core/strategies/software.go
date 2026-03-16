package strategies

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tsc/pkg/util/migration/constants"
	"tsc/pkg/util/migration/core"
)

func init() {
	// 自动注册到全局注册表
	if err := core.RegisterStrategy(&SoftwareStrategy{}); err != nil {
		panic(fmt.Sprintf("failed to register software strategy: %v", err))
	}
}

// SoftwareStrategy 软件配置迁移策略
type SoftwareStrategy struct{}

// Name 返回策略名称
func (s *SoftwareStrategy) Name() string {
	return "软件配置迁移策略"
}

// Type 返回策略类型
func (s *SoftwareStrategy) Type() core.MigrationType {
	return constants.MigrationTypeSoftware
}

// Description 返回策略描述
func (s *SoftwareStrategy) Description() string {
	return "迁移软件配置，包括配置目录、数据文件和注册表项"
}

// Validate 验证配置是否有效
func (s *SoftwareStrategy) Validate(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Source.Path == "" {
		return fmt.Errorf("source path is required")
	}

	if config.Target.Path == "" {
		return fmt.Errorf("target path is required")
	}

	return nil
}

// Execute 执行软件配置迁移
func (s *SoftwareStrategy) Execute(ctx context.Context, config *core.MigrationConfig) (*core.MigrationResult, error) {
	result := core.NewMigrationResult(config.TaskID)
	result.StartTime = time.Now()
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 获取软件名称（从路径或配置中）
	softwareName := config.Name
	if softwareName == "" {
		softwareName = filepath.Base(config.Source.Path)
	}

	// 检查源路径是否存在
	sourceInfo, err := os.Stat(config.Source.Path)
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("源路径不存在或无法访问: %v", err)
		return result, err
	}

	// 备份目标路径（如果需要）
	if config.Target.Backup {
		backupPath := config.Target.BackupPath
		if backupPath == "" {
			backupPath = config.Target.Path + ".backup"
		}
		if _, err := os.Stat(config.Target.Path); err == nil {
			if err := s.backupDirectory(config.Target.Path, backupPath); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("备份失败: %v", err))
			} else {
				record := core.MigrationRecord{
					StepName:    "备份目标目录",
					ActionType:  constants.ActionTypeCopy,
					Key:         backupPath,
					BeforeValue: config.Target.Path,
					AfterValue:  backupPath,
					Status:      constants.RecordStatusSuccess,
					Timestamp:   time.Now(),
				}
				result.Records = append(result.Records, record)
			}
		}
	}

	// 根据源类型执行迁移
	var migrationErr error
	if sourceInfo.IsDir() {
		migrationErr = s.migrateDirectory(ctx, config, result)
	} else {
		migrationErr = s.migrateFile(ctx, config, result)
	}

	if migrationErr != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("迁移失败: %v", migrationErr)
		return result, migrationErr
	}

	// 迁移注册表项（如果配置了）
	if registryPath, ok := config.Source.Variables["registry_path"]; ok && registryPath != "" {
		if err := s.migrateRegistry(ctx, config, registryPath, result); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("注册表迁移失败: %v", err))
		}
	}

	result.Status = constants.TaskStatusCompleted
	result.Summary.Total = result.Summary.Success + result.Summary.Failed + result.Summary.Skipped
	result.Message = fmt.Sprintf("成功迁移软件配置 %s，共处理 %d 项", softwareName, result.Summary.Success)

	return result, nil
}

// Rollback 回滚软件配置迁移
func (s *SoftwareStrategy) Rollback(ctx context.Context, config *core.MigrationConfig) error {
	// 检查备份是否存在
	backupPath := config.Target.BackupPath
	if backupPath == "" {
		backupPath = config.Target.Path + ".backup"
	}

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not found: %s", backupPath)
	}

	// 删除当前目标
	if err := os.RemoveAll(config.Target.Path); err != nil {
		return fmt.Errorf("failed to remove current files: %w", err)
	}

	// 恢复备份
	if err := s.copyDirectory(backupPath, config.Target.Path); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// DryRun 预览软件配置迁移
func (s *SoftwareStrategy) DryRun(ctx context.Context, config *core.MigrationConfig) (*core.MigrationPreview, error) {
	preview := core.NewMigrationPreview(config.TaskID)

	// 检查源路径
	sourceInfo, err := os.Stat(config.Source.Path)
	if err != nil {
		preview.Errors = append(preview.Errors, fmt.Sprintf("源路径不存在: %s", config.Source.Path))
		return preview, nil
	}

	// 检查目标路径
	targetExists := false
	if _, err := os.Stat(config.Target.Path); err == nil {
		targetExists = true
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("目标路径已存在，可能会覆盖: %s", config.Target.Path))
	}

	// 计算要迁移的文件数量
	var fileCount, dirCount int64
	if sourceInfo.IsDir() {
		filepath.Walk(config.Source.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				dirCount++
			} else {
				fileCount++
			}
			return nil
		})
	} else {
		fileCount = 1
	}

	// 添加预览变更
	if fileCount > 0 {
		change := core.PreviewChange{
			ActionType:  constants.ActionTypeCopy,
			Key:         config.Source.Path,
			BeforeValue: fmt.Sprintf("不存在"),
			AfterValue:  config.Target.Path,
			Impact:      s.getImpactLevel(config.Source.Path),
			Description: fmt.Sprintf("将复制 %d 个文件和 %d 个目录", fileCount, dirCount),
		}
		if targetExists {
			change.BeforeValue = "已存在"
			change.ActionType = constants.ActionTypeUpdate
		}
		preview.Changes = append(preview.Changes, change)
		preview.Summary.Total++
		preview.Summary.Create += int(fileCount)
	}

	// 检查注册表项
	if registryPath, ok := config.Source.Variables["registry_path"]; ok && registryPath != "" {
		preview.Changes = append(preview.Changes, core.PreviewChange{
			ActionType:  constants.ActionTypeExport,
			Key:         registryPath,
			Description: fmt.Sprintf("将迁移注册表项: %s", registryPath),
			Impact:      "medium",
		})
		preview.Summary.Total++
	}

	return preview, nil
}

// migrateDirectory 迁移目录
func (s *SoftwareStrategy) migrateDirectory(ctx context.Context, config *core.MigrationConfig, result *core.MigrationResult) error {
	// 确保目标目录存在
	if err := os.MkdirAll(config.Target.Path, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 遍历源目录
	return filepath.Walk(config.Source.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 计算相对路径
		relPath, err := filepath.Rel(config.Source.Path, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(config.Target.Path, relPath)

		record := core.MigrationRecord{
			StepName:  fmt.Sprintf("迁移 %s", relPath),
			Key:       relPath,
			Timestamp: time.Now(),
		}

		if info.IsDir() {
			// 创建目录
			record.ActionType = constants.ActionTypeCreate
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				record.Status = constants.RecordStatusFailed
				record.Message = err.Error()
				result.Summary.Failed++
			} else {
				record.Status = constants.RecordStatusSuccess
				result.Summary.Success++
			}
		} else {
			// 复制文件
			record.ActionType = constants.ActionTypeCopy
			record.BeforeValue = path
			record.AfterValue = targetPath

			if err := s.copyFile(path, targetPath); err != nil {
				record.Status = constants.RecordStatusFailed
				record.Message = err.Error()
				result.Summary.Failed++
			} else {
				record.Status = constants.RecordStatusSuccess
				result.Summary.Success++
			}
		}

		result.Records = append(result.Records, record)
		return nil
	})
}

// migrateFile 迁移单个文件
func (s *SoftwareStrategy) migrateFile(ctx context.Context, config *core.MigrationConfig, result *core.MigrationResult) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(config.Target.Path)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	record := core.MigrationRecord{
		StepName:    "迁移文件",
		ActionType:  constants.ActionTypeCopy,
		Key:         filepath.Base(config.Source.Path),
		BeforeValue: config.Source.Path,
		AfterValue:  config.Target.Path,
		Timestamp:   time.Now(),
	}

	if err := s.copyFile(config.Source.Path, config.Target.Path); err != nil {
		record.Status = constants.RecordStatusFailed
		record.Message = err.Error()
		result.Summary.Failed++
	} else {
		record.Status = constants.RecordStatusSuccess
		result.Summary.Success++
	}

	result.Records = append(result.Records, record)
	return nil
}

// migrateRegistry 迁移注册表项
func (s *SoftwareStrategy) migrateRegistry(ctx context.Context, config *core.MigrationConfig, registryPath string, result *core.MigrationResult) error {
	// 创建注册表策略实例
	registryStrategy := &RegistryStrategy{}

	// 构建注册表迁移配置
	registryConfig := &core.MigrationConfig{
		TaskID: config.TaskID,
		Type:   constants.MigrationTypeRegistry,
		Source: core.MigrationSource{
			Path: registryPath,
		},
		Target: core.MigrationTarget{
			Path: registryPath, // 同一注册表路径
		},
		Context: config.Context,
	}

	// 使用注册表策略迁移
	registryResult, err := registryStrategy.Execute(ctx, registryConfig)
	if err != nil {
		return err
	}

	// 合并结果
	result.Records = append(result.Records, registryResult.Records...)
	result.Summary.Success += registryResult.Summary.Success
	result.Summary.Failed += registryResult.Summary.Failed

	return nil
}

// backupDirectory 备份目录
func (s *SoftwareStrategy) backupDirectory(src, dst string) error {
	return s.copyDirectory(src, dst)
}

// copyDirectory 复制目录
func (s *SoftwareStrategy) copyDirectory(src, dst string) error {
	// 获取源目录信息
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	// 遍历并复制
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制文件
func (s *SoftwareStrategy) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := sourceFile.Read(buf)
		if n > 0 {
			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}

// getImpactLevel 获取影响级别
func (s *SoftwareStrategy) getImpactLevel(path string) string {
	highImpactPatterns := []string{
		"config", "settings", ".env", "credentials",
		"secret", "key", "password", "token",
	}

	pathLower := strings.ToLower(path)
	for _, pattern := range highImpactPatterns {
		if strings.Contains(pathLower, pattern) {
			return "high"
		}
	}
	return "low"
}

// Export 导出软件配置（暂不支持）
func (s *SoftwareStrategy) Export(ctx context.Context, config *core.MigrationConfig) (*core.ExportResult, error) {
	return nil, fmt.Errorf("export is not supported for software strategy")
}

// Import 导入软件配置（暂不支持）
func (s *SoftwareStrategy) Import(ctx context.Context, config *core.MigrationConfig) (*core.ImportResult, error) {
	return nil, fmt.Errorf("import is not supported for software strategy")
}

// ValidateExport 验证导出配置（暂不支持）
func (s *SoftwareStrategy) ValidateExport(config *core.MigrationConfig) error {
	return fmt.Errorf("export is not supported for software strategy")
}

// ValidateImport 验证导入配置（暂不支持）
func (s *SoftwareStrategy) ValidateImport(config *core.MigrationConfig) error {
	return fmt.Errorf("import is not supported for software strategy")
}
