package config

import (
	"time"
)

// ServiceType 定义服务类型
type ServiceType string

// BaseConfig 是所有服务配置的基础结构体
type BaseConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        ServiceType            `json:"type"`
	Description string                 `json:"description,omitempty"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        []string               `json:"tags,omitempty"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"` // 扩展字段
}

// CustomConfig 自定义服务配置
type CustomConfig struct {
	BaseConfig
	ConfigData map[string]interface{} `json:"config_data"`
}
