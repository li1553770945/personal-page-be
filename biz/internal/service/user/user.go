package user

import (
	"context"
	"fmt"
	"strconv"
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
	if normalizedRole := domain.NormalizeRole(findUser.Role); normalizedRole != findUser.Role {
		findUser.Role = normalizedRole
		_ = s.Repo.SaveUser(findUser)
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
		response.Error(c, 4003, "账号已被封禁")
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
	if !domain.IsAdminRole(current.Role) {
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
			Role:       domain.RoleUser,
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

	findUser, err := s.Repo.FindUser(req.Username)
	if err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	if findUser.ID != 0 && findUser.IsActivate {
		response.Error(c, 4003, "用户已经注册")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 5001, "密码加密失败: "+err.Error())
		return
	}
	if findUser.ID == 0 {
		findUser = &domain.UserEntity{
			Username: req.Username,
			Role:     domain.RoleUser,
		}
	}
	findUser.Password = string(hash)
	findUser.Nickname = req.Nickname
	findUser.IsActivate = true
	findUser.CanUse = true
	findUser.Role = domain.NormalizeRole(findUser.Role)
	findUser.ActivateCode = ""

	if err = s.Repo.SaveUser(findUser); err != nil {
		response.Error(c, 5001, "操作失败: "+err.Error())
		return
	}
	response.OK(c, nil, "注册成功")
}

func (s *UserService) ListUsers(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	users, err := s.Repo.ListUsers()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, assembler.UserEntitiesToAdminDTO(users), "ok")
}

func (s *UserService) UpdateUserRole(ctx context.Context, c *app.RequestContext) {
	current, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}

	userID, err := parseIDParam(c.Param("id"))
	if err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	var req dto.UpdateUserRoleReq
	if err = c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	role := domain.NormalizeRole(req.Role)
	if role != req.Role {
		response.Error(c, 2001, "不支持的用户角色")
		return
	}

	target, err := s.Repo.FindUserByID(userID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if target.ID == 0 {
		response.Error(c, 4004, "用户不存在")
		return
	}
	if target.ID == current.ID && role != domain.RoleSuperAdmin {
		if err = s.ensureCanRemoveSuperAdmin(target); err != nil {
			response.Error(c, 4003, err.Error())
			return
		}
	}
	if target.Role == domain.RoleSuperAdmin && role != domain.RoleSuperAdmin {
		if err = s.ensureCanRemoveSuperAdmin(target); err != nil {
			response.Error(c, 4003, err.Error())
			return
		}
	}

	target.Role = role
	if err = s.Repo.SaveUser(target); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, assembler.UserEntityToAdminDTO(target), "ok")
}

func (s *UserService) UpdateUserStatus(ctx context.Context, c *app.RequestContext) {
	current, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}

	userID, err := parseIDParam(c.Param("id"))
	if err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	var req dto.UpdateUserStatusReq
	if err = c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}

	target, err := s.Repo.FindUserByID(userID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if target.ID == 0 {
		response.Error(c, 4004, "用户不存在")
		return
	}
	if target.ID == current.ID && !req.CanUse {
		response.Error(c, 4003, "不能封禁当前登录的超级管理员")
		return
	}
	if target.Role == domain.RoleSuperAdmin && target.CanUse && !req.CanUse {
		if err = s.ensureCanRemoveSuperAdmin(target); err != nil {
			response.Error(c, 4003, err.Error())
			return
		}
	}

	target.CanUse = req.CanUse
	if err = s.Repo.SaveUser(target); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, assembler.UserEntityToAdminDTO(target), "ok")
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
		"role":     domain.NormalizeRole(user.Role),
		"exp":      time.Now().Add(14 * 24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.Config.EffectiveJWTKey()))
}

func (s *UserService) requireSuperAdmin(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
	current, err := s.currentUser(ctx)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return nil, false
	}
	if current.ID == 0 || !current.CanUse || !domain.IsSuperAdminRole(current.Role) {
		response.Error(c, 4003, "无权执行此操作")
		return nil, false
	}
	return current, true
}

func (s *UserService) ensureCanRemoveSuperAdmin(target *domain.UserEntity) error {
	if target.Role != domain.RoleSuperAdmin {
		return nil
	}
	count, err := s.Repo.CountUsersByRole(domain.RoleSuperAdmin, true)
	if err != nil {
		return err
	}
	if target.CanUse && count <= 1 {
		return fmt.Errorf("至少需要保留一个可用的超级管理员")
	}
	return nil
}

func (s *UserService) ensureSuperAdminBootstrap() error {
	count, err := s.Repo.CountUsersByRole(domain.RoleSuperAdmin, false)
	if err != nil || count > 0 {
		return err
	}

	username := s.Config.EffectiveSuperAdminUsername()
	var user *domain.UserEntity
	if username != "" {
		user, err = s.Repo.FindUser(username)
		if err != nil {
			return err
		}
	}
	if user == nil || user.ID == 0 {
		user, err = s.Repo.FindFirstUserByRole(domain.RoleAdmin)
		if err != nil {
			return err
		}
	}
	if user == nil || user.ID == 0 {
		return nil
	}
	user.Role = domain.RoleSuperAdmin
	if !user.IsActivate {
		user.IsActivate = true
	}
	if !user.CanUse {
		user.CanUse = true
	}
	return s.Repo.SaveUser(user)
}

func parseIDParam(id string) (uint, error) {
	value, err := strconv.ParseUint(id, 10, 64)
	return uint(value), err
}
