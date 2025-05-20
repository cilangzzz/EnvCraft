package core

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/jlaffaye/ftp"
)

// FTPDownloader FTP下载实现
type FTPDownloader struct {
	Username string
	Password string
}

func (d *FTPDownloader) Download(urlStr string, dest string, writer io.Writer) error {
	// 解析FTP URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("FTP URL解析失败: %v", err)
	}

	if strings.ToLower(parsedURL.Scheme) != "ftp" {
		return fmt.Errorf("URL协议必须是ftp")
	}

	// 连接FTP服务器
	client, err := ftp.Dial(parsedURL.Host)
	if err != nil {
		return fmt.Errorf("FTP连接失败: %v", err)
	}
	defer client.Quit()

	// 使用提供的凭据或匿名登录
	user := d.Username
	pass := d.Password
	if user == "" {
		user = "anonymous"
		pass = "anonymous"
	}

	err = client.Login(user, pass)
	if err != nil {
		return fmt.Errorf("FTP登录失败: %v", err)
	}

	// 获取文件
	resp, err := client.Retr(parsedURL.Path)
	if err != nil {
		return fmt.Errorf("FTP获取文件失败: %v", err)
	}
	defer resp.Close()

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
	_, err = io.Copy(destWriter, resp)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}
