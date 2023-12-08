package domain

import (
	"personal-page-be/biz/internal/do"
)

type MessageCategoryEntity struct {
	do.BaseModel
	Name   string `json:"name"`
	Value  string `json:"value"`
	CanUse bool   `json:"can_use"`
}
type MessageEntity struct {
	do.BaseModel
	Category   MessageCategoryEntity `json:"category"`
	CategoryID int                   `json:"category_id"`
	Title      string                `json:"title"`
	Message    string                `json:"message"`
	Name       string                `json:"name"`
	Contact    string                `json:"contact"`
	HaveRead   bool                  `json:"have_read"`
	UUID       string                `json:"uuid"`
}

type ReplyEntity struct {
	do.BaseModel
	Content   string `json:"content"`
	Message   MessageEntity
	MessageID uint `json:"message_id"`
	HaveRead  bool `json:"have_read"`
}
