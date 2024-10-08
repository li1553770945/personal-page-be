package message

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type MessageService struct {
	Repo   repo.IRepository
	Config *config.Config
	Logger *logrus.Logger
}

type IMessageService interface {
	FindAllMessageCategories(ctx context.Context, c *app.RequestContext)
	SaveMessage(ctx context.Context, c *app.RequestContext)
	AddReply(ctx context.Context, c *app.RequestContext)
	GetReply(ctx context.Context, c *app.RequestContext)
	GetMessages(ctx context.Context, c *app.RequestContext)
}

func NewMessageService(repo repo.IRepository, config *config.Config, logger *logrus.Logger) IMessageService {
	s := &MessageService{
		Repo:   repo,
		Config: config,
		Logger: logger,
	}
	return s
}
