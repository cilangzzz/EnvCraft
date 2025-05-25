package config

import "tsc/internal/backend_service/model/common"

// ServiceType 定义服务类型
type ServiceType string

// ServerConfig 是所有服务配置的基础结构体
type ServerConfig struct {
	ServerName string
	Tag        string
	FileName   string
	Path       string
	common.BaseEntity
}
