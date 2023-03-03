package global_service

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/internal/repo"
)

type GlobalService struct {
	Repo repo.IRepository
}

type IGlobalService interface {
	HealthCheck(ctx context.Context, c *app.RequestContext)
	DeleteFile()
	StartCronDeleteFile()
}

func NewGlobalService(repo repo.IRepository) IGlobalService {
	s := &GlobalService{
		Repo: repo,
	}
	s.DeleteFile()
	s.StartCronDeleteFile()
	return s
}
