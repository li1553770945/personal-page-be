package repo

import (
	"personal-page-be/biz/internal/domain"

	"gorm.io/gorm"
)

func (Repo *Repository) FindUser(username string) (*domain.UserEntity, error) {
	var user domain.UserEntity
	err := Repo.DB.Where("username = ?", username).Limit(1).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (Repo *Repository) FindUserByID(userID uint) (*domain.UserEntity, error) {
	var user domain.UserEntity
	err := Repo.DB.Where("id = ?", userID).Limit(1).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (Repo *Repository) FindFirstUserByRole(role string) (*domain.UserEntity, error) {
	var user domain.UserEntity
	err := Repo.DB.Where("role = ?", role).Order("id asc").Limit(1).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (Repo *Repository) ListUsers() (*[]domain.UserEntity, error) {
	var users []domain.UserEntity
	err := Repo.DB.Order("id asc").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (Repo *Repository) CountUsersByRole(role string, canUseOnly bool) (int64, error) {
	query := Repo.DB.Model(&domain.UserEntity{}).Where("role = ?", role)
	if canUseOnly {
		query = query.Where("can_use = ?", true)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
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

func (Repo *Repository) SaveUserAndAudit(user *domain.UserEntity, audit *domain.AdminAuditLogEntity) error {
	return Repo.DB.Transaction(func(tx *gorm.DB) error {
		if user.ID == 0 {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}

func (Repo *Repository) RemoveUserAndAudit(userID uint, audit *domain.AdminAuditLogEntity) error {
	return Repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&domain.UserEntity{}, userID).Error; err != nil {
			return err
		}
		if audit != nil {
			return tx.Create(audit).Error
		}
		return nil
	})
}
