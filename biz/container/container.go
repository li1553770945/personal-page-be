package container

import (
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/service/file"
	"personal-page-be/biz/internal/service/global_service"
	"personal-page-be/biz/internal/service/message"
	"personal-page-be/biz/internal/service/user"
)

type Container struct {
	Config         *config.Config
	UserService    user.IUserService
	FileService    file.IFileService
	GlobalService  global_service.IGlobalService
	MessageService message.IMessageService
}

func NewContainer(config *config.Config, userService user.IUserService,
	fileService file.IFileService,
	globalService global_service.IGlobalService,
	messageService message.IMessageService,
) *Container {
	return &Container{
		Config:         config,
		UserService:    userService,
		FileService:    fileService,
		GlobalService:  globalService,
		MessageService: messageService,
	}

}
