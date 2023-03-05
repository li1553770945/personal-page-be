package domain

import "gorm.io/gorm"

type MessageCategoryEntity struct {
	gorm.Model
	Name   string `json:"name"`
	Value  string `json:"value"`
	CanUse bool   `json:"can_use"`
}
type MessageEntity struct {
	gorm.Model
	Category   MessageCategoryEntity
	CategoryID int    `json:"category_id"`
	Message    string `json:"message"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	HaveRead   bool   `json:"have_read"`
}
