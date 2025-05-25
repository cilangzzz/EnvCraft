package db

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
	"tsc/internal/cfg"
)

var Gorm = new(_gorm)

type _gorm struct{}

// Config gorm 自定义配置
// Author [SliverHorn](https://github.com/SliverHorn)
func (g *_gorm) Config(prefix string, singular bool) *gorm.Config {
	var general cfg.GeneralDB
	switch cfg.GlobalServerConfig.DbType {
	case "mysql":
		general = cfg.GlobalServerConfig.DbConfig.Mysql.GeneralDB
	//case "pgsql":
	//	general = global.GVA_CONFIG.Pgsql.GeneralDB
	//case "oracle":
	//	general = global.GVA_CONFIG.Oracle.GeneralDB
	case "sqlite":
		general = cfg.GlobalServerConfig.DbConfig.Sqlite.GeneralDB
	//case "mssql":
	//	general = global.GVA_CONFIG.Mssql.GeneralDB
	default:
		general = cfg.GlobalServerConfig.DbConfig.Mysql.GeneralDB
	}
	return &gorm.Config{
		Logger: logger.New(NewWriter(general, log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      general.LogLevel(),
			Colorful:      true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   prefix,
			SingularTable: singular,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}
}

type Writer struct {
	config cfg.GeneralDB
	writer logger.Writer
}

func NewWriter(config cfg.GeneralDB, writer logger.Writer) *Writer {
	return &Writer{config: config, writer: writer}
}

// Printf 格式化打印日志
func (c *Writer) Printf(message string, data ...any) {
	if c.config.LogZap {
		switch c.config.LogLevel() {
		case logger.Silent:
			zap.L().Debug(fmt.Sprintf(message, data...))
		case logger.Error:
			zap.L().Error(fmt.Sprintf(message, data...))
		case logger.Warn:
			zap.L().Warn(fmt.Sprintf(message, data...))
		case logger.Info:
			zap.L().Info(fmt.Sprintf(message, data...))
		default:
			zap.L().Info(fmt.Sprintf(message, data...))
		}
		return
	}
	c.writer.Printf(message, data...)
}
