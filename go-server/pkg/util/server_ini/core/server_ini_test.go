package core

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 测试用的临时文件和目录
var (
	testZipFile    = "testdata/test.zip"
	testOutputDir  = "testdata/output"
	testSourceFile = "testdata/source.txt"
)

// 初始化测试环境
func setup() error {
	// 创建测试目录
	if err := os.MkdirAll("testdata", 0755); err != nil {
		return err
	}

	// 创建测试源文件
	if err := os.WriteFile(testSourceFile, []byte("This is a test file"), 0644); err != nil {
		return err
	}

	// 创建测试zip文件
	if err := createTestZip(); err != nil {
		return err
	}

	return nil
}

// 清理测试环境
func teardown() {
	os.RemoveAll(testOutputDir)
}

// 创建测试用的zip文件
func createTestZip() error {
	file, err := os.Create(testZipFile)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 添加文件到zip
	srcFile, err := os.Open(testSourceFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(testSourceFile)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, srcFile)
	return err
}

// TestOpenPackage 测试打开压缩包
func TestOpenPackage(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer teardown()

	pkg, err := OpenPackage(testZipFile)
	if err != nil {
		t.Fatalf("OpenPackage failed: %v", err)
	}

	if pkg.Path != testZipFile {
		t.Errorf("Expected path %s, got %s", testZipFile, pkg.Path)
	}

	if pkg.FileCount != 1 {
		t.Errorf("Expected 1 file, got %d", pkg.FileCount)
	}

	if len(pkg.FileInfos) != 1 {
		t.Errorf("Expected 1 file in list, got %d", len(pkg.FileInfos))
	}

	fileInfo := pkg.FileInfos[0]
	if fileInfo.Name() != filepath.Base(testSourceFile) {
		t.Errorf("Expected file name %s, got %s", filepath.Base(testSourceFile), fileInfo.Name())
	}

	if fileInfo.Size() != 18 {
		t.Errorf("Expected file size 18, got %d", fileInfo.Size())
	}
}

// TestExtract 测试解压功能
func TestExtract(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer teardown()

	pkg, err := OpenPackage(testZipFile)
	if err != nil {
		t.Fatalf("OpenPackage failed: %v", err)
	}

	options := ExtractOptions{
		TargetDir:    testOutputDir,
		Overwrite:    true,
		PreservePerm: true,
	}

	if err := pkg.Extract(options); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// 验证解压的文件
	extractedFile := filepath.Join(testOutputDir, filepath.Base(testSourceFile))
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Errorf("Extracted file not found: %s", extractedFile)
	}

	// 验证文件内容
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Errorf("Failed to read extracted file: %v", err)
	}

	if string(content) != "This is a test file" {
		t.Errorf("File content mismatch, expected 'This is a test file', got '%s'", string(content))
	}
}

// TestExtractWithExclude 测试带排除模式的解压
func TestExtractWithExclude(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer teardown()

	// 添加第二个文件到zip
	secondFile := "testdata/excluded.txt"
	if err := os.WriteFile(secondFile, []byte("This should be excluded"), 0644); err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}
	defer os.Remove(secondFile)

	// 重新创建zip包含两个文件
	if err := createTestZip(); err != nil {
		t.Fatalf("Failed to recreate test zip: %v", err)
	}

	pkg, err := OpenPackage(testZipFile)
	if err != nil {
		t.Fatalf("OpenPackage failed: %v", err)
	}

	options := ExtractOptions{
		TargetDir:    testOutputDir,
		Overwrite:    true,
		PreservePerm: true,
		Exclude:      []string{"*.txt"},
	}

	if err := pkg.Extract(options); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// 验证文件是否被排除
	extractedFile := filepath.Join(testOutputDir, filepath.Base(testSourceFile))
	if _, err := os.Stat(extractedFile); !os.IsNotExist(err) {
		t.Errorf("File should be excluded but was extracted: %s", extractedFile)
	}
}

// TestInvalidPackage 测试无效压缩包
func TestInvalidPackage(t *testing.T) {
	invalidFile := "testdata/invalid.zip"
	if err := os.WriteFile(invalidFile, []byte("not a zip file"), 0644); err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}
	defer os.Remove(invalidFile)

	_, err := OpenPackage(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid zip file, got nil")
	}
}

// TestZipSlipProtection 测试ZipSlip防护
func TestZipSlipProtection(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer teardown()

	// 创建恶意的zip文件
	maliciousZip := "testdata/malicious.zip"
	file, err := os.Create(maliciousZip)
	if err != nil {
		t.Fatalf("Failed to create malicious zip: %v", err)
	}
	defer file.Close()
	defer os.Remove(maliciousZip)

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 添加恶意路径的文件
	header := &zip.FileHeader{
		Name: "../malicious.txt",
	}
	header.SetMode(0644)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		t.Fatalf("Failed to add malicious file to zip: %v", err)
	}

	if _, err := writer.Write([]byte("malicious content")); err != nil {
		t.Fatalf("Failed to write malicious content: %v", err)
	}

	zipWriter.Close()

	// 尝试解压
	pkg, err := OpenPackage(maliciousZip)
	if err != nil {
		t.Fatalf("OpenPackage failed: %v", err)
	}

	options := ExtractOptions{
		TargetDir: testOutputDir,
	}

	err = pkg.Extract(options)
	if err == nil {
		t.Fatal("Expected error for ZipSlip attack, got nil")
	}

	if !strings.Contains(err.Error(), "illegal file path") {
		t.Errorf("Expected ZipSlip protection error, got: %v", err)
	}
}

func TestPackageInfoFields(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer teardown()

	// 使用相对路径测试
	relPath := "testdata/test.zip"
	pkg, err := OpenPackage(relPath)
	if err != nil {
		t.Fatalf("OpenPackage failed: %v", err)
	}

	// 验证Path字段
	if pkg.Path != relPath {
		t.Errorf("Expected Path %s, got %s", relPath, pkg.Path)
	}

	// 验证FullPath字段
	absPath, _ := filepath.Abs(relPath)
	if pkg.FullPath != absPath {
		t.Errorf("Expected FullPath %s, got %s", absPath, pkg.FullPath)
	}

	// 验证Name字段
	expectedName := "test.zip"
	if pkg.Name != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, pkg.Name)
	}
}
