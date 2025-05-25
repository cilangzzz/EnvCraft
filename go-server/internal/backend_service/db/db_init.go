package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"tsc/internal/cfg"
)

// GormMysql 初始化Mysql数据库
// Author [piexlmax](https://github.com/piexlmax)
// Author [SliverHorn](https://github.com/SliverHorn)
func GormMysql() *gorm.DB {
	m := cfg.GlobalServerConfig.DbConfig.Mysql
	if m.Dbname == "" {
		log.Fatalf("请检查Mysql数据库配置是否正确")
		return nil
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), Gorm.Config(m.Prefix, m.Singular)); err != nil {
		log.Fatalf("MySQL连接失败: %v", err)
		return nil
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}

// GormSqlite 初始化Sqlite数据库
func GormSqlite() *gorm.DB {
	s := cfg.GlobalServerConfig.DbConfig.Sqlite
	if s.Dbname == "" {
		log.Fatalf("请检查Sqlite数据库配置是否正确")
		return nil
	}

	if db, err := gorm.Open(sqlite.Open(s.Dsn()), Gorm.Config(s.Prefix, s.Singular)); err != nil {
		log.Fatalf("Sqlite连接失败: %v", err)
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(s.MaxIdleConns)
		sqlDB.SetMaxOpenConns(s.MaxOpenConns)
		return db
	}
}
