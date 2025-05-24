package core

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"tsc/pkg/util/downloader"
	"tsc/pkg/util/downloader/util"
)

type ftpDownloader struct {
	options downloader.DownloadOptions
}

func NewFTPDownloader(options downloader.DownloadOptions) downloader.Downloader {
	return &ftpDownloader{
		options: options,
	}
}

func (f *ftpDownloader) Download(info downloader.DownloadInfo, writer io.Writer) error {
	// 设置默认值
	if info.Timeout == 0 {
		info.Timeout = f.options.DefaultTimeout
	}

	parsedURL, err := url.Parse(info.URL)
	if err != nil {
		return fmt.Errorf("parse FTP URL failed: %w", err)
	}

	if strings.ToLower(parsedURL.Scheme) != "ftp" {
		return fmt.Errorf("URL scheme must be ftp")
	}

	// 连接FTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), info.Timeout)
	defer cancel()

	var conn *ftp.ServerConn
	var lastError error

	for i := 0; i < info.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(info.RetryDelay)
		}

		conn, err = ftp.Dial(parsedURL.Host, ftp.DialWithContext(ctx))
		if err == nil {
			break
		}
		lastError = err
	}

	if conn == nil {
		return fmt.Errorf("connect to FTP server failed after %d retries: %w", info.MaxRetries, lastError)
	}
	defer conn.Quit()

	// 登录
	username := info.Username
	password := info.Password
	if username == "" {
		username = "anonymous"
		password = "anonymous"
	}

	if err := conn.Login(username, password); err != nil {
		return fmt.Errorf("FTP login failed: %w", err)
	}

	// 获取文件
	resp, err := conn.Retr(parsedURL.Path)
	if err != nil {
		return fmt.Errorf("retrieve file failed: %w", err)
	}
	defer resp.Close()

	// 创建目标文件
	file, err := os.OpenFile(info.Dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.FileMode)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	// 设置进度写入器
	var destWriter io.Writer = file
	if writer != nil {
		destWriter = io.MultiWriter(file, writer)
	}

	// 执行下载
	if _, err := io.Copy(destWriter, resp); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	// 校验文件
	if info.Checksum != "" {
		if err := util.VerifyChecksum(info.Dest, info.Checksum, info.ChecksumType); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	return nil
}

func (f *ftpDownloader) SetDefaultOptions(options downloader.DownloadOptions) {
	f.options = options
}
