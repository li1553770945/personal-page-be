package domain

import "gorm.io/gorm"

type FileEntity struct {
	gorm.Model
	User     UserEntity
	UserID   int
	FileName string `json:"file_name"`
	FileKey  string `json:"file_key"`
	Count    int    `json:"count"`
}
