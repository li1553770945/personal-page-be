package repo

import "personal-page-be/biz/internal/domain"

func (Repo *Repository) FindAllMessageCategory() (*[]domain.MessageCategoryEntity, error) {
	var entity []domain.MessageCategoryEntity
	err := Repo.DB.Where("can_use = ?", true).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) SaveMessage(entity *domain.MessageEntity) error {
	if entity.ID == 0 {
		err := Repo.DB.Create(&entity).Error
		Repo.DB.Preload("Category").Find(&entity)
		return err
	} else {
		err := Repo.DB.Save(&entity).Error
		return err
	}
}
