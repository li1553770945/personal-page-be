package dto

import "personal-page-be/biz/internal/domain"

type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username" vd:"len($)>5"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	CanUse   bool   `json:"can_use"`
}

type GenerateActivateCodeReq struct {
	Username string `json:"username" vd:"len($)>5"`
}

type RegisterReq struct {
	domain.UserEntity
	ActivateCode string `json:"activate_code"`
	ActiveCode   string `json:"activeCode"`
}

type AdminUserDTO struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Role       string `json:"role"`
	CanUse     bool   `json:"can_use"`
	IsActivate bool   `json:"is_activate"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type UpdateUserRoleReq struct {
	Role string `json:"role"`
}

type UpdateUserStatusReq struct {
	CanUse bool `json:"can_use"`
}

type UserDangerActionReq struct {
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

type UserDangerActionResp struct {
	User         *AdminUserDTO `json:"user,omitempty"`
	ActivateCode string        `json:"activate_code,omitempty"`
	ActiveCode   string        `json:"activeCode,omitempty"`
	RelatedFiles int64         `json:"related_files"`
}
