package assembler

import (
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
)

func UserEntityToDTO(user *domain.UserEntity) *dto.UserDTO {
	return &dto.UserDTO{
		Username: user.Username,
		Nickname: user.Nickname,
		Role:     user.Role,
	}
}
