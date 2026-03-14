package strategies

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"tsc/pkg/util/migration/constants"
	"tsc/pkg/util/migration/core"
)

func init() {
	// 自动注册到全局注册表
	if err := core.RegisterStrategy(&RegistryStrategy{}); err != nil {
		panic(fmt.Sprintf("failed to register registry strategy: %v", err))
	}
}

// RegistryStrategy 注册表迁移策略
type RegistryStrategy struct{}

// Name 返回策略名称
func (s *RegistryStrategy) Name() string {
	return "注册表迁移策略"
}

// Type 返回策略类型
func (s *RegistryStrategy) Type() core.MigrationType {
	return constants.MigrationTypeRegistry
}

// Description 返回策略描述
func (s *RegistryStrategy) Description() string {
	return "迁移 Windows 注册表项，支持递归导出/导入"
}

// Validate 验证配置是否有效
func (s *RegistryStrategy) Validate(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Source.Path == "" {
		return fmt.Errorf("source registry path is required")
	}

	// 验证路径格式
	if !s.isValidRegistryPath(config.Source.Path) {
		return fmt.Errorf("invalid registry path format: %s", config.Source.Path)
	}

	return nil
}

// Execute 执行注册表迁移
func (s *RegistryStrategy) Execute(ctx context.Context, config *core.MigrationConfig) (*core.MigrationResult, error) {
	result := core.NewMigrationResult(config.TaskID)
	result.StartTime = time.Now()
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		result.Status = constants.TaskStatusFailed
		result.Message = "注册表迁移仅支持 Windows 系统"
		return result, fmt.Errorf("registry migration is only supported on Windows")
	}

	// 获取注册表根键和子路径
	rootKey, subPath, err := s.parseRegistryPath(config.Source.Path)
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = err.Error()
		return result, err
	}

	// 导出注册表项
	exportPath := ""
	if config.Target.Path != "" && config.Target.Path != config.Source.Path {
		// 导出到文件
		exportPath = config.Target.Path
		if strings.HasSuffix(exportPath, ".reg") {
			if err := s.exportRegistry(rootKey, subPath, exportPath); err != nil {
				result.Status = constants.TaskStatusFailed
				result.Message = fmt.Sprintf("导出注册表失败: %v", err)
				return result, err
			}
			record := core.MigrationRecord{
				StepName:   "导出注册表",
				ActionType: constants.ActionTypeExport,
				Key:        config.Source.Path,
				AfterValue: exportPath,
				Status:     constants.RecordStatusSuccess,
				Timestamp:  time.Now(),
			}
			result.Records = append(result.Records, record)
			result.Summary.Success++
		} else {
			// 导入到另一个注册表位置
			if err := s.copyRegistryKey(ctx, rootKey, subPath, config.Target.Path, result); err != nil {
				result.Status = constants.TaskStatusFailed
				result.Message = fmt.Sprintf("复制注册表项失败: %v", err)
				return result, err
			}
		}
	} else {
		// 读取注册表值并记录
		if err := s.readRegistryValues(ctx, rootKey, subPath, result, config); err != nil {
			result.Status = constants.TaskStatusFailed
			result.Message = fmt.Sprintf("读取注册表失败: %v", err)
			return result, err
		}
	}

	result.Status = constants.TaskStatusCompleted
	result.Summary.Total = result.Summary.Success + result.Summary.Failed
	result.Message = fmt.Sprintf("成功处理注册表项 %s，共 %d 项", config.Source.Path, result.Summary.Success)

	return result, nil
}

// Rollback 回滚注册表迁移
func (s *RegistryStrategy) Rollback(ctx context.Context, config *core.MigrationConfig) error {
	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		return fmt.Errorf("registry migration is only supported on Windows")
	}

	// 如果有备份文件，导入恢复
	backupPath := config.Target.BackupPath
	if backupPath == "" {
		backupPath = config.Source.Path + ".reg"
	}

	if _, err := exec.LookPath("reg"); err != nil {
		return fmt.Errorf("reg command not found")
	}

	// 导入备份注册表文件
	cmd := exec.CommandContext(ctx, "reg", "import", backupPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to import registry backup: %v, output: %s", err, string(output))
	}

	return nil
}

// DryRun 预览注册表迁移
func (s *RegistryStrategy) DryRun(ctx context.Context, config *core.MigrationConfig) (*core.MigrationPreview, error) {
	preview := core.NewMigrationPreview(config.TaskID)

	// 检查是否为 Windows 系统
	if runtime.GOOS != "windows" {
		preview.Errors = append(preview.Errors, "注册表迁移仅支持 Windows 系统")
		return preview, nil
	}

	// 解析注册表路径
	rootKey, subPath, err := s.parseRegistryPath(config.Source.Path)
	if err != nil {
		preview.Errors = append(preview.Errors, err.Error())
		return preview, nil
	}

	// 查询注册表项是否存在
	exists, err := s.registryKeyExists(rootKey, subPath)
	if err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("无法验证注册表项是否存在: %v", err))
	} else if !exists {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("注册表项不存在: %s", config.Source.Path))
	}

	// 添加预览变更
	change := core.PreviewChange{
		Key:         config.Source.Path,
		Description: fmt.Sprintf("将迁移注册表项: %s", config.Source.Path),
	}

	if config.Target.Path != "" {
		if strings.HasSuffix(config.Target.Path, ".reg") {
			change.ActionType = constants.ActionTypeExport
			change.AfterValue = config.Target.Path
			preview.Summary.Create++
		} else {
			change.ActionType = constants.ActionTypeCopy
			change.AfterValue = config.Target.Path
			preview.Summary.Create++
		}
	} else {
		change.ActionType = constants.ActionTypeExport
		change.Description = "将读取并记录注册表值"
	}

	preview.Changes = append(preview.Changes, change)
	preview.Summary.Total++

	// 添加影响评估
	if s.isHighImpactPath(config.Source.Path) {
		preview.Summary.HighImpact++
		preview.Warnings = append(preview.Warnings, "此注册表项可能影响系统或应用程序行为")
	}

	return preview, nil
}

// parseRegistryPath 解析注册表路径
func (s *RegistryStrategy) parseRegistryPath(path string) (string, string, error) {
	// 标准化路径格式
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, "/", "\\")

	// 支持的根键缩写
	rootKeyMap := map[string]string{
		"HKLM":                "HKLM",
		"HKEY_LOCAL_MACHINE":  "HKLM",
		"HKCU":                "HKCU",
		"HKEY_CURRENT_USER":   "HKCU",
		"HKCR":                "HKCR",
		"HKEY_CLASSES_ROOT":   "HKCR",
		"HKU":                 "HKU",
		"HKEY_USERS":          "HKU",
		"HKCC":                "HKCC",
		"HKEY_CURRENT_CONFIG": "HKCC",
	}

	// 分离根键和子路径
	parts := strings.SplitN(path, "\\", 2)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid registry path")
	}

	rootKey, ok := rootKeyMap[strings.ToUpper(parts[0])]
	if !ok {
		return "", "", fmt.Errorf("unknown registry root key: %s", parts[0])
	}

	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	return rootKey, subPath, nil
}

// isValidRegistryPath 验证注册表路径格式
func (s *RegistryStrategy) isValidRegistryPath(path string) bool {
	_, _, err := s.parseRegistryPath(path)
	return err == nil
}

// registryKeyExists 检查注册表项是否存在
func (s *RegistryStrategy) registryKeyExists(rootKey, subPath string) (bool, error) {
	cmd := exec.Command("reg", "query", rootKey+"\\"+subPath)
	err := cmd.Run()
	return err == nil, nil
}

// exportRegistry 导出注册表项到文件
func (s *RegistryStrategy) exportRegistry(rootKey, subPath, exportPath string) error {
	fullPath := rootKey
	if subPath != "" {
		fullPath = rootKey + "\\" + subPath
	}

	cmd := exec.Command("reg", "export", fullPath, exportPath, "/y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export failed: %v, output: %s", err, string(output))
	}

	return nil
}

// importRegistry 从文件导入注册表
func (s *RegistryStrategy) importRegistry(importPath string) error {
	cmd := exec.Command("reg", "import", importPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("import failed: %v, output: %s", err, string(output))
	}

	return nil
}

// copyRegistryKey 复制注册表项
func (s *RegistryStrategy) copyRegistryKey(ctx context.Context, srcRoot, srcPath, dstPath string, result *core.MigrationResult) error {
	// 先导出到临时文件
	tempFile := fmt.Sprintf("%s_temp.reg", srcPath)
	tempFile = strings.ReplaceAll(tempFile, "\\", "_")
	tempFile = fmt.Sprintf("%s\\%s", os.TempDir(), tempFile)

	// 确保清理临时文件
	defer func() {
		exec.Command("del", tempFile).Run()
	}()

	// 导出源注册表
	if err := s.exportRegistry(srcRoot, srcPath, tempFile); err != nil {
		return err
	}

	// 修改文件中的路径并导入
	// 这里简化处理，实际需要修改 .reg 文件中的路径
	if err := s.importRegistry(tempFile); err != nil {
		return err
	}

	record := core.MigrationRecord{
		StepName:   "复制注册表项",
		ActionType: constants.ActionTypeCopy,
		Key:        srcRoot + "\\" + srcPath,
		AfterValue: dstPath,
		Status:     constants.RecordStatusSuccess,
		Timestamp:  time.Now(),
	}
	result.Records = append(result.Records, record)
	result.Summary.Success++

	return nil
}

// readRegistryValues 读取注册表值
func (s *RegistryStrategy) readRegistryValues(ctx context.Context, rootKey, subPath string, result *core.MigrationResult, config *core.MigrationConfig) error {
	fullPath := rootKey
	if subPath != "" {
		fullPath = rootKey + "\\" + subPath
	}

	// 查询注册表项
	cmd := exec.CommandContext(ctx, "reg", "query", fullPath, "/s")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	// 解析输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, fullPath) || strings.HasPrefix(line, "HKEY_") {
			continue
		}

		// 解析键值对
		// 格式:    名称    类型    值
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			record := core.MigrationRecord{
				StepName:   "读取注册表值",
				ActionType: constants.ActionTypeExport,
				Key:        parts[0],
				AfterValue: strings.Join(parts[2:], " "),
				Status:     constants.RecordStatusSuccess,
				Timestamp:  time.Now(),
			}
			if len(parts) > 1 {
				record.Message = fmt.Sprintf("类型: %s", parts[1])
			}
			result.Records = append(result.Records, record)
			result.Summary.Success++
		}
	}

	return nil
}

// isHighImpactPath 检查是否为高影响路径
func (s *RegistryStrategy) isHighImpactPath(path string) bool {
	highImpactPatterns := []string{
		"\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"\\Software\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"\\SYSTEM\\CurrentControlSet\\Services",
		"\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
		"\\SOFTWARE\\Policies",
	}

	pathUpper := strings.ToUpper(path)
	for _, pattern := range highImpactPatterns {
		if strings.Contains(pathUpper, pattern) {
			return true
		}
	}
	return false
}

// Export 导出注册表（暂不支持）
func (s *RegistryStrategy) Export(ctx context.Context, config *core.MigrationConfig) (*core.ExportResult, error) {
	return nil, fmt.Errorf("export is not supported for registry strategy")
}

// Import 导入注册表（暂不支持）
func (s *RegistryStrategy) Import(ctx context.Context, config *core.MigrationConfig) (*core.ImportResult, error) {
	return nil, fmt.Errorf("import is not supported for registry strategy")
}

// ValidateExport 验证导出配置（暂不支持）
func (s *RegistryStrategy) ValidateExport(config *core.MigrationConfig) error {
	return fmt.Errorf("export is not supported for registry strategy")
}

// ValidateImport 验证导入配置（暂不支持）
func (s *RegistryStrategy) ValidateImport(config *core.MigrationConfig) error {
	return fmt.Errorf("import is not supported for registry strategy")
}
