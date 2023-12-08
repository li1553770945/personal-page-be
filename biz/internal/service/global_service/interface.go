package global_service

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/internal/repo"
)

type GlobalService struct {
	Repo   repo.IRepository
	Logger *logrus.Logger
}

type IGlobalService interface {
	HealthCheck(ctx context.Context, c *app.RequestContext)
	DeleteFile()
	StartCronDeleteFile()
}

func NewGlobalService(repo repo.IRepository, logger *logrus.Logger) IGlobalService {
	s := &GlobalService{
		Repo:   repo,
		Logger: logger,
	}
	s.StartCronDeleteFile()
	return s
}
