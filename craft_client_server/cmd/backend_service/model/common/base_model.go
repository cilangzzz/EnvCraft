package common

import "time"

// ExtensionField 扩展字段类型(JSON格式)
type ExtensionField map[string]interface{}

// BaseEntity 基础实体类
type BaseEntity struct {
	ID          uint64         `json:"id" gorm:"primaryKey;comment:主键ID"`                         // 主键ID
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`             // 创建时间
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`             // 更新时间
	DelFlag     int8           `json:"del_flag" gorm:"default:0;index;comment:删除标志(0-未删除 1-已删除)"` // 删除标志
	DeptID      uint64         `json:"dept_id" gorm:"comment:部门ID"`                               // 部门ID
	Description string         `json:"description" gorm:"type:text;comment:描述信息"`                 // 描述信息
	Extension   ExtensionField `json:"extension" gorm:"type:json;comment:扩展字段"`                   // 扩展字段(JSON格式)
}
