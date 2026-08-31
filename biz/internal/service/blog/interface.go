package blog

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type BlogService struct {
	Repo   repo.IRepository
	Config *config.Config
}

type IBlogService interface {
	ListPublicPosts(ctx context.Context, c *app.RequestContext)
	GetPublicPost(ctx context.Context, c *app.RequestContext)
	ListAdminPosts(ctx context.Context, c *app.RequestContext)
	CreatePost(ctx context.Context, c *app.RequestContext)
	SaveDraft(ctx context.Context, c *app.RequestContext)
	PublishPost(ctx context.Context, c *app.RequestContext)
	UnpublishPost(ctx context.Context, c *app.RequestContext)
	ArchivePost(ctx context.Context, c *app.RequestContext)
	SetPostPinned(ctx context.Context, c *app.RequestContext)
	DeletePost(ctx context.Context, c *app.RequestContext)
	ListRevisions(ctx context.Context, c *app.RequestContext)
	GetRevision(ctx context.Context, c *app.RequestContext)
	RestoreRevision(ctx context.Context, c *app.RequestContext)
	SignAssetUpload(ctx context.Context, c *app.RequestContext)
	ConfirmAssetUpload(ctx context.Context, c *app.RequestContext)
	ServeAsset(ctx context.Context, c *app.RequestContext)
	TrackView(ctx context.Context, c *app.RequestContext)
	GetLikeState(ctx context.Context, c *app.RequestContext)
	ToggleLike(ctx context.Context, c *app.RequestContext)
	ListPublicComments(ctx context.Context, c *app.RequestContext)
	CreateComment(ctx context.Context, c *app.RequestContext)
	ListAdminComments(ctx context.Context, c *app.RequestContext)
	ReviewComment(ctx context.Context, c *app.RequestContext)
}

func NewBlogService(repo repo.IRepository, cfg *config.Config) IBlogService {
	service := &BlogService{Repo: repo, Config: cfg}
	service.importLegacyPosts()
	return service
}
