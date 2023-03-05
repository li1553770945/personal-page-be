package user

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/internal/repo"
)

type UserService struct {
	Repo repo.IRepository
}

type IUserService interface {
	Login(ctx context.Context, c *app.RequestContext)
	Logout(ctx context.Context, c *app.RequestContext)
	GetUserInfo(ctx context.Context, c *app.RequestContext)
	GenerateActivateCode(ctx context.Context, c *app.RequestContext)
	Register(ctx context.Context, c *app.RequestContext)
}

func NewUserService(repo repo.IRepository) IUserService {
	return &UserService{
		Repo: repo,
	}
}
