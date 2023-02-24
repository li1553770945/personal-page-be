package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"personal-page-be/biz/infra/config"
)

func NewDatabase(conf *config.Config) *gorm.DB {
	dbconfig := conf.DatabaseConfig
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", dbconfig.Address, dbconfig.Username, dbconfig.Password, dbconfig.Database, dbconfig.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败:" + err.Error())
	}
	return db
}
