package repo

import (
	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) RemoveProject(projectID uint) error {
	err := Repo.DB.Delete(&domain.ProjectEntity{}, projectID).Error
	if err != nil {
		return err
	}
	return nil
}

func (Repo *Repository) SaveProject(project *domain.ProjectEntity) error {
	if project.ID == 0 {
		err := Repo.DB.Create(&project).Error
		return err
	} else {
		err := Repo.DB.Save(&project).Error
		return err
	}
}

func (Repo *Repository) GetProject(projectID uint) (*domain.ProjectEntity, error) {
	var project domain.ProjectEntity
	err := Repo.DB.Where("id = ?", projectID).Limit(1).Find(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}
func (Repo *Repository) GetProjectsNum() (int64, error) {
	var num int64
	err := Repo.DB.Model(&domain.ProjectEntity{}).Count(&num).Error
	if err != nil {
		return 0, err
	}
	return num, nil
}
func (Repo *Repository) GetProjects(start int, end int) (*[]domain.ProjectEntity, error) {
	var projects []domain.ProjectEntity
	err := Repo.DB.Offset(start).Limit(end - start - 1).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return &projects, nil
}
