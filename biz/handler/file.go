package handler

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"net/url"
)

func UploadFile(ctx context.Context, c *app.RequestContext) {
	// single file
	file, _ := c.FormFile("file")
	fmt.Println(file.Filename)

	// Upload the file to specific dst
	err := c.SaveUploadedFile(file, fmt.Sprintf("./file/upload/%s", file.Filename))
	if err != nil {
		c.JSON(200, utils.H{"code": 2001, "msg": err.Error()})
		return
	}

	c.JSON(200, utils.H{"code": 0, "msg": "上传成功"})
}
func FileAttachment(ctx context.Context, c *app.RequestContext) {
	fileName := url.QueryEscape("hertz")
	c.FileAttachment("./file/file.txt", fileName)
}

func DownloadFile(ctx context.Context, c *app.RequestContext) {
	c.File("./file/file.txt")
}
