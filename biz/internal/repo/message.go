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

func (Repo *Repository) FindMessageByUUID(uuid string) (*domain.MessageEntity, error) {
	var entity domain.MessageEntity
	err := Repo.DB.Where("uuid = ?", uuid).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
func (Repo *Repository) FindMessageByID(messageId uint) (*domain.MessageEntity, error) {
	var entity domain.MessageEntity
	err := Repo.DB.Where("id = ?", messageId).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
func (Repo *Repository) FindReplyByMessageID(messageId uint) (*domain.ReplyEntity, error) {
	var entity domain.ReplyEntity
	err := Repo.DB.Where("message_id = ?", messageId).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) SaveReply(entity *domain.ReplyEntity) error {
	if entity.ID == 0 {
		err := Repo.DB.Create(&entity).Error
		return err
	} else {
		err := Repo.DB.Save(&entity).Error
		return err
	}
}
