package core

import (
	"os"
	"testing"
)

func TestCheckFileExists(t *testing.T) {
	// 创建测试文件
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 写入一些内容
	if _, err := tempFile.WriteString("test content"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// 测试存在的文件
	result := CheckFileExists(tempFile.Name())
	if !result.Exists {
		t.Error("File should exist")
	}
	if !result.IsFile {
		t.Error("Should be a file, not directory")
	}
	if result.Size != 12 {
		t.Errorf("Expected size 12, got %d", result.Size)
	}

	// 测试不存在的文件
	result = CheckFileExists("nonexistent_file")
	if result.Exists {
		t.Error("File should not exist")
	}
}

func TestCalculateFileHash(t *testing.T) {
	// 创建测试文件
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 写入固定内容
	if _, err := tempFile.WriteString("test content"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// 测试各种哈希算法
	tests := []struct {
		name     string
		hashType HashType
		expected string
	}{
		{"MD5", MD5, "6f8db599de986fab7a21625b7916589c"},
		{"SHA1", SHA1, "4e1243bd22c66e76c2ba9eddc1f91394e57f9f83"},
		{"SHA256", SHA256, "6f8db599de986fab7a21625b7916589c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFileHash(tempFile.Name(), tt.hashType)
			if result.Error != "" {
				t.Errorf("Unexpected error: %v", result.Error)
			}
			if result.HashValue != tt.expected {
				t.Errorf("Expected hash %s, got %s", tt.expected, result.HashValue)
			}
		})
	}
}

func TestVerifyFileHash(t *testing.T) {
	// 创建测试文件
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 写入固定内容
	if _, err := tempFile.WriteString("test content"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// 测试验证
	match, err := VerifyFileHash(tempFile.Name(), SHA256, "6f8db599de986fab7a21625b7916589c")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !match {
		t.Error("Hash should match")
	}

	// 测试不匹配的情况
	match, err = VerifyFileHash(tempFile.Name(), SHA256, "wronghash")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if match {
		t.Error("Hash should not match")
	}
}
