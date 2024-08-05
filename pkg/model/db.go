package model

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db            *gorm.DB
	migrateTables []interface{}
)

func GetDB() *gorm.DB {
	return db
}

func registerTable(t interface{}) {
	migrateTables = append(migrateTables, t)
}

func InitDB(path string) error {
	var err error
	// 连接数据库
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}

	// 连接池设置
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	for _, table := range migrateTables {
		db.AutoMigrate(table)
	}

	return nil
}
