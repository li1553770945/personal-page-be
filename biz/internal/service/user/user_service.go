package user

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/sessions"
	"golang.org/x/crypto/bcrypt"
	"personal-page-be/biz/internal/assembler"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/repo"
)

type UserService struct {
	Repo repo.IRepository
}

func (s *UserService) Login(ctx context.Context, c *app.RequestContext) {
	var user domain.UserEntity
	c.Bind(&user)
	findUser, err := s.Repo.FindUser(user.Username)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(user.Password))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户名或密码错误",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Save()

	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "登陆成功",
		"data": assembler.UserEntityToDTO(findUser),
	})
}

func (*UserService) Logout(ctx context.Context, c *app.RequestContext) {
	session := sessions.Default(c)
	session.Delete("username")
	session.Save()
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "登出成功",
	})
}

func (s *UserService) GetUserInfo(ctx context.Context, c *app.RequestContext) {
	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"data": assembler.UserEntityToDTO(user),
	})
}
