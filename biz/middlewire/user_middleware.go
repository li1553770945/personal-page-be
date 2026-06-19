package middlewire

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hertz-contrib/sessions"
)

var jwtSecret = "secret"

func InitAuth(secret string) {
	if secret != "" {
		jwtSecret = secret
	}
}

func UserMiddleware() []app.HandlerFunc {
	return []app.HandlerFunc{func(ctx context.Context, c *app.RequestContext) {
		if username, userID, ok := userFromBearer(c); ok {
			session := sessions.Default(c)
			session.Set("username", username)
			_ = session.Save()
			ctx = context.WithValue(ctx, "username", username)
			if userID != 0 {
				ctx = context.WithValue(ctx, "userId", userID)
			}
			c.Next(ctx)
			return
		}

		session := sessions.Default(c)
		v := session.Get("username")
		if v == nil {
			c.JSON(200, utils.H{"code": 4003, "message": "未登录", "msg": "未登录"})
			c.Abort()
			return
		}
		ctx = context.WithValue(ctx, "username", v)
		c.Next(ctx)
	}}
}

func userFromBearer(c *app.RequestContext) (string, uint, bool) {
	header := string(c.GetHeader("Authorization"))
	if header == "" {
		return "", 0, false
	}
	tokenText := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
	if tokenText == "" || tokenText == header {
		return "", 0, false
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return "", 0, false
	}

	username, _ := claims["username"].(string)
	userID := uint(0)
	switch v := claims["userId"].(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	}
	if username == "" {
		return "", userID, false
	}
	return username, userID, true
}
