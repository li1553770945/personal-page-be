//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/infra/database"
	"personal-page-be/biz/internal/repo"
	"personal-page-be/biz/internal/service/file"
	"personal-page-be/biz/internal/service/global_service"
	"personal-page-be/biz/internal/service/message"
	"personal-page-be/biz/internal/service/user"
)

func GetContainer(path string) *Container {
	panic(wire.Build(

		//infra
		config.InitConfig,
		database.NewDatabase,

		//repo
		repo.NewRepository,

		//service
		user.NewUserService,
		file.NewFileService,
		global_service.NewGlobalService,
		message.NewMessageService,

		NewContainer,
	))
}
