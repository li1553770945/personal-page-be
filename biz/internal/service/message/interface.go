package message

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type MessageService struct {
	Repo   repo.IRepository
	Config *config.Config
}

type IMessageService interface {
	FindAllMessageCategories(ctx context.Context, c *app.RequestContext)
	SaveMessage(ctx context.Context, c *app.RequestContext)
	AddReply(ctx context.Context, c *app.RequestContext)
	GetReply(ctx context.Context, c *app.RequestContext)
	GetMessages(ctx context.Context, c *app.RequestContext)
}

func NewMessageService(repo repo.IRepository, config_ *config.Config) IMessageService {
	s := &MessageService{
		Repo:   repo,
		Config: config_,
	}
	return s
}
