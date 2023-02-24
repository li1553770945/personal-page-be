package domain

import "gorm.io/gorm"

type UserEntity struct {
	gorm.Model
	Username string
	Nickname string
	Password string
}
