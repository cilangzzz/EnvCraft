package server

import (
	"tsc/internal/backend_service/model/common"
	downloaderCore "tsc/pkg/util/downloader/core"
	serverIniCore "tsc/pkg/util/server_ini/core"
)

// ServerConfig 下载任务模型
type ServerConfig struct {
	common.BaseEntity

	// 基本信息
	ServerName string `json:"name" gorm:"size:255;comment:任务名称"`
	Tag        string `json:"tag" gorm:"size:255;comment:任务标签"`

	// 文件信息
	DownloadInfoId string `json:"download_info_id" gorm:"type:json;comment:下载信息"`
	PackageInfoId  string `json:"package_info_id" gorm:"type:json;comment:文件信息"`
}

// ServerDownloadInfo 下载信息模型
type ServerDownloadInfo struct {
	common.BaseEntity
	// 文件信息
	DownloadInfo downloaderCore.DownloadInfo `json:"download_info" gorm:"type:json;comment:下载信息"`
}

// ServerPackageInfo 文件信息模型
type ServerPackageInfo struct {
	common.BaseEntity
	// 文件信息
	PackageInfo serverIniCore.PackageInfo `json:"package_info" gorm:"type:json;comment:文件信息"`
}
