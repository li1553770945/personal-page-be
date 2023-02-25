package domain

import "gorm.io/gorm"

type FileEntity struct {
	gorm.Model
	User     UserEntity
	UserID   int
	FileName string
	FileKey  string
	Count    int
}
