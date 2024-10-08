// Code generated by hertz generator.

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"personal-page-be/biz/middlewire"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	api := r.Group("/api")

	userApi := api.Group("/users")

	userApi.POST("/login", App.UserService.Login)
	userApi.GET("/logout", App.UserService.Logout)
	userApi.GET("/me", append(middlewire.UserMiddleware(), App.UserService.GetUserInfo)...)
	userApi.POST("/activate-code", append(middlewire.UserMiddleware(), App.UserService.GenerateActivateCode)...)
	userApi.POST("/register", App.UserService.Register)

	fileApi := api.Group("/files")

	fileApi.POST("", append(middlewire.UserMiddleware(), App.FileService.UploadFile)...)
	fileApi.GET("", App.FileService.DownloadFile)
	fileApi.GET("/info", App.FileService.FileInfo)
	fileApi.DELETE("/:id", append(middlewire.UserMiddleware(), App.FileService.DeleteFile)...)

	messageApi := api.Group("/messages")
	messageApi.GET("/categories", App.MessageService.FindAllMessageCategories)
	messageApi.POST("", App.MessageService.SaveMessage)
	messageApi.GET("/reply", App.MessageService.GetReply)
	messageApi.POST("/reply", append(middlewire.UserMiddleware(), App.MessageService.AddReply)...)
	messageApi.GET("", App.MessageService.GetMessages)

	projectsApi := api.Group("/projects")
	projectsApi.GET("/num", App.ProjectService.GetNum)
	projectsApi.GET("", App.ProjectService.GetProjects)
	projectsApi.POST("", append(middlewire.UserMiddleware(), App.ProjectService.AddProject)...)
	projectsApi.DELETE("/:id", append(middlewire.UserMiddleware(), App.ProjectService.RemoveProject)...)

	socket := r.Group("/socket")
	{
		socket.GET("/new-chat", App.ChatService.CreateChat)
		socket.GET("/join-chat", App.ChatService.JoinChat)
	}

}
