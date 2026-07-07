package domain

import "personal-page-be/biz/internal/do"

type AdminAuditLogEntity struct {
	do.BaseModel
	ActorUserID    uint   `gorm:"index"`
	ActorUsername  string `gorm:"size:191;index"`
	TargetUserID   uint   `gorm:"index"`
	TargetUsername string `gorm:"size:191;index"`
	Action         string `gorm:"size:64;index"`
	Reason         string `gorm:"size:512"`
	Metadata       string `gorm:"type:text"`
}
