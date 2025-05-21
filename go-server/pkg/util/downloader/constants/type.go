package constants

// DownloadType 下载类型常量
type DownloadType string

const (
	HTTP    DownloadType = "HTTP"
	HTTPS   DownloadType = "HTTPS"
	WGET    DownloadType = "WGET"
	FTP     DownloadType = "FTP"
	UNKNOWN DownloadType = "UNKNOWN"
)
