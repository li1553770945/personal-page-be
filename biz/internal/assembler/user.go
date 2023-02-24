package assembler

import "personal-page-be/biz/internal/domain"

type UserDTO struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

func UserEntityToDTO(user *domain.UserEntity) *UserDTO {
	return &UserDTO{
		Username: user.Username,
		Nickname: user.Nickname,
	}
}
