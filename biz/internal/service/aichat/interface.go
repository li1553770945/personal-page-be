package aichat

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/infra/config"
)

type AIChatService struct {
	Config *config.Config
	Logger *logrus.Logger
}

type IAIChatService interface {
	SendMessage(ctx context.Context, c *app.RequestContext)
}

func NewAIChatService(config *config.Config, logger *logrus.Logger) IAIChatService {
	return &AIChatService{
		Config: config,
		Logger: logger,
	}
}
