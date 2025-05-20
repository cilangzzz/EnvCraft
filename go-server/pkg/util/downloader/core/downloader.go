package core

import (
	"io"
	"tsc/pkg/util/downloader/constants"
)

// Downloader 下载器接口
type Downloader interface {
	// Download 下载文件
	// url: 文件URL
	// dest: 目标路径(可以是文件路径或目录)
	// writer: 可选，用于进度写入(如实现进度条)
	Download(url string, dest string, writer io.Writer) error
}

// DownloadInfo 下载信息
type DownloadInfo struct {
	URL      string                 // 下载URL
	Dest     string                 // 目标路径
	Type     constants.DownloadType // 下载类型
	Username string                 // FTP用户名(可选)
	Password string                 // FTP密码(可选)
}
