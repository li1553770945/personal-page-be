package repo

import (
	"gorm.io/gorm"
	"personal-page-be/biz/internal/domain"
)

type IRepository interface {
	FindUser(username string) (*domain.UserEntity, error)
	SaveUser(user *domain.UserEntity) error

	FindFile(fileKey string) (*domain.FileEntity, error)
	SaveFile(user *domain.FileEntity) error
	RemoveFile(fileID int) error
}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	err := db.AutoMigrate(&domain.UserEntity{})
	if err != nil {
		panic("迁移用户模型失败：" + err.Error())
	}
	err = db.AutoMigrate(&domain.FileEntity{})
	if err != nil {
		panic("迁移文件模型失败：" + err.Error())
	}
	return &Repository{
		DB: db,
	}
}
