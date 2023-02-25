package domain

import "gorm.io/gorm"

type UserEntity struct {
	gorm.Model
	Username     string `vd:"len($)>5" gorm:"index:username_idx,unique"`
	Nickname     string
	Password     string
	Role         string
	CanUse       bool
	IsActivate   bool
	ActivateCode string
}
