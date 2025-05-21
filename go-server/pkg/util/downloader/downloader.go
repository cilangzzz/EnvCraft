package downloader

import (
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

// New 创建下载 pe: 下载类型 (FTP/HTTP/WGET)
// options: 下载器配置
func New(downloadType constants.DownloadType, options core.DownloadOptions) core.Downloader {
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
