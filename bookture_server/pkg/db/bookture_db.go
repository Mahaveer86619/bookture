package db

import (
	"fmt"

	"github.com/Mahaveer86619/bookture/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dsn := "host=" + config.AppConfig.DatabaseHost +
		" user=" + config.AppConfig.DatabaseUser +
		" password=" + config.AppConfig.DatabasePass +
		" dbname=" + config.AppConfig.DatabaseName +
		" port=" + fmt.Sprintf("%d", config.AppConfig.DatabasePort) +
		" sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db
	return err
}

func GetDB() *gorm.DB {
	return DB
}
