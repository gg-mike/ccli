package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Get() *gorm.DB {
	if db == nil {
		panic("database is not initialized")
	}
	return db
}

func Init(url string, logger logger.Interface) error {
	var err error
	db, err = gorm.Open(postgres.Open(url), &gorm.Config{Logger: logger})
	return err
}
