package slide

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/crypto/bcrypt"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
)

var slideSlugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func (s *SlideService) ListPublicSlides(ctx context.Context, c *app.RequestContext) {
	slides, err := s.Repo.ListSlides()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, slideEntitiesToPublicDTO(slides), "ok")
}

func (s *SlideService) ListAdminSlides(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	slides, err := s.Repo.ListSlides()
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, slideEntitiesToDTO(slides), "ok")
}

func (s *SlideService) CreateSlide(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	var req dto.SaveSlideReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if err := validateSlideReq(&req); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	exists, err := s.Repo.FindSlideBySlug(req.Slug)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if exists.ID != 0 {
		response.Error(c, 5002, "幻灯片 id 已存在")
		return
	}

	entity := &domain.SlideEntity{}
	if err := applySlideReq(entity, &req, true); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	if err := s.Repo.SaveSlide(entity); err != nil {
		response.Error(c, 5001, "保存幻灯片失败: "+err.Error())
		return
	}
	response.OK(c, slideEntityToDTO(entity), "ok")
}

func (s *SlideService) UpdateSlide(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	slideID, err := parseSlideID(c.Param("id"))
	if err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	entity, err := s.Repo.FindSlideByID(slideID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到幻灯片")
		return
	}

	var req dto.SaveSlideReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if err := validateSlideReq(&req); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	if req.Slug != entity.Slug {
		exists, err := s.Repo.FindSlideBySlug(req.Slug)
		if err != nil {
			response.Error(c, 5001, err.Error())
			return
		}
		if exists.ID != 0 && exists.ID != entity.ID {
			response.Error(c, 5002, "幻灯片 id 已存在")
			return
		}
	}
	if err := applySlideReq(entity, &req, false); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	if err := s.Repo.SaveSlide(entity); err != nil {
		response.Error(c, 5001, "保存幻灯片失败: "+err.Error())
		return
	}
	response.OK(c, slideEntityToDTO(entity), "ok")
}

func (s *SlideService) DeleteSlide(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	slideID, err := parseSlideID(c.Param("id"))
	if err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	entity, err := s.Repo.FindSlideByID(slideID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到幻灯片")
		return
	}
	if err := s.Repo.RemoveSlide(slideID); err != nil {
		response.Error(c, 5001, "删除幻灯片失败: "+err.Error())
		return
	}
	response.OK(c, nil, "删除成功")
}

func (s *SlideService) UnlockSlide(ctx context.Context, c *app.RequestContext) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		response.Error(c, 2001, "缺少幻灯片 id")
		return
	}
	entity, err := s.Repo.FindSlideBySlug(slug)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if entity.ID == 0 {
		response.Error(c, 4004, "未找到幻灯片")
		return
	}
	if !entity.Protected {
		response.OK(c, slideEntityToDTO(entity), "ok")
		return
	}

	var req dto.UnlockSlideReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if entity.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(entity.PasswordHash), []byte(req.Password)) != nil {
		response.Error(c, 4003, "密码不正确")
		return
	}
	response.OK(c, slideEntityToDTO(entity), "ok")
}

func (s *SlideService) requireSuperAdmin(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return nil, false
	}
	if user.ID == 0 || !user.CanUse || !domain.IsSuperAdminRole(domain.NormalizeRole(user.Role)) {
		response.Error(c, 4003, "无权执行此操作")
		return nil, false
	}
	return user, true
}

func validateSlideReq(req *dto.SaveSlideReq) error {
	req.Slug = strings.TrimSpace(req.Slug)
	req.Title = strings.TrimSpace(req.Title)
	req.TitleEn = strings.TrimSpace(req.TitleEn)
	req.Cover = strings.TrimSpace(req.Cover)
	req.Entry = strings.TrimSpace(req.Entry)
	req.ObjectPrefix = strings.TrimSpace(req.ObjectPrefix)
	if req.Slug == "" {
		return fmt.Errorf("幻灯片 id 不能为空")
	}
	if !slideSlugPattern.MatchString(req.Slug) {
		return fmt.Errorf("幻灯片 id 只能包含字母、数字、下划线和短横线，并且必须以字母或数字开头")
	}
	if req.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	return nil
}

func applySlideReq(entity *domain.SlideEntity, req *dto.SaveSlideReq, creating bool) error {
	entity.Slug = req.Slug
	entity.Title = req.Title
	entity.TitleEn = req.TitleEn
	entity.Description = strings.TrimSpace(req.Description)
	entity.DescriptionEn = strings.TrimSpace(req.DescriptionEn)
	entity.Cover = req.Cover
	entity.Entry = req.Entry
	entity.ObjectPrefix = req.ObjectPrefix
	entity.Tags = encodeTags(req.Tags)
	entity.Protected = req.Protected

	if entity.Entry == "" {
		entity.Entry = "/slides/decks/" + entity.Slug + "/"
	}

	password := strings.TrimSpace(req.Password)
	if entity.Protected {
		if password == "" && (creating || entity.PasswordHash == "") {
			return fmt.Errorf("受保护的幻灯片必须设置密码")
		}
		if password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			entity.PasswordHash = string(hash)
		}
	} else {
		entity.PasswordHash = ""
	}
	return nil
}

func slideEntitiesToDTO(slides *[]domain.SlideEntity) []*dto.SlideDTO {
	result := make([]*dto.SlideDTO, 0, len(*slides))
	for i := range *slides {
		result = append(result, slideEntityToDTO(&(*slides)[i]))
	}
	return result
}

func slideEntitiesToPublicDTO(slides *[]domain.SlideEntity) []*dto.SlideDTO {
	result := make([]*dto.SlideDTO, 0, len(*slides))
	for i := range *slides {
		item := slideEntityToDTO(&(*slides)[i])
		if item.Protected {
			item.Entry = "/slides/protected?id=" + item.ID
		}
		result = append(result, item)
	}
	return result
}

func slideEntityToDTO(slide *domain.SlideEntity) *dto.SlideDTO {
	return &dto.SlideDTO{
		DatabaseID:    slide.ID,
		ID:            slide.Slug,
		Title:         slide.Title,
		TitleEn:       slide.TitleEn,
		Description:   slide.Description,
		DescriptionEn: slide.DescriptionEn,
		Cover:         slide.Cover,
		Entry:         slide.Entry,
		ObjectPrefix:  slide.ObjectPrefix,
		Tags:          decodeTags(slide.Tags),
		Protected:     slide.Protected,
		HasPassword:   slide.PasswordHash != "",
		CreatedAt:     slide.CreatedAt.Unix(),
		UpdatedAt:     slide.UpdatedAt.Unix(),
	}
}

func encodeTags(tags []string) string {
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleaned = append(cleaned, tag)
		}
	}
	data, err := json.Marshal(cleaned)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeTags(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(value), &tags); err == nil {
		return tags
	}
	parts := strings.Split(value, ",")
	tags = make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tags = append(tags, part)
		}
	}
	return tags
}

func parseSlideID(id string) (uint, error) {
	value, err := strconv.ParseUint(id, 10, 64)
	return uint(value), err
}
