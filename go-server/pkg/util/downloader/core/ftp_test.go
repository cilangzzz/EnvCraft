package core

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tsc/pkg/util/downloader"
)

// mockFTPServer 模拟FTP服务器
func mockFTPServer(t *testing.T, dir string) (string, func()) {
	// 创建测试文件
	testFile := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 启动TCP监听
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	// 返回服务器地址
	addr := listener.Addr().(*net.TCPAddr)

	// 使用简单的TCP echo服务器模拟FTP控制连接
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("220 Mock FTP Server\r\n"))

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}

					cmd := string(buf[:n])
					switch {
					case strings.HasPrefix(cmd, "USER"):
						c.Write([]byte("331 Password required\r\n"))
					case strings.HasPrefix(cmd, "PASS"):
						c.Write([]byte("230 User logged in\r\n"))
					case strings.HasPrefix(cmd, "RETR"):
						c.Write([]byte("150 Opening data connection\r\n"))
						// 模拟数据传输
						dataConn, err := net.Dial("tcp", "127.0.0.1:0")
						if err != nil {
							c.Write([]byte("425 Can't open data connection\r\n"))
							return
						}
						defer dataConn.Close()
						dataConn.Write([]byte("test content"))
						c.Write([]byte("226 Transfer complete\r\n"))
					case strings.HasPrefix(cmd, "QUIT"):
						c.Write([]byte("221 Goodbye\r\n"))
						return
					default:
						c.Write([]byte("500 Command not understood\r\n"))
					}
				}
			}(conn)
		}
	}()

	return addr.String(), func() {
		listener.Close()
	}
}

func TestFTPDownloader_Download(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ftp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 启动模拟FTP服务器
	addr, shutdown := mockFTPServer(t, tempDir)
	defer shutdown()

	// 准备测试文件路径
	tmpFile, err := os.CreateTemp("", "ftp_test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 创建下载器
	downloader := NewFTPDownloader(downloader.DownloadOptions{
		DefaultTimeout: time.Second * 5,
	})

	// 测试下载
	info := NewDownloadInfo("ftp://"+addr+"/testfile.txt", tmpFile.Name())
	info.Username = "testuser"
	info.Password = "testpass"

	if err := downloader.Download(info, nil); err != nil {
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

func TestFTPDownloader_WithChecksum(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ftp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 启动模拟FTP服务器
	addr, shutdown := mockFTPServer(t, tempDir)
	defer shutdown()

	// 准备测试文件路径
	tmpFile, err := os.CreateTemp("", "ftp_checksum_test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 创建下载器
	downloader := NewFTPDownloader(downloader.DownloadOptions{})

	// 测试带校验的下载
	info := NewDownloadInfo("ftp://"+addr+"/testfile.txt", tmpFile.Name())
	info.Username = "testuser"
	info.Password = "testpass"
	info.Checksum = "6fe13b5c9a94c9da9d3cc3e1977f778c" // "test content"的MD5
	info.ChecksumType = "md5"

	if err := downloader.Download(info, nil); err != nil {
		t.Errorf("Download with checksum failed: %v", err)
	}

	// 测试错误的校验码
	info.Checksum = "wrong_checksum"
	err = downloader.Download(info, nil)
	if err == nil {
		t.Error("Expected checksum error, got nil")
	}
}
