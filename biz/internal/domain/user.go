package domain

import (
	"personal-page-be/biz/internal/do"
)

type UserEntity struct {
	do.BaseModel
	Username     string `vd:"len($)>5" gorm:"index:username_idx,unique"`
	Nickname     string
	Password     string
	Role         string
	CanUse       bool
	IsActivate   bool
	ActivateCode string
}
