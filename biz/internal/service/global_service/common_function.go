package global_service

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/sessions"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/service/error_type"
)

func (s *GlobalService) CheckLogin(c *app.RequestContext, needAdmin bool) (*error_type.ErrorType, *domain.UserEntity) {
	session := sessions.Default(c)
	v := session.Get("username")
	if v == nil {
		return error_type.ErrNotLogin, nil
	}
	user, err := s.Repo.FindUser(v.(string))
	if err != nil {
		return error_type.ErrNotLogin, nil
	}

	if needAdmin && user.Role != "admin" {
		return error_type.ErrNotPermitted, nil
	}
	return nil, user
}
