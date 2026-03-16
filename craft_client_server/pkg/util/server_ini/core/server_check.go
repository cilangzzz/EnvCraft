package core

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

// HashType 支持的哈希算法类型
type HashType string

const (
	MD5    HashType = "md5"
	SHA1   HashType = "sha1"
	SHA256 HashType = "sha256"
	SHA512 HashType = "sha512"
)

// FileCheckResult 文件校验结果
type FileCheckResult struct {
	Exists    bool   `json:"exists"`     // 文件是否存在
	IsFile    bool   `json:"is_file"`    // 是否是文件(不是目录)
	Size      int64  `json:"size"`       // 文件大小(字节)
	HashType  string `json:"hash_type"`  // 使用的哈希算法
	HashValue string `json:"hash_value"` // 哈希值
	Error     string `json:"error"`      // 错误信息
}

// CheckFileExists 检查文件是否存在
func CheckFileExists(filePath string) FileCheckResult {
	result := FileCheckResult{
		Exists: false,
		IsFile: false,
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Error = "file does not exist"
		} else {
			result.Error = err.Error()
		}
		return result
	}

	result.Exists = true
	result.IsFile = !fileInfo.IsDir()
	result.Size = fileInfo.Size()
	return result
}

// CalculateFileHash 计算文件哈希值
func CalculateFileHash(filePath string, hashType HashType) FileCheckResult {
	result := CheckFileExists(filePath)
	if !result.Exists || !result.IsFile {
		return result
	}

	file, err := os.Open(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to open file: %v", err)
		return result
	}
	defer file.Close()

	var hasher hash.Hash
	switch hashType {
	case MD5:
		hasher = md5.New()
	case SHA1:
		hasher = sha1.New()
	case SHA256:
		hasher = sha256.New()
	case SHA512:
		hasher = sha512.New()
	default:
		result.Error = "unsupported hash type"
		return result
	}

	if _, err := io.Copy(hasher, file); err != nil {
		result.Error = fmt.Sprintf("failed to calculate hash: %v", err)
		return result
	}

	result.HashType = string(hashType)
	result.HashValue = hex.EncodeToString(hasher.Sum(nil))
	return result
}

// VerifyFileHash 验证文件哈希值
func VerifyFileHash(filePath string, hashType HashType, expectedHash string) (bool, error) {
	result := CalculateFileHash(filePath, hashType)
	if result.Error != "" {
		return false, errors.New(result.Error)
	}

	return result.HashValue == expectedHash, nil
}
