package strategies

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"

	"tsc/pkg/util/migration/constants"
	"tsc/pkg/util/migration/core"
)

func init() {
	// 自动注册到全局注册表
	if err := core.RegisterStrategy(&ConfigFileStrategy{}); err != nil {
		panic(fmt.Sprintf("failed to register config file strategy: %v", err))
	}
}

// ConfigFileStrategy 配置文件迁移策略
type ConfigFileStrategy struct{}

// Name 返回策略名称
func (s *ConfigFileStrategy) Name() string {
	return "配置文件迁移策略"
}

// Type 返回策略类型
func (s *ConfigFileStrategy) Type() core.MigrationType {
	return constants.MigrationTypeConfigFile
}

// Description 返回策略描述
func (s *ConfigFileStrategy) Description() string {
	return "迁移配置文件，支持 JSON/YAML/INI/TOML 格式的配置文件迁移"
}

// Validate 验证配置是否有效
func (s *ConfigFileStrategy) Validate(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Source.Path == "" {
		return fmt.Errorf("source path is required")
	}

	if config.Target.Path == "" {
		return fmt.Errorf("target path is required")
	}

	// 验证文件格式
	if config.Source.Format == "" && config.Target.Format == "" {
		// 尝试从文件名推断格式
		ext := strings.ToLower(filepath.Ext(config.Source.Path))
		supportedFormats := []string{".json", ".yaml", ".yml", ".ini", ".toml"}
		valid := false
		for _, f := range supportedFormats {
			if ext == f {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported file format: %s", ext)
		}
	}

	return nil
}

// Execute 执行配置文件迁移
func (s *ConfigFileStrategy) Execute(ctx context.Context, config *core.MigrationConfig) (*core.MigrationResult, error) {
	result := core.NewMigrationResult(config.TaskID)
	result.StartTime = time.Now()
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 读取源配置文件
	sourceData, err := s.readConfigFile(config.Source.Path, config.Source.Format, config.Source.Encoding)
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("读取源配置文件失败: %v", err)
		return result, err
	}

	// 备份目标文件（如果需要）
	if config.Target.Backup {
		backupPath := config.Target.BackupPath
		if backupPath == "" {
			backupPath = config.Target.Path + ".backup"
		}
		if err := s.backupFile(config.Target.Path, backupPath); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("备份文件失败: %v", err))
		}
	}

	// 读取目标配置文件（如果存在）
	var targetData map[string]interface{}
	if _, err := os.Stat(config.Target.Path); err == nil {
		targetData, err = s.readConfigFile(config.Target.Path, config.Target.Format, config.Target.Encoding)
		if err != nil {
			targetData = make(map[string]interface{})
		}
	} else {
		targetData = make(map[string]interface{})
	}

	// 应用过滤条件
	filteredSource := s.applyFilter(sourceData, config.Source.Filter)

	// 根据合并模式处理
	var mergedData map[string]interface{}
	switch config.Target.MergeMode {
	case "overwrite":
		mergedData = filteredSource
	case "merge":
		mergedData = s.mergeConfig(targetData, filteredSource)
	case "skip":
		mergedData = targetData
		for k, v := range filteredSource {
			if _, exists := targetData[k]; !exists {
				mergedData[k] = v
			}
		}
	default:
		mergedData = filteredSource
	}

	// 记录变更
	for key, newValue := range mergedData {
		record := core.MigrationRecord{
			StepName:   fmt.Sprintf("迁移配置项 %s", key),
			ActionType: s.getActionType(targetData, key),
			Key:        key,
			AfterValue: fmt.Sprintf("%v", newValue),
			Timestamp:  time.Now(),
		}

		if oldValue, exists := targetData[key]; exists {
			record.BeforeValue = fmt.Sprintf("%v", oldValue)
		}

		record.Status = constants.RecordStatusSuccess
		result.Records = append(result.Records, record)
		result.Summary.Total++
		result.Summary.Success++
	}

	// 写入目标文件
	format := config.Target.Format
	if format == "" {
		format = config.Source.Format
	}
	if err := s.writeConfigFile(config.Target.Path, mergedData, format, config.Target.Encoding); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("写入目标配置文件失败: %v", err)
		return result, err
	}

	// 如果需要创建不存在的目录
	if config.Target.CreateIfNotExists {
		dir := filepath.Dir(config.Target.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("创建目录失败: %v", err))
		}
	}

	result.Status = constants.TaskStatusCompleted
	result.Message = fmt.Sprintf("成功迁移配置文件，共 %d 项", result.Summary.Success)

	return result, nil
}

// Rollback 回滚配置文件迁移
func (s *ConfigFileStrategy) Rollback(ctx context.Context, config *core.MigrationConfig) error {
	// 检查是否有备份文件
	backupPath := config.Target.BackupPath
	if backupPath == "" {
		backupPath = config.Target.Path + ".backup"
	}

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// 恢复备份文件
	if err := s.copyFile(backupPath, config.Target.Path); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// DryRun 预览配置文件迁移
func (s *ConfigFileStrategy) DryRun(ctx context.Context, config *core.MigrationConfig) (*core.MigrationPreview, error) {
	preview := core.NewMigrationPreview(config.TaskID)

	// 检查源文件是否存在
	if _, err := os.Stat(config.Source.Path); os.IsNotExist(err) {
		preview.Errors = append(preview.Errors, fmt.Sprintf("源配置文件不存在: %s", config.Source.Path))
		return preview, nil
	}

	// 读取源配置文件
	sourceData, err := s.readConfigFile(config.Source.Path, config.Source.Format, config.Source.Encoding)
	if err != nil {
		preview.Errors = append(preview.Errors, fmt.Sprintf("读取源配置文件失败: %v", err))
		return preview, nil
	}

	// 读取目标配置文件（如果存在）
	var targetData map[string]interface{}
	if _, err := os.Stat(config.Target.Path); err == nil {
		targetData, err = s.readConfigFile(config.Target.Path, config.Target.Format, config.Target.Encoding)
		if err != nil {
			targetData = make(map[string]interface{})
		}
	} else {
		targetData = make(map[string]interface{})
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("目标配置文件不存在，将创建新文件: %s", config.Target.Path))
	}

	// 应用过滤条件
	filteredSource := s.applyFilter(sourceData, config.Source.Filter)

	// 生成预览
	for key, newValue := range filteredSource {
		change := core.PreviewChange{
			Key:        key,
			AfterValue: fmt.Sprintf("%v", newValue),
		}

		if oldValue, exists := targetData[key]; exists {
			change.BeforeValue = fmt.Sprintf("%v", oldValue)
			if change.BeforeValue != change.AfterValue {
				change.ActionType = constants.ActionTypeUpdate
				change.Description = fmt.Sprintf("将更新配置项 %s", key)
				preview.Summary.Update++
			} else {
				change.ActionType = "none"
				change.Description = fmt.Sprintf("配置项 %s 无变化", key)
			}
		} else {
			change.ActionType = constants.ActionTypeCreate
			change.Description = fmt.Sprintf("将创建配置项 %s", key)
			preview.Summary.Create++
		}

		preview.Changes = append(preview.Changes, change)
		preview.Summary.Total++
	}

	return preview, nil
}

// readConfigFile 读取配置文件
func (s *ConfigFileStrategy) readConfigFile(path, format, encoding string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 如果未指定格式，从文件扩展名推断
	if format == "" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		case ".ini":
			format = "ini"
		case ".toml":
			format = "toml"
		default:
			return nil, fmt.Errorf("unsupported file format: %s", ext)
		}
	}

	// 根据格式解析
	switch strings.ToLower(format) {
	case "json":
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case "ini":
		cfg, err := ini.Load(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse INI: %w", err)
		}
		for _, section := range cfg.Sections() {
			for _, key := range section.Keys() {
				fullKey := section.Name() + "." + key.Name()
				if section.Name() == ini.DefaultSection {
					fullKey = key.Name()
				}
				data[fullKey] = key.Value()
			}
		}
	case "toml":
		// 简化的 TOML 支持，使用类似 INI 的方式
		cfg, err := ini.Load(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TOML: %w", err)
		}
		for _, section := range cfg.Sections() {
			for _, key := range section.Keys() {
				fullKey := section.Name() + "." + key.Name()
				if section.Name() == ini.DefaultSection {
					fullKey = key.Name()
				}
				data[fullKey] = key.Value()
			}
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return data, nil
}

// writeConfigFile 写入配置文件
func (s *ConfigFileStrategy) writeConfigFile(path string, data map[string]interface{}, format, encoding string) error {
	var content []byte
	var err error

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 根据格式序列化
	switch strings.ToLower(format) {
	case "json":
		content, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize JSON: %w", err)
		}
	case "yaml":
		content, err = yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to serialize YAML: %w", err)
		}
	case "ini":
		cfg := ini.Empty()
		for key, value := range data {
			cfg.Section("").Key(key).SetValue(fmt.Sprintf("%v", value))
		}
		content = []byte(cfgToString(cfg))
	case "toml":
		// 简化的 TOML 支持
		var sb strings.Builder
		for key, value := range data {
			sb.WriteString(fmt.Sprintf("%s = %v\n", key, value))
		}
		content = []byte(sb.String())
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// 写入文件
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// cfgToString 将 ini 配置转为字符串
func cfgToString(cfg *ini.File) string {
	var sb strings.Builder
	for _, section := range cfg.Sections() {
		if section.Name() != ini.DefaultSection {
			sb.WriteString(fmt.Sprintf("[%s]\n", section.Name()))
		}
		for _, key := range section.Keys() {
			sb.WriteString(fmt.Sprintf("%s = %s\n", key.Name(), key.Value()))
		}
	}
	return sb.String()
}

// applyFilter 应用过滤条件
func (s *ConfigFileStrategy) applyFilter(data map[string]interface{}, filter core.SourceFilter) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// 检查是否在排除列表中
		excluded := false
		for _, exclude := range filter.Exclude {
			if key == exclude || strings.HasPrefix(key, exclude+".") {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// 检查是否在包含列表中
		if len(filter.Include) > 0 {
			included := false
			for _, include := range filter.Include {
				if key == include || strings.HasPrefix(key, include+".") {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}

		result[key] = value
	}

	return result
}

// mergeConfig 合并配置
func (s *ConfigFileStrategy) mergeConfig(target, source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 复制目标配置
	for k, v := range target {
		result[k] = v
	}

	// 合并源配置
	for k, v := range source {
		if existing, ok := result[k]; ok {
			// 如果两者都是 map，递归合并
			if existingMap, ok1 := existing.(map[string]interface{}); ok1 {
				if sourceMap, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = s.mergeConfig(existingMap, sourceMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

// backupFile 备份文件
func (s *ConfigFileStrategy) backupFile(src, dst string) error {
	return s.copyFile(src, dst)
}

// copyFile 复制文件
func (s *ConfigFileStrategy) copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// getActionType 获取操作类型
func (s *ConfigFileStrategy) getActionType(data map[string]interface{}, key string) string {
	if _, exists := data[key]; exists {
		return constants.ActionTypeUpdate
	}
	return constants.ActionTypeCreate
}

// Export 导出配置文件
func (s *ConfigFileStrategy) Export(ctx context.Context, config *core.MigrationConfig) (*core.ExportResult, error) {
	result := core.NewExportResult(config.TaskID)
	result.ExportID = generateExportID()
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 1. 验证源文件
	if _, err := os.Stat(config.Source.Path); os.IsNotExist(err) {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("源配置文件不存在: %s", config.Source.Path)
		return result, err
	}

	// 2. 读取源配置
	sourceData, err := s.readConfigFile(config.Source.Path, config.Source.Format, config.Source.Encoding)
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("读取源配置失败: %v", err)
		return result, err
	}

	// 3. 应用过滤条件
	filteredData := s.applyFilter(sourceData, config.Source.Filter)

	// 4. 构建导出包
	exportPkg := core.NewExportPackage()
	exportPkg.Metadata.ExportID = result.ExportID
	exportPkg.Metadata.SourceType = string(constants.MigrationTypeConfigFile)
	exportPkg.Metadata.OriginalFormat = s.detectFormat(config.Source.Path, config.Source.Format)
	exportPkg.Metadata.OriginalPath = config.Source.Path
	exportPkg.Metadata.OriginalEncoding = config.Source.Encoding
	exportPkg.Metadata.Checksum = s.calculateChecksum(filteredData)

	exportPkg.Content.Data = filteredData

	// 5. 可选：包含原始内容
	if config.Options.IncludeRawContent {
		rawContent, err := os.ReadFile(config.Source.Path)
		if err == nil {
			exportPkg.Content.RawContent = base64.StdEncoding.EncodeToString(rawContent)
		}
	}

	// 6. 确定导出路径
	exportPath := config.Options.ExportPath
	if exportPath == "" {
		exportPath = config.Source.Path + ".export.json"
	}

	// 7. 确保导出目录存在
	exportDir := filepath.Dir(exportPath)
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("创建导出目录失败: %v", err)
		return result, err
	}

	// 8. 写入导出文件
	exportJSON, err := json.MarshalIndent(exportPkg, "", "  ")
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("序列化导出包失败: %v", err)
		return result, err
	}

	if err := os.WriteFile(exportPath, exportJSON, 0644); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("写入导出文件失败: %v", err)
		return result, err
	}

	// 9. 记录导出操作
	record := core.MigrationRecord{
		StepName:   "导出配置文件",
		ActionType: constants.ActionTypeExport,
		Key:        config.Source.Path,
		AfterValue: fmt.Sprintf("导出到 %s", exportPath),
		Status:     constants.RecordStatusSuccess,
		Timestamp:  time.Now(),
	}
	result.Records = append(result.Records, record)

	result.Status = constants.TaskStatusCompleted
	result.Message = fmt.Sprintf("成功导出配置文件到: %s", exportPath)
	result.ExportPath = exportPath
	result.Package = exportPkg

	return result, nil
}

// Import 导入配置文件
func (s *ConfigFileStrategy) Import(ctx context.Context, config *core.MigrationConfig) (*core.ImportResult, error) {
	result := core.NewImportResult(config.TaskID)
	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	// 1. 确定导入文件路径
	importPath := config.Options.ImportPath
	if importPath == "" {
		importPath = config.Source.Path
	}

	// 2. 读取导入文件
	importContent, err := os.ReadFile(importPath)
	if err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("读取导入文件失败: %v", err)
		return result, err
	}

	// 3. 解析导出包
	var exportPkg core.ExportPackage
	if err := json.Unmarshal(importContent, &exportPkg); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("解析导出包失败: %v", err)
		return result, err
	}

	result.SourcePackage = &exportPkg

	// 4. 验证导出包
	if err := s.validateExportPackage(&exportPkg); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("导出包验证失败: %v", err)
		return result, err
	}

	// 5. 确定目标格式
	targetFormat := config.Target.Format
	if config.Options.PreserveFormat || targetFormat == "" {
		targetFormat = exportPkg.Metadata.OriginalFormat
	}

	// 6. 确定目标路径
	targetPath := config.Target.Path
	if targetPath == "" {
		targetPath = exportPkg.Metadata.OriginalPath
	}

	// 7. 备份现有文件（如果需要）
	if config.Target.Backup {
		if _, err := os.Stat(targetPath); err == nil {
			backupPath := config.Target.BackupPath
			if backupPath == "" {
				backupPath = targetPath + ".backup"
			}
			if err := s.backupFile(targetPath, backupPath); err != nil {
				result.Status = constants.TaskStatusFailed
				result.Message = fmt.Sprintf("备份失败: %v", err)
				return result, err
			}
		}
	}

	// 8. 读取目标现有配置（如果存在）
	var targetData map[string]interface{}
	if _, err := os.Stat(targetPath); err == nil {
		targetData, _ = s.readConfigFile(targetPath, targetFormat, config.Target.Encoding)
	}
	if targetData == nil {
		targetData = make(map[string]interface{})
	}

	// 9. 应用合并策略
	var mergedData map[string]interface{}
	switch config.Target.MergeMode {
	case "overwrite":
		mergedData = exportPkg.Content.Data
	case "merge":
		mergedData = s.mergeConfig(targetData, exportPkg.Content.Data)
	case "skip":
		mergedData = targetData
		for k, v := range exportPkg.Content.Data {
			if _, exists := targetData[k]; !exists {
				mergedData[k] = v
			}
		}
	default:
		mergedData = exportPkg.Content.Data
	}

	// 10. 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("创建目标目录失败: %v", err)
		return result, err
	}

	// 11. 写入目标文件
	// 优先使用原始内容（如果有）
	if exportPkg.Content.RawContent != "" && config.Options.PreserveFormat {
		// 解码原始内容并直接写入
		rawBytes, err := base64.StdEncoding.DecodeString(exportPkg.Content.RawContent)
		if err != nil {
			result.Status = constants.TaskStatusFailed
			result.Message = fmt.Sprintf("解码原始内容失败: %v", err)
			return result, err
		}
		if err := os.WriteFile(targetPath, rawBytes, 0644); err != nil {
			result.Status = constants.TaskStatusFailed
			result.Message = fmt.Sprintf("写入目标文件失败: %v", err)
			return result, err
		}
		// 记录导入操作
		record := core.MigrationRecord{
			StepName:   "导入配置文件（原始内容）",
			ActionType: constants.ActionTypeImport,
			Key:        targetPath,
			AfterValue: fmt.Sprintf("从 %s 导入", importPath),
			Status:     constants.RecordStatusSuccess,
			Timestamp:  time.Now(),
		}
		result.Records = append(result.Records, record)
		result.Summary.Total++
		result.Summary.Success++
	} else if err := s.writeConfigFile(targetPath, mergedData, targetFormat, config.Target.Encoding); err != nil {
		result.Status = constants.TaskStatusFailed
		result.Message = fmt.Sprintf("写入目标文件失败: %v", err)
		return result, err
	} else {
		// 12. 记录导入操作
		for key, value := range mergedData {
			record := core.MigrationRecord{
				StepName:   fmt.Sprintf("导入配置项 %s", key),
				ActionType: constants.ActionTypeImport,
				Key:        key,
				AfterValue: fmt.Sprintf("%v", value),
				Status:     constants.RecordStatusSuccess,
				Timestamp:  time.Now(),
			}

			if oldValue, exists := targetData[key]; exists {
				record.BeforeValue = fmt.Sprintf("%v", oldValue)
			}

			result.Records = append(result.Records, record)
			result.Summary.Total++
			result.Summary.Success++
		}
	}

	result.Status = constants.TaskStatusCompleted
	result.Message = fmt.Sprintf("成功导入 %d 个配置项到: %s", result.Summary.Success, targetPath)

	return result, nil
}

// ValidateExport 验证导出配置
func (s *ConfigFileStrategy) ValidateExport(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Source.Path == "" {
		return fmt.Errorf("source path is required for export")
	}

	if _, err := os.Stat(config.Source.Path); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", config.Source.Path)
	}

	return nil
}

// ValidateImport 验证导入配置
func (s *ConfigFileStrategy) ValidateImport(config *core.MigrationConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	importPath := config.Options.ImportPath
	if importPath == "" && config.Source.Path == "" {
		return fmt.Errorf("import path is required")
	}

	if importPath == "" {
		importPath = config.Source.Path
	}

	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		return fmt.Errorf("import file does not exist: %s", importPath)
	}

	return nil
}

// detectFormat 检测文件格式
func (s *ConfigFileStrategy) detectFormat(path, format string) string {
	if format != "" {
		return format
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".ini":
		return "ini"
	case ".toml":
		return "toml"
	case ".xml":
		return "xml"
	default:
		return "unknown"
	}
}

// calculateChecksum 计算内容校验和
func (s *ConfigFileStrategy) calculateChecksum(data map[string]interface{}) string {
	content, _ := json.Marshal(data)
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// validateExportPackage 验证导出包
func (s *ConfigFileStrategy) validateExportPackage(pkg *core.ExportPackage) error {
	if pkg.Metadata.Version == "" {
		return fmt.Errorf("export package version is missing")
	}
	if pkg.Content.Data == nil {
		return fmt.Errorf("export package content is empty")
	}
	return nil
}

// generateExportID 生成导出ID
func generateExportID() string {
	return fmt.Sprintf("export_%d", time.Now().UnixNano())
}
