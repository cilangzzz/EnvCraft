package constants

import (
	"tsc/pkg/util/downloader/core"
)

const (
	HTTP    core.DownloadType = "HTTP"
	HTTPS   core.DownloadType = "HTTPS"
	WGET    core.DownloadType = "WGET"
	FTP     core.DownloadType = "FTP"
	UNKNOWN core.DownloadType = "UNKNOWN"
)
