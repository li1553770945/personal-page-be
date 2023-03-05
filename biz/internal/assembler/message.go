package assembler

import (
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
)

func MessageCategoriesEntityToDTO(entities *[]domain.MessageCategoryEntity) []*dto.MessageCategoryDTO {
	var dtos []*dto.MessageCategoryDTO
	for _, entity := range *entities {
		var categoryDTO dto.MessageCategoryDTO
		categoryDTO.Value = entity.Value
		categoryDTO.Name = entity.Name
		categoryDTO.ID = entity.ID
		dtos = append(dtos, &categoryDTO)
	}
	return dtos
}
