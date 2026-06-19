package database

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"personal-page-be/biz/infra/config"
)

func NewDatabase(conf *config.Config) *gorm.DB {
	dbconfig := conf.DatabaseConfig
	driver := strings.ToLower(dbconfig.Driver)
	if v := os.Getenv("DB_DRIVER"); v != "" {
		driver = strings.ToLower(v)
	}
	if driver == "" {
		driver = "postgres"
	}

	var db *gorm.DB
	var err error
	switch driver {
	case "mysql":
		tls := "false"
		if dbconfig.UseTLS {
			tls = "true"
		}
		port := dbconfig.Port
		if port == 0 {
			port = 3306
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", dbconfig.Username, dbconfig.Password, dbconfig.Address, port, dbconfig.Database, tls)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		port := dbconfig.Port
		if port == 0 {
			port = 5432
		}
		sslmode := dbconfig.SSLMode
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai", dbconfig.Address, dbconfig.Username, dbconfig.Password, dbconfig.Database, port, sslmode)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		panic("database connection failed: " + err.Error())
	}
	return db
}
