package server

import (
	"tsc/internal/backend_service/model/common"
	downloader "tsc/pkg/util/downloader/core"
	server_ini "tsc/pkg/util/server_ini/core"
)

// ServerConfig 下载任务模型
type ServerConfig struct {
	common.BaseEntity

	// 基本信息
	ServerName string `json:"name" gorm:"size:255;comment:任务名称"`
	Tag        string `json:"tag" gorm:"size:255;comment:任务标签"`

	// 文件信息
	DownloadInfo downloader.DownloadInfo `json:"download_info" gorm:"type:json;comment:下载信息"`
	PackageInfo  server_ini.PackageInfo  `json:"package_info" gorm:"type:json;comment:文件信息"`
}
