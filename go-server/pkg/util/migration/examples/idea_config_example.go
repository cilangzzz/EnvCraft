package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func main() {
	fmt.Println("=== IDEA 配置文件导出/导入测试 ===")
	fmt.Printf("源目录: %s\n", ideaSourceDir)
	fmt.Printf("导出目录: %s\n", exportTargetDir)
	fmt.Printf("导入目录: %s\n", importTargetDir)
	fmt.Println()

	// 确保目标目录存在
	if err := os.MkdirAll(exportTargetDir, 0755); err != nil {
		fmt.Printf("创建导出目录失败: %v\n", err)
		return
	}
	if err := os.MkdirAll(importTargetDir, 0755); err != nil {
		fmt.Printf("创建导入目录失败: %v\n", err)
		return
	}

	// 1. 导出 IDEA 配置文件
	fmt.Println("=== 开始导出 ===")
	exportCount := exportIDEAConfig()
	fmt.Printf("导出完成，共导出 %d 个文件\n\n", exportCount)

	// 2. 导入 IDEA 配置文件
	fmt.Println("=== 开始导入 ===")
	importCount := importIDEAConfig()
	fmt.Printf("导入完成，共导入 %d 个文件\n\n", importCount)

	// 3. 验证结果
	fmt.Println("=== 验证结果 ===")
	verifyResult()

	fmt.Println("=== 测试完成 ===")
}

// exportIDEAConfig 导出 IDEA 配置文件
func exportIDEAConfig() int {
	count := 0
	subDirs := []string{"options", "codestyles", "colors", "keymaps", "inspection", "fileTemplates"}

	for _, subDir := range subDirs {
		subPath := filepath.Join(ideaSourceDir, subDir)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			fmt.Printf("跳过不存在的目录: %s\n", subPath)
			continue
		}

		// 创建导出子目录
		exportSubDir := filepath.Join(exportTargetDir, subDir)
		if err := os.MkdirAll(exportSubDir, 0755); err != nil {
			fmt.Printf("创建导出子目录失败: %v\n", err)
			continue
		}

		// 遍历目录中的文件
		files, err := os.ReadDir(subPath)
		if err != nil {
			fmt.Printf("读取目录失败: %v\n", err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filePath := filepath.Join(subPath, file.Name())
			fmt.Printf("导出文件: %s\n", filePath)

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
			config.Options.IncludeRawContent = true // 包含原始内容
			config.Options.Verbose = true

			// 获取策略并执行导出
			strategy, err := migration.GetStrategy(migration.MigrationType.ConfigFile)
			if err != nil {
				fmt.Printf("获取迁移策略失败: %v\n", err)
				continue
			}

			// 验证导出配置
			if err := strategy.ValidateExport(config); err != nil {
				fmt.Printf("验证导出配置失败: %v\n", err)
				continue
			}

			// 执行导出
			result, err := strategy.Export(context.Background(), config)
			if err != nil {
				fmt.Printf("导出失败: %v\n", err)
				continue
			}

			fmt.Printf("  -> 状态: %s, 路径: %s\n", result.Status, result.ExportPath)
			count++
		}
	}

	// 导出目录结构信息
	exportStructure()

	return count
}

// importIDEAConfig 导入 IDEA 配置文件
func importIDEAConfig() int {
	count := 0
	subDirs := []string{"options", "codestyles", "colors", "keymaps", "inspection", "fileTemplates"}

	for _, subDir := range subDirs {
		exportSubDir := filepath.Join(exportTargetDir, subDir)
		if _, err := os.Stat(exportSubDir); os.IsNotExist(err) {
			continue
		}

		// 创建导入子目录
		importSubDir := filepath.Join(importTargetDir, subDir)
		if err := os.MkdirAll(importSubDir, 0755); err != nil {
			fmt.Printf("创建导入子目录失败: %v\n", err)
			continue
		}

		// 遍历导出的文件
		files, err := os.ReadDir(exportSubDir)
		if err != nil {
			fmt.Printf("读取导出目录失败: %v\n", err)
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

			// 从导出包中获取原始文件名
			originalName, err := getOriginalFileName(exportFilePath)
			if err != nil {
				fmt.Printf("获取原始文件名失败: %v\n", err)
				continue
			}

			fmt.Printf("导入文件: %s -> %s\n", exportFilePath, originalName)

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
				fmt.Printf("获取迁移策略失败: %v\n", err)
				continue
			}

			// 验证导入配置
			if err := strategy.ValidateImport(config); err != nil {
				fmt.Printf("验证导入配置失败: %v\n", err)
				continue
			}

			// 执行导入
			result, err := strategy.Import(context.Background(), config)
			if err != nil {
				fmt.Printf("导入失败: %v\n", err)
				continue
			}

			fmt.Printf("  -> 状态: %s, 成功: %d\n", result.Status, result.Summary.Success)
			count++
		}
	}

	return count
}

// exportStructure 导出目录结构信息
func exportStructure() {
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
		fmt.Printf("序列化目录结构失败: %v\n", err)
		return
	}

	if err := os.WriteFile(structurePath, data, 0644); err != nil {
		fmt.Printf("写入目录结构文件失败: %v\n", err)
		return
	}

	fmt.Printf("目录结构已保存到: %s\n", structurePath)
}

// verifyResult 验证导入结果
func verifyResult() {
	// 读取目录结构
	structurePath := filepath.Join(exportTargetDir, "structure.json")
	structureData, err := os.ReadFile(structurePath)
	if err != nil {
		fmt.Printf("读取目录结构文件失败: %v\n", err)
		return
	}

	var structure map[string][]string
	if err := json.Unmarshal(structureData, &structure); err != nil {
		fmt.Printf("解析目录结构失败: %v\n", err)
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
				fmt.Printf("  [OK] %s\n", importPath)
			} else {
				fmt.Printf("  [缺失] %s\n", importPath)
			}
		}
	}

	fmt.Printf("\n验证结果: 总计=%d, 成功=%d, 失败=%d\n", totalFiles, successFiles, totalFiles-successFiles)
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
