package migration

import (
	"log"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&MigrationTask{},
		&MigrationRecord{},
	)
	if err != nil {
		log.Printf("Failed to migrate migration tables: %v", err)
		return err
	}
	log.Println("Migration tables migrated successfully")
	return nil
}

// InitTables 初始化迁移表（包含自动迁移）
func InitTables(db *gorm.DB) error {
	return AutoMigrate(db)
}
