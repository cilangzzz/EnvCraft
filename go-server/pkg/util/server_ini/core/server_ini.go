package core

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PackageInfo 压缩包信息结构体
type PackageInfo struct {
	Path         string      `gorm:"type:varchar(512);not null;comment:压缩包路径（用户传入的原始路径）"`
	FullPath     string      `gorm:"type:varchar(512);not null;uniqueIndex;comment:压缩包完整绝对路径"`
	Name         string      `gorm:"type:varchar(256);not null;comment:压缩包文件名（不含路径）"`
	FileCount    int         `gorm:"type:int;not null;default:0;comment:包含的文件数量"`
	TotalSize    int64       `gorm:"type:bigint;not null;default:0;comment:解压后总大小"`
	FileInfos    []*FileInfo `gorm:"foreignKey:PackageID;references:ID;comment:文件列表"`
	ModifiedTime time.Time   `gorm:"type:datetime;not null;comment:压缩包修改时间"`
}

// FileInfo 文件详细信息结构体（扩展 fs.FileInfo）
type FileInfo struct {
	PackageID    uint   `gorm:"type:int;not null;index;comment:所属压缩包ID"`
	RelativePath string `gorm:"type:varchar(512);not null;comment:文件在压缩包中的相对路径"`
	IsCompressed bool   `gorm:"type:tinyint(1);not null;default:0;comment:是否被压缩"`
	CRC32        uint32 `gorm:"type:int unsigned;not null;default:0;comment:CRC32校验值"`
	Method       uint16 `gorm:"type:smallint unsigned;not null;default:0;comment:压缩方法"`
	fs.FileInfo         // 嵌入官方 FileInfo
}

// ExtractOptions 解压选项
type ExtractOptions struct {
	TargetDir    string   // 目标目录
	Overwrite    bool     // 是否覆盖已存在文件
	PreservePerm bool     // 是否保留文件权限
	Exclude      []string // 排除的文件模式
}

// OpenPackage 打开并解析压缩包（更新后的实现）
func OpenPackage(path string) (*PackageInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("package file does not exist: %s", path)
	}

	// 获取完整绝对路径
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 获取文件名
	fileName := filepath.Base(path)

	// 打开zip文件
	r, err := zip.OpenReader(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// 收集文件信息
	var totalSize int64
	files := make([]*FileInfo, 0, len(r.File))

	for _, f := range r.File {
		// 跳过目录
		if f.FileInfo().IsDir() {
			continue
		}

		totalSize += int64(f.UncompressedSize64)

		files = append(files, &FileInfo{
			FileInfo:     f.FileInfo(),
			RelativePath: f.Name,
			IsCompressed: f.Method != zip.Store,
			CRC32:        f.CRC32,
			Method:       f.Method,
		})
	}

	// 获取压缩包修改时间
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get package stats: %w", err)
	}

	return &PackageInfo{
		Path:         path,
		FullPath:     fullPath,
		Name:         fileName,
		FileCount:    len(files),
		TotalSize:    totalSize,
		FileInfos:    files,
		ModifiedTime: fileInfo.ModTime(),
	}, nil
}

// Extract 解压压缩包
func (p *PackageInfo) Extract(options ExtractOptions) error {
	// 验证目标目录
	if options.TargetDir == "" {
		return errors.New("target directory cannot be empty")
	}

	// 创建目标目录
	if err := os.MkdirAll(options.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 重新打开zip文件
	r, err := zip.OpenReader(p.Path)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// 解压每个文件
	for _, f := range r.File {
		// 检查是否在排除列表中
		if isExcluded(f.Name, options.Exclude) {
			continue
		}

		destPath := filepath.Join(options.TargetDir, f.Name)

		// 检查ZipSlip漏洞
		if !strings.HasPrefix(destPath, filepath.Clean(options.TargetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// 检查文件是否已存在
		if !options.Overwrite {
			if _, err := os.Stat(destPath); err == nil {
				continue // 文件已存在且不覆盖
			}
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// 打开目标文件
		flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !options.Overwrite {
			flags |= os.O_EXCL
		}

		outFile, err := os.OpenFile(destPath, flags, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// 打开源文件
		inFile, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		// 复制文件内容
		if _, err := io.Copy(outFile, inFile); err != nil {
			outFile.Close()
			inFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		// 关闭文件
		outFile.Close()
		inFile.Close()

		// 设置文件权限（如果需要）
		if options.PreservePerm {
			if err := os.Chmod(destPath, f.Mode()); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		}
	}

	return nil
}

// isExcluded 检查文件是否在排除列表中
func isExcluded(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue // 忽略无效模式
		}
		if matched {
			return true
		}
	}
	return false
}

// PrintSummary 更新后的打印方法
func (p *PackageInfo) PrintSummary() {
	fmt.Printf("Package Path: %s\n", p.Path)
	fmt.Printf("Full Path: %s\n", p.FullPath)
	fmt.Printf("Package Name: %s\n", p.Name)
	fmt.Printf("File Count: %d\n", p.FileCount)
	fmt.Printf("Total Size: %d bytes\n", p.TotalSize)
	fmt.Printf("Modified Time: %s\n", p.ModifiedTime.Format(time.RFC3339))
	fmt.Println("FileInfos:")
	for _, file := range p.FileInfos {
		fmt.Printf("  %s (%d bytes, compressed: %v)\n",
			file.RelativePath, file.Size(), file.IsCompressed)
	}
}
