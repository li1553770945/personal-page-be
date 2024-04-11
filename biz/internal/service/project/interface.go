package project

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/internal/repo"
	"personal-page-be/biz/internal/service/global_service"
)

type ProjectService struct {
	Repo          repo.IRepository
	GlobalService global_service.IGlobalService
	Logger        *logrus.Logger
}

type IProjectService interface {
	AddProject(ctx context.Context, c *app.RequestContext)
	RemoveProject(ctx context.Context, c *app.RequestContext)
	GetNum(ctx context.Context, c *app.RequestContext)
	GetProjects(ctx context.Context, c *app.RequestContext)
}

func NewProjectService(repo repo.IRepository, globalService global_service.IGlobalService, logger *logrus.Logger) IProjectService {
	return &ProjectService{
		Repo:          repo,
		GlobalService: globalService,
		Logger:        logger,
	}
}
