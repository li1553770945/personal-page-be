package domain

import (
	"gorm.io/gorm"
	"personal-page-be/biz/internal/do"
)

type FileEntity struct {
	do.BaseModel
	User            UserEntity
	UserID          int
	FileName        string `json:"file_name"`
	FileKey         string `json:"file_key"`
	Count           int    `json:"count"`
	SaveName        string `json:"-"`
	Name            string `json:"name"`
	Key             string `gorm:"column:key" json:"key"`
	OSSPath         string `json:"ossPath"`
	MaxDownload     int32  `json:"maxDownload"`
	DownloadCount   int32  `json:"downloadCount"`
	ExpiredTime     gorm.DeletedAt
	DeleteOnOssTime gorm.DeletedAt
}
