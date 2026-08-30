package chat

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/internal/repo"
)

type roomClient struct {
	id      string
	token   string
	conn    *websocket.Conn
	writeMu sync.Mutex
}

type roomState struct {
	id        string
	expiresAt time.Time
	clients   map[string]*roomClient
	mu        sync.Mutex
}

type ChatService struct {
	Repo    repo.IRepository
	Log     *logrus.Logger
	rooms   map[string]*roomState
	roomsMu sync.RWMutex
}

type IChatService interface {
	CreateRoom(ctx context.Context, c *app.RequestContext)
	JoinRoom(ctx context.Context, c *app.RequestContext)
	Connect(ctx context.Context, c *app.RequestContext)
}

func NewChatService(repo repo.IRepository, log *logrus.Logger) IChatService {
	service := &ChatService{Repo: repo, Log: log, rooms: map[string]*roomState{}}
	go service.cleanupExpiredRooms()
	return service
}
