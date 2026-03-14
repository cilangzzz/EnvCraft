package migration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tsc/pkg/util/migration"
	"tsc/pkg/util/migration/core"
)

const (
	// IDEA 配置源目录
	ideaSourceDir = `C:\Users\sysadmin\AppData\Roaming\JetBrains\IntelliJIdea2024.1\settingsSync`
	// 导出目标目录
	exportTargetDir = `H:\basePlatform\testData\idea-config-export`
	// 导入目标目录
	importTargetDir = `H:\basePlatform\testData\idea-config-import`
)

// TestIDEAConfigExportImport 测试 IDEA 配置文件的导出和导入
func TestIDEAConfigExportImport(t *testing.T) {
	// 确保目标目录存在
	if err := os.MkdirAll(exportTargetDir, 0755); err != nil {
		t.Fatalf("创建导出目录失败: %v", err)
	}
	if err := os.MkdirAll(importTargetDir, 0755); err != nil {
		t.Fatalf("创建导入目录失败: %v", err)
	}

	// 1. 导出 IDEA 配置文件
	t.Run("ExportIDEAConfig", func(t *testing.T) {
		exportIDEAConfig(t)
	})

	// 2. 导入 IDEA 配置文件
	t.Run("ImportIDEAConfig", func(t *testing.T) {
		importIDEAConfig(t)
	})
}

// exportIDEAConfig 导出 IDEA 配置文件
func exportIDEAConfig(t *testing.T) {
	// 遍历 IDEA 配置目录
	subDirs := []string{"options", "codestyles", "colors", "keymaps", "inspection", "fileTemplates"}

	for _, subDir := range subDirs {
		subPath := filepath.Join(ideaSourceDir, subDir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			t.Logf("跳过不存在的目录: %s", subPath)
			continue
		}

		// 创建导出子目录
		exportSubDir := filepath.Join(exportTargetDir, subDir)
		if err := os.MkdirAll(exportSubDir, 0755); err != nil {
			t.Errorf("创建导出子目录失败: %v", err)
			continue
		}

		// 遍历目录中的文件
		files, err := os.ReadDir(subPath)
		if err != nil {
			t.Errorf("读取目录失败: %v", err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filePath := filepath.Join(subPath, file.Name())
			t.Logf("导出文件: %s", filePath)

			// 创建迁移配置
			config := migration.NewConfig()
			config.TaskID = fmt.Sprintf("export_%s_%d", file.Name(), time.Now().Unix())
			config.Name = fmt.Sprintf("导出 IDEA 配置文件: %s", file.Name())
			config.Type = migration.MigrationType.ConfigFile

			// 设置源配置
			config.Source.Path = filePath
			config.Source.Format = detectFileFormat(filePath)
			config.Source.Encoding = "utf-8"

			// 设置导出选项
			config.Options.OperationMode = "export"
			config.Options.ExportPath = filepath.Join(exportSubDir, file.Name()+".export.json")
			config.Options.IncludeRawContent = true // 包含原始内容，确保 XML 文件能完整导出
			config.Options.Verbose = true

			// 获取策略并执行导出
			strategy, err := migration.GetStrategy(migration.MigrationType.ConfigFile)
			if err != nil {
				t.Errorf("获取迁移策略失败: %v", err)
				continue
			}

			// 验证导出配置
			if err := strategy.ValidateExport(config); err != nil {
				t.Errorf("验证导出配置失败: %v", err)
				continue
			}

			// 执行导出
			result, err := strategy.Export(context.Background(), config)
			if err != nil {
				t.Errorf("导出失败: %v", err)
				continue
			}

			t.Logf("导出结果: 状态=%s, 消息=%s, 路径=%s", result.Status, result.Message, result.ExportPath)
		}
	}

	// 导出目录结构信息
	exportStructure(t)
}

// importIDEAConfig 导入 IDEA 配置文件
func importIDEAConfig(t *testing.T) {
	// 遍历导出目录
	subDirs := []string{"options", "codestyles", "colors", "keymaps", "inspection", "fileTemplates"}

	for _, subDir := range subDirs {
		exportSubDir := filepath.Join(exportTargetDir, subDir)
		if _, err := os.Stat(exportSubDir); os.IsNotExist(err) {
			t.Logf("跳过不存在的导出目录: %s", exportSubDir)
			continue
		}

		// 创建导入子目录
		importSubDir := filepath.Join(importTargetDir, subDir)
		if err := os.MkdirAll(importSubDir, 0755); err != nil {
			t.Errorf("创建导入子目录失败: %v", err)
			continue
		}

		// 遍历导出的文件
		files, err := os.ReadDir(exportSubDir)
		if err != nil {
			t.Errorf("读取导出目录失败: %v", err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// 只处理导出文件
			if filepath.Ext(file.Name()) != ".json" {
				continue
			}

			exportFilePath := filepath.Join(exportSubDir, file.Name())
			t.Logf("导入文件: %s", exportFilePath)

			// 从导出包中获取原始文件名
			originalName, err := getOriginalFileName(exportFilePath)
			if err != nil {
				t.Errorf("获取原始文件名失败: %v", err)
				continue
			}

			// 创建迁移配置
			config := migration.NewConfig()
			config.TaskID = fmt.Sprintf("import_%s_%d", originalName, time.Now().Unix())
			config.Name = fmt.Sprintf("导入 IDEA 配置文件: %s", originalName)
			config.Type = migration.MigrationType.ConfigFile

			// 设置导入选项
			config.Options.OperationMode = "import"
			config.Options.ImportPath = exportFilePath
			config.Options.PreserveFormat = true // 保持原始格式

			// 设置目标配置
			config.Target.Path = filepath.Join(importSubDir, originalName)
			config.Target.Backup = true
			config.Target.MergeMode = "overwrite"
			config.Target.CreateIfNotExists = true

			// 获取策略并执行导入
			strategy, err := migration.GetStrategy(migration.MigrationType.ConfigFile)
			if err != nil {
				t.Errorf("获取迁移策略失败: %v", err)
				continue
			}

			// 验证导入配置
			if err := strategy.ValidateImport(config); err != nil {
				t.Errorf("验证导入配置失败: %v", err)
				continue
			}

			// 执行导入
			result, err := strategy.Import(context.Background(), config)
			if err != nil {
				t.Errorf("导入失败: %v", err)
				continue
			}

			t.Logf("导入结果: 状态=%s, 消息=%s, 成功=%d", result.Status, result.Message, result.Summary.Success)
		}
	}

	// 验证导入结果
	verifyImportResult(t)
}

// exportStructure 导出目录结构信息
func exportStructure(t *testing.T) {
	structurePath := filepath.Join(exportTargetDir, "structure.json")

	// 收集目录结构
	structure := make(map[string][]string)
	subDirs := []string{"options", "codestyles", "colors", "keymaps", "inspection", "fileTemplates"}

	for _, subDir := range subDirs {
		subPath := filepath.Join(ideaSourceDir, subDir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}

		var fileNames []string
		for _, file := range files {
			if !file.IsDir() {
				fileNames = append(fileNames, file.Name())
			}
		}
		structure[subDir] = fileNames
	}

	// 写入结构文件
	data, err := json.MarshalIndent(structure, "", "  ")
	if err != nil {
		t.Errorf("序列化目录结构失败: %v", err)
		return
	}

	if err := os.WriteFile(structurePath, data, 0644); err != nil {
		t.Errorf("写入目录结构文件失败: %v", err)
		return
	}

	t.Logf("目录结构已保存到: %s", structurePath)
}

// verifyImportResult 验证导入结果
func verifyImportResult(t *testing.T) {
	// 读取目录结构
	structurePath := filepath.Join(exportTargetDir, "structure.json")
	structureData, err := os.ReadFile(structurePath)
	if err != nil {
		t.Logf("读取目录结构文件失败: %v", err)
		return
	}

	var structure map[string][]string
	if err := json.Unmarshal(structureData, &structure); err != nil {
		t.Logf("解析目录结构失败: %v", err)
		return
	}

	// 验证每个文件
	totalFiles := 0
	successFiles := 0

	for subDir, files := range structure {
		for _, fileName := range files {
			totalFiles++
			importPath := filepath.Join(importTargetDir, subDir, fileName)
			if _, err := os.Stat(importPath); err == nil {
				successFiles++
			} else {
				t.Logf("缺失文件: %s", importPath)
			}
		}
	}

	t.Logf("导入验证: 总计=%d, 成功=%d, 失败=%d", totalFiles, successFiles, totalFiles-successFiles)
}

// detectFileFormat 检测文件格式
func detectFileFormat(path string) string {
	ext := filepath.Ext(path)
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

// getOriginalFileName 从导出包获取原始文件名
func getOriginalFileName(exportPath string) (string, error) {
	data, err := os.ReadFile(exportPath)
	if err != nil {
		return "", err
	}

	var exportPkg core.ExportPackage
	if err := json.Unmarshal(data, &exportPkg); err != nil {
		return "", err
	}

	return filepath.Base(exportPkg.Metadata.OriginalPath), nil
}