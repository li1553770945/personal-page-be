package container

import "personal-page-be/biz/infra/config"

type Container struct {
	Config *config.Config
}

func NewContainer(config *config.Config) *Container {
	return &Container{
		Config: config,
	}

}
