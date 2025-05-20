package downloader

import "tsc/pkg/util/downloader/constants"

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
