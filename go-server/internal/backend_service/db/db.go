// Package db 用于暴露给外部使用获取数据库引擎
package db

import (
	"gorm.io/gorm"
	"log"
	"sync"
	"tsc/internal/cfg"
)

var (
	DbEngine     *gorm.DB
	DBEngineList map[string]*gorm.DB
	DbType       string
	lock         sync.RWMutex
)

// InitDB 初始化数据库
func InitDB(dbType string) (*gorm.DB, error) {
	lock.Lock()
	defer lock.Unlock()
	DbType = dbType
	switch DbType {
	case cfg.DB_MYSQL:
		DbEngine = GormMysql()
		return DbEngine, nil
	case cfg.DB_SQLITE:
		DbEngine = GormSqlite()
		return DbEngine, nil
	default:
		return nil, nil
	}
}

// RegisterDB 注册数据库
func RegisterDB(dbType string) {
	lock.Lock()
	defer lock.Unlock()
	dbEngine, err := InitDB(dbType)
	if err != nil {
		log.Fatal("init db error: ", err)
	} else {
		DbEngine = dbEngine
	}
}
