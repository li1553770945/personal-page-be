package repo

import (
	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) FindFileByFileKey(fileKey string) (*domain.FileEntity, error) {
	var file domain.FileEntity
	err := Repo.DB.Where("file_key = ?", fileKey).Limit(1).Find(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (Repo *Repository) FindFileByID(fileID uint) (*domain.FileEntity, error) {
	var file domain.FileEntity
	err := Repo.DB.Where("id = ?", fileID).Limit(1).Find(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
func (Repo *Repository) RemoveFile(fileID uint) error {
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

func (Repo *Repository) FindFileBySaveName(saveName string) (*domain.FileEntity, error) {
	var file domain.FileEntity
	err := Repo.DB.Where("save_name = ?", saveName).Limit(1).Find(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
