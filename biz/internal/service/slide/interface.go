package slide

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/internal/repo"
)

type SlideService struct {
	Repo repo.IRepository
}

type ISlideService interface {
	ListPublicSlides(ctx context.Context, c *app.RequestContext)
	ListAdminSlides(ctx context.Context, c *app.RequestContext)
	CreateSlide(ctx context.Context, c *app.RequestContext)
	UpdateSlide(ctx context.Context, c *app.RequestContext)
	DeleteSlide(ctx context.Context, c *app.RequestContext)
	UnlockSlide(ctx context.Context, c *app.RequestContext)
}

func NewSlideService(repo repo.IRepository) ISlideService {
	return &SlideService{Repo: repo}
}
