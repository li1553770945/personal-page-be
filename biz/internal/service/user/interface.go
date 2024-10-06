package user

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type UserService struct {
	Repo   repo.IRepository
	Config *config.Config
}

type IUserService interface {
	Login(ctx context.Context, c *app.RequestContext)
	Logout(ctx context.Context, c *app.RequestContext)
	GetUserInfo(ctx context.Context, c *app.RequestContext)
	GenerateActivateCode(ctx context.Context, c *app.RequestContext)
	Register(ctx context.Context, c *app.RequestContext)
}

func NewUserService(repo repo.IRepository, config *config.Config) IUserService {
	return &UserService{
		Repo:   repo,
		Config: config,
	}
}
