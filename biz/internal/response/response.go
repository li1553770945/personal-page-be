package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func OK(c *app.RequestContext, data interface{}, message string) {
	if message == "" {
		message = "ok"
	}
	c.JSON(consts.StatusOK, utils.H{
		"code":    0,
		"message": message,
		"msg":     message,
		"data":    data,
	})
}

func Error(c *app.RequestContext, code int, message string) {
	c.JSON(consts.StatusOK, utils.H{
		"code":    code,
		"message": message,
		"msg":     message,
	})
}
