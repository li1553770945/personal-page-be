package aichat

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/repo"
)

type aiChatRepository interface {
	FindUser(username string) (*domain.UserEntity, error)
	FindUserByID(userID uint) (*domain.UserEntity, error)
	SaveAIUsage(usage *domain.AIUsageEntity) error
	ReserveAIUsageDailyQuota(quotaDay string, identityKey string, ipKey string, limit int) (bool, error)
	ListAIUsage(userID *uint, startAt *time.Time, endAt *time.Time, model string, channel string) (*[]domain.AIUsageEntity, error)
}

type AIChatService struct {
	Config *config.Config
	Logger *logrus.Logger
	Repo   aiChatRepository
	guard  *requestGuard
}

type IAIChatService interface {
	SendMessage(ctx context.Context, c *app.RequestContext)
	GetMyUsageStats(ctx context.Context, c *app.RequestContext)
	GetAdminUsageStats(ctx context.Context, c *app.RequestContext)
}

func NewAIChatService(repo repo.IRepository, config *config.Config, logger *logrus.Logger) IAIChatService {
	limits, err := config.EffectiveAIChatLimits()
	if err != nil {
		panic("invalid AI chat limits: " + err.Error())
	}
	return &AIChatService{
		Config: config,
		Logger: logger,
		Repo:   repo,
		guard:  newRequestGuard(limits),
	}
}
