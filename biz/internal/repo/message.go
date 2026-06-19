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
	}
	return Repo.DB.Save(&entity).Error
}

func (Repo *Repository) FindMessageByUUID(uuid string) (*domain.MessageEntity, error) {
	var entity domain.MessageEntity
	err := Repo.DB.Preload("Category").Where("uuid = ?", uuid).Find(&entity).Error
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
		return Repo.DB.Create(&entity).Error
	}
	return Repo.DB.Save(&entity).Error
}

func (Repo *Repository) GetUnreadMsg() (*[]domain.MessageEntity, error) {
	var entity []domain.MessageEntity
	err := Repo.DB.Preload("Category").Select("ID", "Name", "CreatedAt", "Title", "UUID").Where("have_read = ?", false).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) FindAllFeedbackCategory() (*[]domain.FeedbackCategoryEntity, error) {
	var entity []domain.FeedbackCategoryEntity
	err := Repo.DB.Where("can_use = ?", true).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) SaveFeedback(entity *domain.FeedbackEntity) error {
	if entity.ID == 0 {
		err := Repo.DB.Create(&entity).Error
		Repo.DB.Preload("Category").Find(&entity)
		return err
	}
	return Repo.DB.Save(&entity).Error
}

func (Repo *Repository) FindFeedbackByUUID(uuid string) (*domain.FeedbackEntity, error) {
	var entity domain.FeedbackEntity
	err := Repo.DB.Preload("Category").Where("uuid = ?", uuid).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) FindFeedbackByID(feedbackID uint) (*domain.FeedbackEntity, error) {
	var entity domain.FeedbackEntity
	err := Repo.DB.Where("id = ?", feedbackID).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) FindReplyByFeedbackID(feedbackID uint) (*domain.FeedbackReplyEntity, error) {
	var entity domain.FeedbackReplyEntity
	err := Repo.DB.Where("message_id = ?", feedbackID).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (Repo *Repository) SaveFeedbackReply(entity *domain.FeedbackReplyEntity) error {
	if entity.ID == 0 {
		return Repo.DB.Create(&entity).Error
	}
	return Repo.DB.Save(&entity).Error
}

func (Repo *Repository) GetUnreadFeedback() (*[]domain.FeedbackEntity, error) {
	var entity []domain.FeedbackEntity
	err := Repo.DB.Preload("Category").Select("ID", "Name", "CreatedAt", "Title", "UUID", "Category_ID").Where("have_read = ?", false).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
