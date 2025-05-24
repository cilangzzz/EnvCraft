package downloader

import (
	"io"
	"net/http"
	"os"
	"time"
	"tsc/pkg/util/downloader/constants"
	"tsc/pkg/util/downloader/core"
)

// DownloadInfo 扩展的下载信息结构
type DownloadInfo struct {
	URL            string                 // 下载URL (必需)
	Dest           string                 // 目标路径 (必需)
	Type           constants.DownloadType // 下载类型 (可选，自动推断)
	Username       string                 // FTP/HTTP认证用户名
	Password       string                 // FTP/HTTP认证密码
	Headers        map[string]string      // HTTP请求头
	Timeout        time.Duration          // 请求超时时间
	Checksum       string                 // 文件校验值 (md5/sha1等)
	ChecksumType   string                 // 校验类型 (md5/sha1等)
	MaxRetries     int                    // 最大重试次数
	RetryDelay     time.Duration          // 重试延迟时间
	FileMode       os.FileMode            // 目标文件权限 (默认0644)
	ResumeDownload bool                   // 是否支持断点续传
	ProxyURL       string                 // 代理服务器地址
	UserAgent      string                 // 自定义User-Agent
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

// DownloadType 下载类型常量
var DownloadType = struct {
	HTTP constants.DownloadType
	WGET constants.DownloadType
	FTP  constants.DownloadType
}{
	HTTP: constants.HTTP,
	WGET: constants.WGET,
	FTP:  constants.FTP,
}

// New 创建下载 pe: 下载类型 (FTP/HTTP/WGET)
// options: 下载器配置
func New(downloadType constants.DownloadType, options DownloadOptions) Downloader {
	switch downloadType {
	case DownloadType.FTP:
		return core.NewFTPDownloader(options)
	case DownloadType.HTTP:
		return core.NewFTPDownloader(options)
	case DownloadType.WGET:
		return core.NewFTPDownloader(options)
	default:
		return core.NewFTPDownloader(options) // 默认返回HTTP下载器
	}
}
