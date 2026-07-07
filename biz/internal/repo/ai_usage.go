package repo

import (
	"time"

	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) SaveAIUsage(usage *domain.AIUsageEntity) error {
	return Repo.DB.Create(usage).Error
}

func (Repo *Repository) ListAIUsage(userID *uint, startAt *time.Time, endAt *time.Time, model string, channel string) (*[]domain.AIUsageEntity, error) {
	var rows []domain.AIUsageEntity
	query := Repo.DB.Order("created_at asc")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if startAt != nil {
		query = query.Where("created_at >= ?", *startAt)
	}
	if endAt != nil {
		query = query.Where("created_at < ?", *endAt)
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return &rows, nil
}
