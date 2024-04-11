package global_service

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/repo"
	"personal-page-be/biz/internal/service/error_type"
)

type GlobalService struct {
	Repo   repo.IRepository
	Logger *logrus.Logger
}

type IGlobalService interface {
	HealthCheck(ctx context.Context, c *app.RequestContext)
	DeleteFile()
	StartCronDeleteFile()
	CheckLogin(c *app.RequestContext, needAdmin bool) (*error_type.ErrorType, *domain.UserEntity)
}

func NewGlobalService(repo repo.IRepository, logger *logrus.Logger) IGlobalService {
	s := &GlobalService{
		Repo:   repo,
		Logger: logger,
	}
	s.StartCronDeleteFile()
	return s
}
