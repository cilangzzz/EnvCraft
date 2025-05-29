package core

import (
	"io"
	"net/http"
	"time"
)

// DownloadType 下载类型常量
type DownloadType string

// DownloadInfo 扩展的下载信息结构
type DownloadInfo struct {
	URL            string            `gorm:"type:varchar(1024);not null;comment:下载URL"`
	Dest           string            `gorm:"type:varchar(512);not null;comment:目标路径"`
	Type           DownloadType      `gorm:"type:varchar(20);comment:下载类型"`
	Username       string            `gorm:"type:varchar(128);comment:FTP/HTTP认证用户名"`
	Password       string            `gorm:"type:varchar(256);comment:FTP/HTTP认证密码"`
	Headers        map[string]string `gorm:"type:json;comment:HTTP请求头"` // 自定义JSONMap类型
	Timeout        time.Duration     `gorm:"type:bigint;comment:请求超时时间(毫秒)"`
	Checksum       string            `gorm:"type:varchar(128);comment:文件校验值"`
	ChecksumType   string            `gorm:"type:varchar(20);comment:校验类型"`
	MaxRetries     int               `gorm:"type:int;default:3;comment:最大重试次数"`
	RetryDelay     time.Duration     `gorm:"type:bigint;comment:重试延迟时间(毫秒)"`
	FileMode       int               `gorm:"type:int;default:420;comment:目标文件权限(十进制)"` // 0644 = 420
	ResumeDownload bool              `gorm:"type:tinyint(1);default:1;comment:是否支持断点续传"`
	ProxyURL       string            `gorm:"type:varchar(512);comment:代理服务器地址"`
	UserAgent      string            `gorm:"type:varchar(256);comment:自定义User-Agent"`
}

// Downloader 下载器接口
type Downloader interface {
	// Download 执行下载操作
	Download(info DownloadInfo, writer io.Writer) error

	// SetDefaultOptions 设置默认下载选项
	SetDefaultOptions(options DownloadOptions)
}

// DownloadOptions 下载器配置选项
type DownloadOptions struct {
	DefaultTimeout    time.Duration
	DefaultMaxRetries int
	DefaultUserAgent  string
	HTTPClient        *http.Client
}

// ProgressWriter 进度写入器接口
type ProgressWriter interface {
	io.Writer
	SetTotal(total int64)
}

// NewDownloadInfo 创建DownloadInfo实例
func NewDownloadInfo(url, dest string) DownloadInfo {
	return DownloadInfo{
		URL:          url,
		Dest:         dest,
		MaxRetries:   3,
		RetryDelay:   time.Second * 5,
		FileMode:     0644,
		Headers:      make(map[string]string),
		Timeout:      time.Second * 30,
		ChecksumType: "md5",
	}
}
