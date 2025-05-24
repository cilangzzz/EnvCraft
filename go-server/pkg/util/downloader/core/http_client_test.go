package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"tsc/pkg/util/downloader"
)

func TestHTTPDownloader_Download(t *testing.T) {
	// 创建测试HTTP服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	// 准备测试文件路径
	tmpFile, err := os.CreateTemp("", "http_test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 创建下载器
	newHTTPDownloader := NewHTTPDownloader(downloader.DownloadOptions{
		DefaultTimeout: time.Second * 5,
	})

	// 测试下载
	info := downloader.NewDownloadInfo(ts.URL, tmpFile.Name())
	if err := newHTTPDownloader.Download(info, nil); err != nil {
		t.Errorf("Download failed: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("File content mismatch, got %s", content)
	}
}

func TestHTTPDownloader_WithChecksum(t *testing.T) {
	// 创建测试HTTP服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	// 准备测试文件路径
	tmpFile, err := os.CreateTemp("", "http_checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 创建下载器
	newHTTPDownloader := NewHTTPDownloader(downloader.DownloadOptions{})

	// 测试带校验的下载
	info := downloader.NewDownloadInfo(ts.URL, tmpFile.Name())
	info.Checksum = "6fe13b5c9a94c9da9d3cc3e1977f778c"
	info.ChecksumType = "md5"
	newHTTPDownloader.Download(info, nil)
	if err := newHTTPDownloader.Download(info, nil); err != nil {
		t.Errorf("Download with checksum failed: %v", err)
	}

	// 测试错误的校验码
	info.Checksum = "wrong_checksum"
	err = newHTTPDownloader.Download(info, nil)
	if err == nil {
		t.Error("Expected checksum error, got nil")
	}
}
