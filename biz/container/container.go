package container

import (
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/internal/service/user"
)

type Container struct {
	Config      *config.Config
	UserService user.IUserService
}

func NewContainer(config *config.Config, userService user.IUserService) *Container {
	return &Container{
		Config:      config,
		UserService: userService,
	}

}
