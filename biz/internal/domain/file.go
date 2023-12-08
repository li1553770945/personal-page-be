package domain

import (
	"personal-page-be/biz/internal/do"
)

type FileEntity struct {
	do.BaseModel
	User     UserEntity
	UserID   int
	FileName string `json:"file_name"`
	FileKey  string `json:"file_key"`
	Count    int    `json:"count"`
	SaveName string `json:"-"`
}
