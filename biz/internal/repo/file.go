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

func (Repo *Repository) FindFileByKey(fileKey string) (*domain.FileEntity, error) {
	var file domain.FileEntity
	err := Repo.DB.Where(&domain.FileEntity{Key: fileKey}).Limit(1).Find(&file).Error
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

func (Repo *Repository) RemoveFileByKey(fileKey string) error {
	return Repo.DB.Where(&domain.FileEntity{Key: fileKey}).Delete(&domain.FileEntity{}).Error
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

func (Repo *Repository) ListFiles(userID *uint) (*[]domain.FileEntity, error) {
	var files []domain.FileEntity
	query := Repo.DB.Preload("User").Order("created_at desc")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Find(&files).Error
	if err != nil {
		return nil, err
	}
	return &files, nil
}

func (Repo *Repository) CountFilesByUserID(userID uint) (int64, error) {
	var count int64
	err := Repo.DB.Model(&domain.FileEntity{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
