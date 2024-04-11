package project

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type ProjectService struct {
	Repo   repo.IRepository
	Config *config.Config
}

type IProjectService interface {
	AddProject(ctx context.Context, c *app.RequestContext)
	RemoveProject(ctx context.Context, c *app.RequestContext)
	GetPages(ctx context.Context, c *app.RequestContext)
	GetProjects(ctx context.Context, c *app.RequestContext)
}

func NewProjectService(repo repo.IRepository, config *config.Config) IProjectService {
	return &ProjectService{
		Repo:   repo,
		Config: config,
	}
}
