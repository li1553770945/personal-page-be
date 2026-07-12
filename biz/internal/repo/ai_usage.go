package repo

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"personal-page-be/biz/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errAIUsageDailyQuotaExceeded = errors.New("AI usage daily quota exceeded")

func (Repo *Repository) SaveAIUsage(usage *domain.AIUsageEntity) error {
	return Repo.DB.Create(usage).Error
}

func (Repo *Repository) ReserveAIUsageDailyQuota(quotaDay string, identityKey string, ipKey string, limit int) (bool, error) {
	if quotaDay == "" || identityKey == "" || ipKey == "" || limit <= 0 {
		return false, errors.New("invalid AI usage daily quota reservation")
	}
	quotaKeys := dailyQuotaKeys(identityKey, ipKey)
	err := Repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, quotaKey := range quotaKeys {
			row := &domain.AIUsageDailyQuotaEntity{QuotaDay: quotaDay, QuotaKey: quotaKey}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "quota_day"}, {Name: "quota_key"}},
				DoNothing: true,
			}).Create(row).Error; err != nil {
				return err
			}
		}

		for _, quotaKey := range quotaKeys {
			var row domain.AIUsageDailyQuotaEntity
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("quota_day = ? AND quota_key = ?", quotaDay, quotaKey).
				Take(&row).Error; err != nil {
				return err
			}
			if row.RequestCount >= int64(limit) {
				return errAIUsageDailyQuotaExceeded
			}
		}

		for _, quotaKey := range quotaKeys {
			result := tx.Model(&domain.AIUsageDailyQuotaEntity{}).
				Where("quota_day = ? AND quota_key = ?", quotaDay, quotaKey).
				UpdateColumn("request_count", gorm.Expr("request_count + ?", 1))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return fmt.Errorf("AI usage quota row update affected %d rows", result.RowsAffected)
			}
		}
		return nil
	})
	if errors.Is(err, errAIUsageDailyQuotaExceeded) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func dailyQuotaKeys(identityKey string, ipKey string) []string {
	keys := []string{"identity:" + identityKey, "ip:" + ipKey}
	sort.Strings(keys)
	if keys[0] == keys[1] {
		return keys[:1]
	}
	return keys
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
