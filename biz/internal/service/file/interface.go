package file

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/repo"
)

type FileService struct {
	Repo   repo.IRepository
	Config *config.Config
	Logger *logrus.Logger
}

type IFileService interface {
	UploadFile(ctx context.Context, c *app.RequestContext)
	DownloadFile(ctx context.Context, c *app.RequestContext)
	DownloadSignedFile(ctx context.Context, c *app.RequestContext)
	FileInfo(ctx context.Context, c *app.RequestContext)
	ListMyFiles(ctx context.Context, c *app.RequestContext)
	ListAllFiles(ctx context.Context, c *app.RequestContext)
	DeleteFile(ctx context.Context, c *app.RequestContext)
}

func NewFileService(repo repo.IRepository, config *config.Config, logger *logrus.Logger) IFileService {
	return &FileService{
		Repo:   repo,
		Config: config,
		Logger: logger,
	}
}
