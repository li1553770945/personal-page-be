package repo

import (
	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) FindUser(username string) (*domain.UserEntity, error) {
	var user domain.UserEntity
	err := Repo.DB.Where("username = ?", username).Limit(1).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (Repo *Repository) SaveUser(user *domain.UserEntity) error {
	if user.ID == 0 {
		err := Repo.DB.Create(&user).Error
		return err
	} else {
		err := Repo.DB.Save(&user).Error
		return err
	}
}
