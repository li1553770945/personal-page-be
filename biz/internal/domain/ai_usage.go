package domain

import "personal-page-be/biz/internal/do"

const (
	AIUsageStatusSuccess = "success"
	AIUsageStatusFailed  = "failed"
)

type AIUsageEntity struct {
	do.BaseModel
	UserID           uint    `gorm:"index"`
	Username         string  `gorm:"size:191;index"`
	Channel          string  `gorm:"size:191;index"`
	Model            string  `gorm:"size:191;index"`
	ConversationID   string  `gorm:"size:191;index"`
	MessageID        string  `gorm:"size:191;index"`
	RequestID        string  `gorm:"size:64;uniqueIndex"`
	Status           string  `gorm:"size:32;index"`
	PromptTokens     int64   `gorm:"index"`
	CompletionTokens int64   `gorm:"index"`
	TotalTokens      int64   `gorm:"index"`
	TotalPrice       float64 `gorm:"index"`
	Currency         string  `gorm:"size:16"`
	ErrorMessage     string  `gorm:"size:1024"`
}
