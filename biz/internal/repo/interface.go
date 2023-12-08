package repo

import (
	"gorm.io/gorm"
	"personal-page-be/biz/internal/domain"
)

type IRepository interface {
	FindUser(username string) (*domain.UserEntity, error)
	SaveUser(user *domain.UserEntity) error

	FindFileByFileKey(fileKey string) (*domain.FileEntity, error)
	FindFileBySaveName(saveName string) (*domain.FileEntity, error)
	SaveFile(user *domain.FileEntity) error
	RemoveFile(fileID uint) error

	FindAllMessageCategory() (*[]domain.MessageCategoryEntity, error)
	SaveMessage(entity *domain.MessageEntity) error
	FindMessageByUUID(uuid string) (*domain.MessageEntity, error)
	FindReplyByMessageID(messageId uint) (*domain.ReplyEntity, error)
	SaveReply(entity *domain.ReplyEntity) error
	FindMessageByID(messageId uint) (*domain.MessageEntity, error)
	GetUnreadMsg() (*[]domain.MessageEntity, error)
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
	err = db.AutoMigrate(&domain.MessageCategoryEntity{})
	if err != nil {
		panic("迁移消息类别模型失败：" + err.Error())
	}
	err = db.AutoMigrate(&domain.MessageEntity{})
	if err != nil {
		panic("迁移消息模型失败：" + err.Error())
	}
	err = db.AutoMigrate(&domain.ReplyEntity{})
	if err != nil {
		panic("迁移回复模型失败：" + err.Error())
	}
	return &Repository{
		DB: db,
	}
}
