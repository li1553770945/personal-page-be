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
	"personal-page-be/biz/internal/dto"
	U "personal-page-be/biz/internal/utils"
)

func (s *UserService) Login(ctx context.Context, c *app.RequestContext) {
	var user domain.UserEntity
	err := c.BindAndValidate(&user)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	findUser, err := s.Repo.FindUser(user.Username)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if !findUser.IsActivate {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户未注册激活，请注册后使用",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(user.Password))
	if findUser.ID == 0 || err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户名或密码错误",
		})
		return
	}

	if findUser.CanUse == false {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "抱歉，您的账户已被禁用",
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

func (s *UserService) GenTencentIMUserSig(ctx context.Context, c *app.RequestContext) {
	var user domain.UserEntity
	err := c.BindAndValidate(&user)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	findUser, err := s.Repo.FindUser(user.Username)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if !findUser.IsActivate {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户未注册激活，请注册后使用",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(user.Password))
	if findUser.ID == 0 || err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户名或密码错误",
		})
		return
	}

	if findUser.CanUse == false {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "抱歉，您的账户已被禁用",
		})
		return
	}
	tencentConfig := s.Config.TencentConfig
	userSigStr, err := GenUserSig(tencentConfig.APPID, tencentConfig.SecretKey, findUser.Username, 24*60*60*180)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "获取UserSig调用API失败",
		})
		return
	}
	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "获取成功",
		"data": userSigStr,
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
	return
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

func (s *UserService) GenerateActivateCode(ctx context.Context, c *app.RequestContext) {
	username := ctx.Value("username")
	user, err := s.Repo.FindUser(username.(string))
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if user.Role != "admin" {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "您无权执行此操作",
		})
		return
	}

	var req dto.GenerateActivateCodeReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "username必须大于5位" + err.Error(),
		})
		return
	}

	registerUsername := req.Username
	findUser, err := s.Repo.FindUser(registerUsername)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  err.Error(),
		})
		return
	}
	if findUser.ID != 0 && findUser.IsActivate == true {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户名已被注册",
		})
		return
	}

	var RegisterUser *domain.UserEntity
	if findUser.ID != 0 {
		RegisterUser = findUser
		RegisterUser.ActivateCode = U.RandSeq(10)
	} else {
		RegisterUser = &domain.UserEntity{
			Username:     registerUsername,
			IsActivate:   false,
			CanUse:       true,
			ActivateCode: U.RandSeq(10),
		}
	}

	err = s.Repo.SaveUser(RegisterUser)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": "5001",
			"msg":  "操作失败:" + err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"code": "0",
		"data": utils.H{
			"activate_code": RegisterUser.ActivateCode,
		},
	})
	return
}

func (s *UserService) Register(ctx context.Context, c *app.RequestContext) {
	var req dto.RegisterReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "操作失败：" + err.Error(),
		})
		return
	}

	findUser, err := s.Repo.FindUser(req.Username)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "操作失败：" + err.Error(),
		})
		return
	}
	if findUser.ID == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "请联系管理员获得激活码",
		})
		return
	}

	if findUser.ActivateCode != req.ActivateCode {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "激活码错误",
		})
		return
	}

	if findUser.IsActivate == true {
		c.JSON(consts.StatusOK, utils.H{
			"code": 4003,
			"msg":  "用户已经激活",
		})
		return
	}

	findUser.IsActivate = true

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) //加密处理
	encodePWD := string(hash)                                                          // 保存在数据库的密码，虽然每次生成都不同，只需保存一份即可
	findUser.Password = encodePWD
	findUser.Nickname = req.Nickname
	err = s.Repo.SaveUser(findUser)
	if err != nil {
		c.JSON(consts.StatusOK, utils.H{
			"code": 5001,
			"msg":  "操作失败:" + err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"code": 0,
		"msg":  "注册成功",
	})

}
