package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// HttpClientDownloader HTTP下载实现
type HttpClientDownloader struct{}

func (d *HttpClientDownloader) Download(url string, dest string, writer io.Writer) error {
	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求返回非200状态码: %s", resp.Status)
	}

	// 创建目标文件
	file, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 使用带进度写的writer
	var destWriter io.Writer = file
	if writer != nil {
		destWriter = io.MultiWriter(file, writer)
	}

	// 复制内容到文件
	_, err = io.Copy(destWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}
