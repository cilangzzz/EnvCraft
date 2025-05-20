package core

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// WgetDownloader Wget风格下载实现
type WgetDownloader struct{}

func (d *WgetDownloader) Download(urlStr string, dest string, writer io.Writer) error {
	// 解析URL获取文件名
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL解析失败: %v", err)
	}

	// 如果dest是目录，则使用URL中的文件名
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		filename := filepath.Base(parsedURL.Path)
		if filename == "" || filename == "." {
			filename = "index.html"
		}
		dest = filepath.Join(dest, filename)
	}

	// 使用HttpClient实现下载
	client := &HttpClientDownloader{}
	return client.Download(urlStr, dest, writer)
}
