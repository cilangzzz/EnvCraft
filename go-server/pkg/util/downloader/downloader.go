package downloader

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"tsc/pkg/util/downloader/constants"
	"tsc/pkg/util/downloader/core"
)

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

// New 创建下载器实例
func New(downloadType constants.DownloadType, opts ...interface{}) (*Downloader, error) {
	var coreDownloader core.Downloader
	var err error

	switch downloadType {
	case constants.HTTP:
		var httpOpts []http.Option
		for _, opt := range opts {
			if hOpt, ok := opt.(http.Option); ok {
				httpOpts = append(httpOpts, hOpt)
			}
		}
		coreDownloader = http.NewDefaultDownloader(&http.Client{})
	case constants.Wget:
		var httpOpts []http.Option
		for _, opt := range opts {
			if hOpt, ok := opt.(http.Option); ok {
				httpOpts = append(httpOpts, hOpt)
			}
		}
		coreDownloader = wget.New(httpOpts...)
	case constants.FTP:
		var username, password string
		for _, opt := range opts {
			if creds, ok := opt.(struct{ user, pass string }); ok {
				username = creds.user
				password = creds.pass
			}
		}
		coreDownloader = ftp.New(username, password)
	default:
		return nil, fmt.Errorf("不支持的下载类型: %s", downloadType)
	}

	return &Downloader{
		coreDownloader: coreDownloader,
	}, nil
}

// Download 执行下载
func (d *Downloader) Download(url, dest string) error {
	return d.coreDownloader.Download(url, dest)
}

// SetProgressWriter 设置进度写入器
func (d *Downloader) SetProgressWriter(writer io.Writer) {
	d.coreDownloader.SetProgressWriter(writer)
}

// SimpleDownload 简单下载(自动判断类型)
func SimpleDownload(url, dest string, writer io.Writer) error {
	var downloadType constants.DownloadType

	switch {
	case strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"):
		downloadType = constants.HTTP
	case strings.HasPrefix(url, "ftp://"):
		downloadType = constants.FTP
	default:
		return fmt.Errorf("无法识别的URL协议")
	}

	downloader, err := New(downloadType)
	if err != nil {
		return err
	}

	if writer != nil {
		downloader.SetProgressWriter(writer)
	}

	return downloader.Download(url, dest)
}
