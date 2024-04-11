package file

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"personal-page-be/biz/internal/repo"
)

type FileService struct {
	Repo   repo.IRepository
	Logger *logrus.Logger
}

type IFileService interface {
	UploadFile(ctx context.Context, c *app.RequestContext)
	DownloadFile(ctx context.Context, c *app.RequestContext)
	FileInfo(ctx context.Context, c *app.RequestContext)
	DeleteFile(ctx context.Context, c *app.RequestContext)
}

func NewFileService(repo repo.IRepository, logger *logrus.Logger) IFileService {
	return &FileService{
		Repo:   repo,
		Logger: logger,
	}
}
