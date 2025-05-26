package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// ========== 测试辅助工具 ==========

// TestOutputHandler 测试用的输出处理器
type TestOutputHandler struct {
	outputs   []string
	errors    []error
	completed []*ExecuteResponse
	mu        sync.Mutex
}

func NewTestOutputHandler() *TestOutputHandler {
	return &TestOutputHandler{
		outputs:   make([]string, 0),
		errors:    make([]error, 0),
		completed: make([]*ExecuteResponse, 0),
	}
}

func (h *TestOutputHandler) OnOutput(line string, isStderr bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	prefix := "[STDOUT]"
	if isStderr {
		prefix = "[STDERR]"
	}
	h.outputs = append(h.outputs, fmt.Sprintf("%s %s", prefix, line))
}

func (h *TestOutputHandler) OnComplete(response *ExecuteResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.completed = append(h.completed, response)
}

func (h *TestOutputHandler) OnError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errors = append(h.errors, err)
}

func (h *TestOutputHandler) GetOutputs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.outputs...)
}

func (h *TestOutputHandler) GetErrors() []error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]error(nil), h.errors...)
}

func (h *TestOutputHandler) GetCompleted() []*ExecuteResponse {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*ExecuteResponse(nil), h.completed...)
}

// createTestBatchFile 创建测试用的批处理文件
func createTestBatchFile(t *testing.T, content string) string {
	var ext string
	if IsWindows() {
		ext = ".bat"
	} else {
		ext = ".sh"
	}

	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("test_*%s", ext))
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	tmpFile.Close()

	// 为Unix系统设置执行权限
	if !IsWindows() {
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			t.Fatalf("Failed to set execute permission: %v", err)
		}
	}

	return tmpFile.Name()
}

// ========== 单元测试 ==========

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()
	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}
}

func TestValidateRequest(t *testing.T) {
	executor := NewExecutor()

	tests := []struct {
		name    string
		req     *ExecuteRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty command",
			req: &ExecuteRequest{
				Type:    TypeCommand,
				Command: "",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			req: &ExecuteRequest{
				Type:    "invalid",
				Command: "echo",
			},
			wantErr: true,
		},
		{
			name: "valid command request",
			req: &ExecuteRequest{
				Type:    TypeCommand,
				Command: "echo",
				Args:    []string{"hello"},
			},
			wantErr: false,
		},
		{
			name: "batch file not found",
			req: &ExecuteRequest{
				Type:    TypeBatch,
				Command: "/non/existent/file.bat",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteCommand_Success(t *testing.T) {
	executor := NewExecutor()

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "echo"
		args = []string{"hello world"}
	} else {
		cmd = "echo"
		args = []string{"hello world"}
	}

	response, err := executor.ExecuteCommand(cmd, args, nil)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	if response.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, response.Status)
	}

	if response.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", response.ExitCode)
	}

	if !strings.Contains(response.Stdout, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got: %s", response.Stdout)
	}
}

func TestExecuteCommand_Failure(t *testing.T) {
	executor := NewExecutor()

	response, err := executor.ExecuteCommand("nonexistentcommand", nil, nil)
	if err == nil {
		t.Fatal("Expected error for non-existent command")
	}

	if response.Status != StatusFailed {
		t.Errorf("Expected status %s, got %s", StatusFailed, response.Status)
	}
}

func TestExecuteBatch_Success(t *testing.T) {
	executor := NewExecutor()

	var content string
	if IsWindows() {
		content = "@echo off\necho Hello from batch\necho Second line"
	} else {
		content = "#!/bin/bash\necho 'Hello from batch'\necho 'Second line'"
	}

	batchFile := createTestBatchFile(t, content)
	defer os.Remove(batchFile)

	response, err := executor.ExecuteBatch(batchFile, nil, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if response.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, response.Status)
	}

	if response.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", response.ExitCode)
	}

	if !strings.Contains(response.Stdout, "Hello from batch") {
		t.Errorf("Expected output to contain 'Hello from batch', got: %s", response.Stdout)
	}
}

func TestExecuteBatch_WithArgs(t *testing.T) {
	executor := NewExecutor()

	var content string
	if IsWindows() {
		content = "@echo off\necho Arg1: %1\necho Arg2: %2"
	} else {
		content = "#!/bin/bash\necho \"Arg1: $1\"\necho \"Arg2: $2\""
	}

	batchFile := createTestBatchFile(t, content)
	defer os.Remove(batchFile)

	args := []string{"test1", "test2"}
	response, err := executor.ExecuteBatch(batchFile, args, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch with args failed: %v", err)
	}

	if response.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, response.Status)
	}

	if !strings.Contains(response.Stdout, "test1") || !strings.Contains(response.Stdout, "test2") {
		t.Errorf("Expected output to contain args, got: %s", response.Stdout)
	}
}

func TestExecuteAsync(t *testing.T) {
	executor := NewExecutor()
	handler := NewTestOutputHandler()

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "ping"
		args = []string{"127.0.0.1", "-n", "3"}
	} else {
		cmd = "ping"
		args = []string{"-c", "3", "127.0.0.1"}
	}

	opts := &ExecuteOptions{
		Async:         true,
		OutputHandler: handler,
	}

	response, err := executor.ExecuteCommand(cmd, args, opts)
	if err != nil {
		t.Fatalf("Async execute failed: %v", err)
	}

	// 异步执行应该立即返回
	if response.Status != StatusPending && response.Status != StatusRunning {
		t.Errorf("Expected status pending or running for async execution, got %s", response.Status)
	}

	// 等待执行完成
	time.Sleep(5 * time.Second)

	completed := handler.GetCompleted()
	if len(completed) == 0 {
		t.Fatal("Expected at least one completed execution")
	}

	finalResponse := completed[0]
	if finalResponse.Status != StatusCompleted {
		t.Errorf("Expected final status %s, got %s", StatusCompleted, finalResponse.Status)
	}
}

func TestExecuteWithTimeout(t *testing.T) {
	executor := NewExecutor()

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "ping"
		args = []string{"127.0.0.1", "-n", "10"}
	} else {
		cmd = "sleep"
		args = []string{"5"}
	}

	req := &ExecuteRequest{
		Type:          TypeCommand,
		Command:       cmd,
		Args:          args,
		Timeout:       1 * time.Second,
		CaptureOutput: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := &ExecuteOptions{
		Context: ctx,
	}

	response, err := executor.Execute(req, opts)
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if response.Status != StatusCanceled {
		t.Errorf("Expected status %s, got %s", StatusCanceled, response.Status)
	}
}

func TestCancelExecution(t *testing.T) {
	executor := NewExecutor()

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "ping"
		args = []string{"127.0.0.1", "-n", "100"}
	} else {
		cmd = "sleep"
		args = []string{"30"}
	}

	opts := &ExecuteOptions{
		Async: true,
	}

	response, err := executor.ExecuteCommand(cmd, args, opts)
	if err != nil {
		t.Fatalf("Failed to start async command: %v", err)
	}

	// 等待命令开始执行
	time.Sleep(500 * time.Millisecond)

	// 取消执行
	err = executor.CancelExecution(response.ID)
	if err != nil {
		t.Fatalf("Failed to cancel execution: %v", err)
	}

	// 验证执行已被取消
	execution, exists := executor.GetExecution(response.ID)
	if !exists {
		t.Fatal("Execution not found")
	}

	// 等待取消完成
	time.Sleep(1 * time.Second)

	if execution.Response.Status != StatusCanceled {
		t.Errorf("Expected status %s, got %s", StatusCanceled, execution.Response.Status)
	}
}

func TestExecuteWithWorkDir(t *testing.T) {
	executor := NewExecutor()

	tempDir, err := ioutil.TempDir("", "test_workdir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 在临时目录中创建一个文件
	testFile := filepath.Join(tempDir, "test.txt")
	err = ioutil.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "dir"
		args = []string{"/b"}
	} else {
		cmd = "ls"
		args = []string{"-1"}
	}

	req := &ExecuteRequest{
		Type:          TypeCommand,
		Command:       cmd,
		Args:          args,
		WorkDir:       tempDir,
		CaptureOutput: true,
	}

	response, err := executor.Execute(req, nil)
	if err != nil {
		t.Fatalf("Execute with work dir failed: %v", err)
	}

	if !strings.Contains(response.Stdout, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt', got: %s", response.Stdout)
	}
}

func TestExecuteWithEnv(t *testing.T) {
	executor := NewExecutor()

	var cmd string
	var args []string

	if IsWindows() {
		cmd = "echo"
		args = []string{"%TEST_VAR%"}
	} else {
		cmd = "echo"
		args = []string{"$TEST_VAR"}
	}

	req := &ExecuteRequest{
		Type:    TypeCommand,
		Command: cmd,
		Args:    args,
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
		CaptureOutput: true,
	}

	response, err := executor.Execute(req, nil)
	if err != nil {
		t.Fatalf("Execute with env failed: %v", err)
	}

	if !strings.Contains(response.Stdout, "test_value") {
		t.Errorf("Expected output to contain 'test_value', got: %s", response.Stdout)
	}
}

func TestListExecutions(t *testing.T) {
	executor := NewExecutor()

	// 执行几个命令
	for i := 0; i < 3; i++ {
		var cmd string
		if IsWindows() {
			cmd = "echo"
		} else {
			cmd = "echo"
		}

		_, err := executor.ExecuteCommand(cmd, []string{fmt.Sprintf("test%d", i)}, nil)
		if err != nil {
			t.Fatalf("Failed to execute command %d: %v", i, err)
		}
	}

	executions := executor.ListExecutions()
	if len(executions) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executions))
	}
}

// ========== 集成测试 ==========

func TestIntegration_ComplexBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	executor := NewExecutor()

	var content string
	if IsWindows() {
		content = `@echo off
echo Starting complex batch test
set /a result=10+20
echo Result: %result%
echo Current directory: %CD%
dir /b | find /c /v ""
echo Batch test completed`
	} else {
		content = `#!/bin/bash
echo "Starting complex batch test"
result=$((10+20))
echo "Result: $result"
echo "Current directory: $(pwd)"
ls -1 | wc -l
echo "Batch test completed"`
	}

	batchFile := createTestBatchFile(t, content)
	defer os.Remove(batchFile)

	handler := NewTestOutputHandler()
	opts := &ExecuteOptions{
		OutputHandler: handler,
	}

	response, err := executor.ExecuteBatch(batchFile, nil, opts)
	if err != nil {
		t.Fatalf("Complex batch integration test failed: %v", err)
	}

	if response.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, response.Status)
	}

	if !strings.Contains(response.Stdout, "30") {
		t.Errorf("Expected calculation result '30' in output, got: %s", response.Stdout)
	}

	if !strings.Contains(response.Stdout, "completed") {
		t.Errorf("Expected 'completed' in output, got: %s", response.Stdout)
	}
}

// ========== 基准测试 ==========

func BenchmarkExecuteSimpleCommand(b *testing.B) {
	executor := NewExecutor()

	var cmd string
	if IsWindows() {
		cmd = "echo"
	} else {
		cmd = "echo"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.ExecuteCommand(cmd, []string{"benchmark"}, nil)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkExecuteConcurrent(b *testing.B) {
	executor := NewExecutor()

	var cmd string
	if IsWindows() {
		cmd = "echo"
	} else {
		cmd = "echo"
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := executor.ExecuteCommand(cmd, []string{"concurrent"}, nil)
			if err != nil {
				b.Fatalf("Concurrent benchmark failed: %v", err)
			}
		}
	})
}

// ========== 工具函数测试 ==========

func TestIsWindows(t *testing.T) {
	result := IsWindows()
	expected := runtime.GOOS == "windows"
	if result != expected {
		t.Errorf("IsWindows() = %v, expected %v", result, expected)
	}
}

func TestGetShellCommand(t *testing.T) {
	cmd, args := GetShellCommand()

	if IsWindows() {
		if cmd != "cmd" {
			t.Errorf("Expected cmd 'cmd' on Windows, got %s", cmd)
		}
		if len(args) != 1 || args[0] != "/C" {
			t.Errorf("Expected args ['/C'] on Windows, got %v", args)
		}
	} else {
		if cmd != "bash" {
			t.Errorf("Expected cmd 'bash' on Unix, got %s", cmd)
		}
		if len(args) != 1 || args[0] != "-c" {
			t.Errorf("Expected args ['-c'] on Unix, got %v", args)
		}
	}
}

func TestValidateBatchFile(t *testing.T) {
	// 测试不存在的文件
	err := ValidateBatchFile("/non/existent/file.bat")
	if err != ErrFileNotFound {
		t.Errorf("Expected ErrFileNotFound, got %v", err)
	}

	// 测试有效文件
	tempFile, err := ioutil.TempFile("", "test_*.bat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = ValidateBatchFile(tempFile.Name())
	if err != nil {
		t.Errorf("ValidateBatchFile failed for valid file: %v", err)
	}

	// 测试目录
	tempDir, err := ioutil.TempDir("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = ValidateBatchFile(tempDir)
	if err == nil {
		t.Error("Expected error for directory, got nil")
	}
}

// ========== 示例测试 ==========

func ExampleExecutor_ExecuteCommand() {
	executor := NewExecutor()

	response, err := executor.ExecuteCommand("echo", []string{"Hello, World!"}, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Exit Code: %d\n", response.ExitCode)
	fmt.Printf("Output: %s", response.Stdout)
	// Output:
	// Status: completed
	// Exit Code: 0
	// Output: Hello, World!
}

func ExampleExecutor_ExecuteBatch() {
	executor := NewExecutor()

	// 注意：这个示例需要实际的批处理文件
	response, err := executor.ExecuteBatch("test.bat", []string{"arg1"}, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Duration: %v\n", response.Duration)
}
