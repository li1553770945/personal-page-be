//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/infra/log"
)

func GetContainer(path string) *Container {
	panic(wire.Build(

		//infra
		config.InitConfig,
		log.NewLogger,

		NewContainer,
	))
}
