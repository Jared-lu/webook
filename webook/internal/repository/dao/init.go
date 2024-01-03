package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	// gorm自动建表
	return db.AutoMigrate(&User{})
}
