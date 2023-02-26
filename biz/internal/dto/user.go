package dto

import "personal-page-be/biz/internal/domain"

type UserDTO struct {
	Username string `json:"username" vd:"len($)>5"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type GenerateActivateCodeReq struct {
	Username string `json:"username" vd:"len($)>5"`
}

type RegisterReq struct {
	domain.UserEntity
	ActivateCode string `json:"activate_code" vd:"len($)>5"`
}
