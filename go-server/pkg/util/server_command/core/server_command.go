// Package core 提供批处理文件和命令执行功能
package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ========== 常量定义 ==========

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

// ========== 模型定义 ==========

// ExecuteRequest 执行请求
type ExecuteRequest struct {
	// 命令类型: batch 或 command
	Type string `json:"type"`

	// 命令内容或批处理文件路径
	Command string `json:"command"`

	// 命令参数
	Args []string `json:"args,omitempty"`

	// 工作目录
	WorkDir string `json:"work_dir,omitempty"`

	// 环境变量
	Env map[string]string `json:"env,omitempty"`

	// 超时时间
	Timeout time.Duration `json:"timeout,omitempty"`

	// 是否捕获输出
	CaptureOutput bool `json:"capture_output"`

	// 是否实时输出
	StreamOutput bool `json:"stream_output"`
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
	// 执行ID
	ID string `json:"id"`

	// 执行状态
	Status string `json:"status"`

	// 退出码
	ExitCode int `json:"exit_code"`

	// 标准输出
	Stdout string `json:"stdout,omitempty"`

	// 标准错误
	Stderr string `json:"stderr,omitempty"`

	// 开始时间
	StartTime time.Time `json:"start_time"`

	// 结束时间
	EndTime time.Time `json:"end_time"`

	// 执行耗时
	Duration time.Duration `json:"duration"`

	// 错误信息
	Error string `json:"error,omitempty"`
}

// OutputHandler 输出处理器
type OutputHandler interface {
	OnOutput(line string, isStderr bool)
	OnComplete(response *ExecuteResponse)
	OnError(err error)
}

// ExecuteOptions 执行选项
type ExecuteOptions struct {
	// 上下文
	Context context.Context

	// 输出处理器
	OutputHandler OutputHandler

	// 是否异步执行
	Async bool
}

// Execution 执行实例
type Execution struct {
	ID       string
	Request  *ExecuteRequest
	Response *ExecuteResponse
	Process  *os.Process
	Cancel   context.CancelFunc
	mu       sync.RWMutex
}

// ========== 核心实现 ==========

// Executor 命令执行器
type Executor struct {
	executions sync.Map // map[string]*Execution
	idCounter  int64
	mu         sync.Mutex
}

// Execute 执行命令
func (e *Executor) Execute(req *ExecuteRequest, opts *ExecuteOptions) (*ExecuteResponse, error) {
	if err := e.validateRequest(req); err != nil {
		return nil, err
	}

	// 创建执行实例
	execution := e.createExecution(req)

	// 设置默认选项
	if opts == nil {
		opts = &ExecuteOptions{}
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// 异步执行
	if opts.Async {
		go e.executeAsync(execution, opts)
		return execution.Response, nil
	}

	// 同步执行
	return e.executeSync(execution, opts)
}

// ExecuteBatch 执行批处理文件
func (e *Executor) ExecuteBatch(filePath string, args []string, opts *ExecuteOptions) (*ExecuteResponse, error) {
	req := &ExecuteRequest{
		Type:          TypeBatch,
		Command:       filePath,
		Args:          args,
		CaptureOutput: true,
		Timeout:       DefaultTimeout,
	}

	return e.Execute(req, opts)
}

// ExecuteCommand 执行单个命令
func (e *Executor) ExecuteCommand(command string, args []string, opts *ExecuteOptions) (*ExecuteResponse, error) {
	req := &ExecuteRequest{
		Type:          TypeCommand,
		Command:       command,
		Args:          args,
		CaptureOutput: true,
		Timeout:       DefaultTimeout,
	}

	return e.Execute(req, opts)
}

// GetExecution 获取执行实例
func (e *Executor) GetExecution(id string) (*Execution, bool) {
	if val, ok := e.executions.Load(id); ok {
		return val.(*Execution), true
	}
	return nil, false
}

// ListExecutions 列出所有执行实例
func (e *Executor) ListExecutions() []*Execution {
	var executions []*Execution
	e.executions.Range(func(key, value interface{}) bool {
		executions = append(executions, value.(*Execution))
		return true
	})
	return executions
}

// CancelExecution 取消执行
func (e *Executor) CancelExecution(id string) error {
	execution, exists := e.GetExecution(id)
	if !exists {
		return fmt.Errorf("execution not found: %s", id)
	}

	execution.mu.Lock()
	defer execution.mu.Unlock()

	if execution.Cancel != nil {
		execution.Cancel()
	}

	if execution.Process != nil {
		if err := execution.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
	}

	execution.Response.Status = StatusCanceled
	execution.Response.EndTime = time.Now()
	execution.Response.Duration = execution.Response.EndTime.Sub(execution.Response.StartTime)

	return nil
}

// ========== 私有方法 ==========

// validateRequest 验证请求
func (e *Executor) validateRequest(req *ExecuteRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Command == "" {
		return ErrInvalidCommand
	}

	if req.Type != TypeBatch && req.Type != TypeCommand {
		return fmt.Errorf("invalid command type: %s", req.Type)
	}

	if req.Type == TypeBatch {
		if _, err := os.Stat(req.Command); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	}

	if req.Timeout <= 0 {
		req.Timeout = DefaultTimeout
	}

	return nil
}

// createExecution 创建执行实例
func (e *Executor) createExecution(req *ExecuteRequest) *Execution {
	e.mu.Lock()
	e.idCounter++
	id := fmt.Sprintf("exec_%d_%d", time.Now().Unix(), e.idCounter)
	e.mu.Unlock()

	execution := &Execution{
		ID:      id,
		Request: req,
		Response: &ExecuteResponse{
			ID:        id,
			Status:    StatusPending,
			StartTime: time.Now(),
		},
	}

	e.executions.Store(id, execution)
	return execution
}

// executeSync 同步执行
func (e *Executor) executeSync(execution *Execution, opts *ExecuteOptions) (*ExecuteResponse, error) {
	ctx, cancel := context.WithTimeout(opts.Context, execution.Request.Timeout)
	execution.Cancel = cancel
	defer cancel()

	return e.doExecute(ctx, execution, opts)
}

// executeAsync 异步执行
func (e *Executor) executeAsync(execution *Execution, opts *ExecuteOptions) {
	ctx, cancel := context.WithTimeout(opts.Context, execution.Request.Timeout)
	execution.Cancel = cancel
	defer cancel()

	response, err := e.doExecute(ctx, execution, opts)

	if opts.OutputHandler != nil {
		if err != nil {
			opts.OutputHandler.OnError(err)
		} else {
			opts.OutputHandler.OnComplete(response)
		}
	}
}

// doExecute 执行命令
func (e *Executor) doExecute(ctx context.Context, execution *Execution, opts *ExecuteOptions) (*ExecuteResponse, error) {
	req := execution.Request
	resp := execution.Response

	// 更新状态
	resp.Status = StatusRunning

	// 准备命令
	var cmd *exec.Cmd
	if req.Type == TypeBatch {
		cmd = e.prepareBatchCommand(req)
	} else {
		cmd = e.prepareCommand(req)
	}

	cmd.Dir = req.WorkDir
	if req.Env != nil {
		cmd.Env = append(os.Environ(), e.mapToEnvSlice(req.Env)...)
	}

	// 设置上下文
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

	// 处理输出
	var stdout, stderr strings.Builder

	if req.CaptureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if req.StreamOutput && opts.OutputHandler != nil {
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()

		go e.streamOutput(stdoutPipe, opts.OutputHandler, false)
		go e.streamOutput(stderrPipe, opts.OutputHandler, true)
	}

	// 执行命令
	err := cmd.Start()
	if err != nil {
		resp.Status = StatusFailed
		resp.Error = err.Error()
		resp.EndTime = time.Now()
		resp.Duration = resp.EndTime.Sub(resp.StartTime)
		return resp, err
	}

	// 保存进程信息
	execution.Process = cmd.Process

	// 等待完成
	err = cmd.Wait()

	// 更新响应
	resp.EndTime = time.Now()
	resp.Duration = resp.EndTime.Sub(resp.StartTime)

	if ctx.Err() == context.DeadlineExceeded {
		resp.Status = StatusCanceled
		resp.Error = ErrExecutionTimeout.Error()
		return resp, ErrExecutionTimeout
	}

	if ctx.Err() == context.Canceled {
		resp.Status = StatusCanceled
		resp.Error = ErrExecutionCanceled.Error()
		return resp, ErrExecutionCanceled
	}

	if err != nil {
		resp.Status = StatusFailed
		resp.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitError.ExitCode()
		}
	} else {
		resp.Status = StatusCompleted
		resp.ExitCode = 0
	}

	if req.CaptureOutput {
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
	}

	return resp, err
}

// prepareBatchCommand 准备批处理命令
func (e *Executor) prepareBatchCommand(req *ExecuteRequest) *exec.Cmd {
	switch runtime.GOOS {
	case WindowsPlatform:
		args := []string{"/C", req.Command}
		args = append(args, req.Args...)
		return exec.Command("cmd", args...)
	default:
		// Linux/macOS
		if filepath.Ext(req.Command) == ".sh" {
			args := []string{req.Command}
			args = append(args, req.Args...)
			return exec.Command("bash", args...)
		}
		return exec.Command(req.Command, req.Args...)
	}
}

// prepareCommand 准备普通命令
func (e *Executor) prepareCommand(req *ExecuteRequest) *exec.Cmd {
	return exec.Command(req.Command, req.Args...)
}

// streamOutput 流式输出处理
func (e *Executor) streamOutput(pipe io.ReadCloser, handler OutputHandler, isStderr bool) {
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		handler.OnOutput(scanner.Text(), isStderr)
	}
}

// mapToEnvSlice 将map转换为环境变量切片
func (e *Executor) mapToEnvSlice(env map[string]string) []string {
	var result []string
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// ========== 辅助方法 ==========

// IsWindows 判断是否为Windows平台
func IsWindows() bool {
	return runtime.GOOS == WindowsPlatform
}

// GetShellCommand 获取Shell命令
func GetShellCommand() (string, []string) {
	if IsWindows() {
		return "cmd", []string{"/C"}
	}
	return "bash", []string{"-c"}
}

// ValidateBatchFile 验证批处理文件
func ValidateBatchFile(filePath string) error {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return ErrFileNotFound
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// 检查文件权限
	file, err := os.Open(filePath)
	if err != nil {
		return ErrPermissionDenied
	}
	file.Close()

	return nil
}

// NewExecutor 创建新的执行器
func NewExecutor() *Executor {
	return &Executor{}
}
