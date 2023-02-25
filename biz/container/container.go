package container

import (
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/service/file"
	"personal-page-be/biz/internal/service/user"
)

type Container struct {
	Config      *config.Config
	UserService user.IUserService
	FileService file.IFileService
}

func NewContainer(config *config.Config, userService user.IUserService, fileService file.IFileService) *Container {
	return &Container{
		Config:      config,
		UserService: userService,
		FileService: fileService,
	}

}
