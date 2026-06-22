package assembler

import (
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
)

func UserEntityToDTO(user *domain.UserEntity) *dto.UserDTO {
	return &dto.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Role:     domain.NormalizeRole(user.Role),
		CanUse:   user.CanUse,
	}
}

func UserEntityToAdminDTO(user *domain.UserEntity) *dto.AdminUserDTO {
	return &dto.AdminUserDTO{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Avatar:     user.Avatar,
		Role:       domain.NormalizeRole(user.Role),
		CanUse:     user.CanUse,
		IsActivate: user.IsActivate,
		CreatedAt:  user.CreatedAt.Unix(),
		UpdatedAt:  user.UpdatedAt.Unix(),
	}
}

func UserEntitiesToAdminDTO(users *[]domain.UserEntity) []*dto.AdminUserDTO {
	result := make([]*dto.AdminUserDTO, 0, len(*users))
	for i := range *users {
		result = append(result, UserEntityToAdminDTO(&(*users)[i]))
	}
	return result
}
