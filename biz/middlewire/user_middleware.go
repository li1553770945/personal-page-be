package middlewire

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hertz-contrib/sessions"
)

func UserMiddleware() []app.HandlerFunc {
	return []app.HandlerFunc{func(ctx context.Context, c *app.RequestContext) {
		session := sessions.Default(c)
		v := session.Get("username")
		if v == nil {
			c.JSON(200, utils.H{"code": 4003, "msg": "您还未登陆，请先登录"})
			c.Abort()
		}
		ctx = context.WithValue(ctx, "username", v)
		c.Next(ctx)
	}}
}
