package constants

// DownloadType 下载类型常量
type DownloadType string

const (
	HTTP DownloadType = "HTTP"
	WGET DownloadType = "WGET"
	FTP  DownloadType = "FTP"
)
