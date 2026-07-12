package repo

import (
	"time"

	"gorm.io/gorm"
	"personal-page-be/biz/internal/domain"
)

type IRepository interface {
	FindUser(username string) (*domain.UserEntity, error)
	FindUserByID(userID uint) (*domain.UserEntity, error)
	FindFirstUserByRole(role string) (*domain.UserEntity, error)
	ListUsers() (*[]domain.UserEntity, error)
	CountUsersByRole(role string, canUseOnly bool) (int64, error)
	SaveUser(user *domain.UserEntity) error
	SaveUserAndAudit(user *domain.UserEntity, audit *domain.AdminAuditLogEntity) error
	RemoveUserAndAudit(userID uint, audit *domain.AdminAuditLogEntity) error
	SaveAIUsage(usage *domain.AIUsageEntity) error
	ReserveAIUsageDailyQuota(quotaDay string, identityKey string, ipKey string, limit int) (bool, error)
	ListAIUsage(userID *uint, startAt *time.Time, endAt *time.Time, model string, channel string) (*[]domain.AIUsageEntity, error)

	FindFileByID(fileID uint) (*domain.FileEntity, error)
	FindFileByFileKey(fileKey string) (*domain.FileEntity, error)
	FindFileByKey(fileKey string) (*domain.FileEntity, error)
	FindFileBySaveName(saveName string) (*domain.FileEntity, error)
	ListFiles(userID *uint) (*[]domain.FileEntity, error)
	CountFilesByUserID(userID uint) (int64, error)
	SaveFile(user *domain.FileEntity) error
	RemoveFile(fileID uint) error
	RemoveFileByKey(fileKey string) error

	ListSlides() (*[]domain.SlideEntity, error)
	FindSlideByID(slideID uint) (*domain.SlideEntity, error)
	FindSlideBySlug(slug string) (*domain.SlideEntity, error)
	SaveSlide(slide *domain.SlideEntity) error
	RemoveSlide(slideID uint) error

	FindAllMessageCategory() (*[]domain.MessageCategoryEntity, error)
	SaveMessage(entity *domain.MessageEntity) error
	FindMessageByUUID(uuid string) (*domain.MessageEntity, error)
	FindReplyByMessageID(messageId uint) (*domain.ReplyEntity, error)
	SaveReply(entity *domain.ReplyEntity) error
	FindMessageByID(messageId uint) (*domain.MessageEntity, error)
	GetUnreadMsg() (*[]domain.MessageEntity, error)

	FindAllFeedbackCategory() (*[]domain.FeedbackCategoryEntity, error)
	SaveFeedback(entity *domain.FeedbackEntity) error
	FindFeedbackByUUID(uuid string) (*domain.FeedbackEntity, error)
	FindFeedbackByID(feedbackID uint) (*domain.FeedbackEntity, error)
	FindReplyByFeedbackID(feedbackID uint) (*domain.FeedbackReplyEntity, error)
	SaveFeedbackReply(entity *domain.FeedbackReplyEntity) error
	GetUnreadFeedback() (*[]domain.FeedbackEntity, error)

	SaveProject(project *domain.ProjectEntity) error
	RemoveProject(projectID uint) error
	GetProject(projectID uint) (*domain.ProjectEntity, error)
	GetProjectsNum() (int64, error)
	GetProjects(from int, end int, order string, status int) (*[]domain.ProjectEntity, error)
}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	if err := db.AutoMigrate(&domain.UserEntity{}); err != nil {
		panic("migrate user model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.AdminAuditLogEntity{}); err != nil {
		panic("migrate admin audit log model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.AIUsageEntity{}); err != nil {
		panic("migrate ai usage model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.AIUsageDailyQuotaEntity{}); err != nil {
		panic("migrate ai usage daily quota model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.FileEntity{}); err != nil {
		panic("migrate file model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.SlideEntity{}); err != nil {
		panic("migrate slide model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.MessageCategoryEntity{}); err != nil {
		panic("migrate message category model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.FeedbackCategoryEntity{}); err != nil {
		panic("migrate feedback category model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.MessageEntity{}); err != nil {
		panic("migrate message model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.FeedbackEntity{}); err != nil {
		panic("migrate feedback model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.ReplyEntity{}); err != nil {
		panic("migrate reply model failed: " + err.Error())
	}
	if err := db.AutoMigrate(&domain.ProjectEntity{}); err != nil {
		panic("migrate project model failed: " + err.Error())
	}
	return &Repository{
		DB: db,
	}
}
