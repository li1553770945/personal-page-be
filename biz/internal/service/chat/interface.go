package chat

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type ChatService struct {
	Cache *cache.Cache
	Log   *logrus.Logger
}

type IChatService interface {
	CreateChat(ctx context.Context, c *app.RequestContext)
	JoinChat(ctx context.Context, c *app.RequestContext)
	CloseChat(ctx context.Context, c *app.RequestContext)
}

func NewChatService(cache *cache.Cache, log *logrus.Logger) IChatService {
	s := &ChatService{
		Cache: cache,
		Log:   log,
	}
	return s
}
