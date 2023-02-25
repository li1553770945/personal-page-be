package repo

import (
	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) FindFile(fileKey string) (*domain.FileEntity, error) {
	var file domain.FileEntity
	err := Repo.DB.Where("file_key = ?", fileKey).Limit(1).Find(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
func (Repo *Repository) RemoveFile(fileID int) error {
	err := Repo.DB.Delete(&domain.FileEntity{}, fileID).Error
	if err != nil {
		return err
	}
	return nil
}

func (Repo *Repository) SaveFile(file *domain.FileEntity) error {
	if file.ID == 0 {
		err := Repo.DB.Create(&file).Error
		return err
	} else {
		err := Repo.DB.Save(&file).Error
		return err
	}
}
