package user

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hertz-contrib/sessions"
	"golang.org/x/crypto/bcrypt"
	"personal-page-be/biz/internal/assembler"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
	U "personal-page-be/biz/internal/utils"
)

func (s *UserService) Login(ctx context.Context, c *app.RequestContext) {
	var loginReq domain.UserEntity
	if err := c.BindAndValidate(&loginReq); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}

	findUser, err := s.Repo.FindUser(loginReq.Username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if findUser.ID == 0 {
		response.Error(c, 4003, "用户名或密码错误")
		return
	}
	if !findUser.IsActivate {
		response.Error(c, 4003, "用户未激活，请注册后使用")
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(loginReq.Password)); err != nil {
		response.Error(c, 4003, "用户名或密码错误")
		return
	}
	if !findUser.CanUse {
		response.Error(c, 4003, "账号已被禁用")
		return
	}

	session := sessions.Default(c)
	session.Set("username", findUser.Username)
	_ = session.Save()

	token, err := s.issueToken(findUser)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, utils.H{
		"token": token,
		"user":  assembler.UserEntityToDTO(findUser),
	}, "登录成功")
}

func (*UserService) Logout(ctx context.Context, c *app.RequestContext) {
	session := sessions.Default(c)
	session.Delete("username")
	_ = session.Save()
	response.OK(c, nil, "登出成功")
}

func (s *UserService) GetUserInfo(ctx context.Context, c *app.RequestContext) {
	user, err := s.currentUser(ctx)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if user.ID == 0 {
		response.Error(c, 4004, "用户不存在")
		return
	}
	response.OK(c, assembler.UserEntityToDTO(user), "ok")
}

func (s *UserService) GenerateActivateCode(ctx context.Context, c *app.RequestContext) {
	current, err := s.currentUser(ctx)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if current.Role != "admin" {
		response.Error(c, 4003, "无权执行此操作")
		return
	}

	var req dto.GenerateActivateCodeReq
	if err = c.BindAndValidate(&req); err != nil {
		response.Error(c, 5001, "username 必须大于 5 位: "+err.Error())
		return
	}

	registerUser, err := s.Repo.FindUser(req.Username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if registerUser.ID != 0 && registerUser.IsActivate {
		response.Error(c, 4003, "用户名已被注册")
		return
	}
	if registerUser.ID == 0 {
		registerUser = &domain.UserEntity{
			Username:   req.Username,
			IsActivate: false,
			CanUse:     true,
		}
	}
	registerUser.ActivateCode = U.RandSeq(10)

	if err = s.Repo.SaveUser(registerUser); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	response.OK(c, utils.H{
		"activate_code": registerUser.ActivateCode,
		"activeCode":    registerUser.ActivateCode,
	}, "ok")
}

func (s *UserService) Register(ctx context.Context, c *app.RequestContext) {
	var req dto.RegisterReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}

	activeCode := req.ActivateCode
	if activeCode == "" {
		activeCode = req.ActiveCode
	}
	findUser, err := s.Repo.FindUser(req.Username)
	if err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	if findUser.ID == 0 {
		response.Error(c, 4003, "请联系管理员获得激活码")
		return
	}
	if findUser.ActivateCode != activeCode {
		response.Error(c, 4003, "激活码错误")
		return
	}
	if findUser.IsActivate {
		response.Error(c, 4003, "用户已经激活")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 5001, "密码加密失败: "+err.Error())
		return
	}
	findUser.Password = string(hash)
	findUser.Nickname = req.Nickname
	findUser.IsActivate = true
	findUser.CanUse = true

	if err = s.Repo.SaveUser(findUser); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	response.OK(c, nil, "注册成功")
}

func (s *UserService) currentUser(ctx context.Context) (*domain.UserEntity, error) {
	if username, ok := ctx.Value("username").(string); ok && username != "" {
		return s.Repo.FindUser(username)
	}
	if userID, ok := ctx.Value("userId").(uint); ok && userID != 0 {
		return s.Repo.FindUserByID(userID)
	}
	return &domain.UserEntity{}, nil
}

func (s *UserService) issueToken(user *domain.UserEntity) (string, error) {
	claims := jwt.MapClaims{
		"userId":   user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(14 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.Config.EffectiveJWTKey()))
}
