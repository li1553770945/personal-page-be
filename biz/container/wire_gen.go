// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package container

import (
	"personal-page-be/biz/infra/config"
	"personal-page-be/biz/infra/database"
	"personal-page-be/biz/internal/repo"
	"personal-page-be/biz/internal/service/user"
)

// Injectors from wire.go:

func GetContainer(path string) *Container {
	configConfig := config.InitConfig(path)
	db := database.NewDatabase(configConfig)
	iRepository := repo.NewRepository(db)
	iUserService := user.NewUserService(iRepository)
	container := NewContainer(configConfig, iUserService)
	return container
}
