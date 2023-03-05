// Code generated by hertz generator.

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"personal-page-be/biz/middlewire"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	api := r.Group("/api")
	{
		api.POST("/login", App.UserService.Login)
		api.GET("/logout", App.UserService.Logout)
		api.GET("/user-info", append(middlewire.UserMiddleware(), App.UserService.GetUserInfo)...)
		api.POST("/generate-activate-code", append(middlewire.UserMiddleware(), App.UserService.GenerateActivateCode)...)
		api.POST("/register", App.UserService.Register)

		api.POST("/upload-file", append(middlewire.UserMiddleware(), App.FileService.UploadFile)...)
		api.GET("/download-file", App.FileService.DownloadFile)
		api.GET("/file-info", App.FileService.FileInfo)
		api.DELETE("/delete-file", append(middlewire.UserMiddleware(), App.FileService.DeleteFile)...)

		api.GET("/all-message-categories", App.MessageService.FindAllMessageCategories)
		api.POST("/message", App.MessageService.SaveMessage)
		api.GET("/reply", App.MessageService.GetReply)
		api.POST("/reply", append(middlewire.UserMiddleware(), App.MessageService.AddReply)...)
	}

}
