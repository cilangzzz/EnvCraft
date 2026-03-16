package constants

import (
	"fmt"
	"time"
)

const (
	// StatusPending 执行状态
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"

	// TypeBatch 命令类型
	TypeBatch   = "batch"
	TypeCommand = "command"

	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second

	// DefaultBufferSize 缓冲区大小
	DefaultBufferSize = 1024

	// WindowsPlatform 平台相关
	WindowsPlatform = "windows"
	LinuxPlatform   = "linux"
	MacOSPlatform   = "darwin"
)

// 错误常量
var (
	ErrInvalidCommand    = fmt.Errorf("invalid command")
	ErrExecutionTimeout  = fmt.Errorf("execution timeout")
	ErrExecutionCanceled = fmt.Errorf("execution canceled")
	ErrFileNotFound      = fmt.Errorf("file not found")
	ErrPermissionDenied  = fmt.Errorf("permission denied")
)
