package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWgetDownloader_Download(t *testing.T) {
	// 创建测试HTTP服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "wget_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建下载器
	downloader := NewWgetDownloader(DownloadOptions{})

	// 测试下载到目录
	info := NewDownloadInfo(ts.URL, tmpDir)
	if err := downloader.Download(info, nil); err != nil {
		t.Errorf("Download failed: %v", err)
	}

	// 验证文件是否创建
	expectedFile := filepath.Join(tmpDir, "index.html")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("File not created: %s", expectedFile)
	}

	// 验证文件内容
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("File content mismatch, got %s", content)
	}
}

func TestWgetDownloader_WithFilename(t *testing.T) {
	// 创建测试HTTP服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "wget_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建下载器
	downloader := NewWgetDownloader(DownloadOptions{})

	// 测试带文件名的URL
	info := NewDownloadInfo(ts.URL+"/testfile.txt", tmpDir)
	if err := downloader.Download(info, nil); err != nil {
		t.Errorf("Download failed: %v", err)
	}

	// 验证文件名是否正确
	expectedFile := filepath.Join(tmpDir, "testfile.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("File not created: %s", expectedFile)
	}
}
