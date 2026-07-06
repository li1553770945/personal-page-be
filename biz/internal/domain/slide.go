package domain

import "personal-page-be/biz/internal/do"

type SlideEntity struct {
	do.BaseModel
	Slug          string `gorm:"uniqueIndex;size:128" json:"slug"`
	Title         string `json:"title"`
	TitleEn       string `json:"title_en"`
	Description   string `gorm:"type:text" json:"description"`
	DescriptionEn string `gorm:"type:text" json:"description_en"`
	Cover         string `json:"cover"`
	Entry         string `json:"entry"`
	ObjectPrefix  string `json:"object_prefix"`
	Tags          string `gorm:"type:text" json:"tags"`
	Protected     bool   `json:"protected"`
	PasswordHash  string `json:"-"`
}
