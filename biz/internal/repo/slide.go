package repo

import "personal-page-be/biz/internal/domain"

func (Repo *Repository) ListSlides() (*[]domain.SlideEntity, error) {
	var slides []domain.SlideEntity
	err := Repo.DB.Order("created_at desc").Find(&slides).Error
	if err != nil {
		return nil, err
	}
	return &slides, nil
}

func (Repo *Repository) FindSlideByID(slideID uint) (*domain.SlideEntity, error) {
	var slide domain.SlideEntity
	err := Repo.DB.Where("id = ?", slideID).Limit(1).Find(&slide).Error
	if err != nil {
		return nil, err
	}
	return &slide, nil
}

func (Repo *Repository) FindSlideBySlug(slug string) (*domain.SlideEntity, error) {
	var slide domain.SlideEntity
	err := Repo.DB.Where("slug = ?", slug).Limit(1).Find(&slide).Error
	if err != nil {
		return nil, err
	}
	return &slide, nil
}

func (Repo *Repository) SaveSlide(slide *domain.SlideEntity) error {
	if slide.ID == 0 {
		return Repo.DB.Create(&slide).Error
	}
	return Repo.DB.Save(&slide).Error
}

func (Repo *Repository) RemoveSlide(slideID uint) error {
	return Repo.DB.Delete(&domain.SlideEntity{}, slideID).Error
}
