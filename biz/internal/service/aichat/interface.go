package aichat

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type AIChatService struct {
	Config *config.Config
	Logger *logrus.Logger
	Repo   repo.IRepository
}

type IAIChatService interface {
	SendMessage(ctx context.Context, c *app.RequestContext)
	GetMyUsageStats(ctx context.Context, c *app.RequestContext)
	GetAdminUsageStats(ctx context.Context, c *app.RequestContext)
}

func NewAIChatService(repo repo.IRepository, config *config.Config, logger *logrus.Logger) IAIChatService {
	return &AIChatService{
		Config: config,
		Logger: logger,
		Repo:   repo,
	}
}
