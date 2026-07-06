package slide

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/tencentyun/cos-go-sdk-v5"
	"golang.org/x/crypto/bcrypt"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
)

var slideSlugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
var slideDeckBasePattern = regexp.MustCompile(`/slides/decks/[^"'()\s]+/`)

const slideAccessMaxAgeSeconds = 3600

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

func (s *SlideService) UploadSlideDeck(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	slug := strings.TrimSpace(string(c.FormValue("id")))
	if err := validateSlideSlug(slug); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		response.Error(c, 2001, "请上传 Slidev 导出的 zip 文件")
		return
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".zip") {
		response.Error(c, 2001, "幻灯片包必须是 zip 文件")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, 5001, "读取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, 5001, "读取上传文件失败: "+err.Error())
		return
	}
	entry, prefix, count, err := s.uploadSlideZip(ctx, slug, data)
	if err != nil {
		response.Error(c, 5001, "上传幻灯片包失败: "+err.Error())
		return
	}
	response.OK(c, &dto.SlideUploadDTO{
		Entry:        entry,
		ObjectPrefix: prefix,
		FileCount:    count,
	}, "ok")
}

func (s *SlideService) UploadSlideCover(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	slug := strings.TrimSpace(string(c.FormValue("id")))
	if err := validateSlideSlug(slug); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		response.Error(c, 2001, "请上传封面图")
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".png"
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, 5001, "读取封面失败: "+err.Error())
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, 5001, "读取封面失败: "+err.Error())
		return
	}
	objectPath := fmt.Sprintf("slides/%s/cover%s", slug, ext)
	if err = s.putCOSObject(ctx, objectPath, data, contentTypeForPath(objectPath, data)); err != nil {
		response.Error(c, 5001, "上传封面失败: "+err.Error())
		return
	}
	response.OK(c, &dto.SlideUploadDTO{
		Cover:           "/api/slides/" + slug + "/cover",
		CoverObjectPath: objectPath,
	}, "ok")
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
	c.SetCookie(slideAccessCookieName(slug), s.issueSlideAccessToken(slug), slideAccessMaxAgeSeconds, "/", "", protocol.CookieSameSiteLaxMode, true, true)
	response.OK(c, slideEntityToDTO(entity), "ok")
}

func (s *SlideService) ServeSlideAsset(ctx context.Context, c *app.RequestContext) {
	slug := strings.TrimSpace(c.Param("slug"))
	entity, ok := s.findSlideForAsset(slug, c)
	if !ok {
		return
	}
	if entity.Protected && !s.hasSlideAccess(slug, c) {
		c.Data(consts.StatusForbidden, "text/plain; charset=utf-8", []byte("forbidden"))
		return
	}
	assetPath, ok := cleanAssetPath(c.Param("path"))
	if !ok {
		c.Data(consts.StatusBadRequest, "text/plain; charset=utf-8", []byte("bad asset path"))
		return
	}
	if assetPath == "" {
		assetPath = "index.html"
	}
	prefix := strings.Trim(strings.TrimSpace(entity.ObjectPrefix), "/")
	if prefix == "" {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("slide assets not uploaded"))
		return
	}
	s.serveCOSObject(ctx, c, prefix+"/"+assetPath)
}

func (s *SlideService) ServeSlideCover(ctx context.Context, c *app.RequestContext) {
	slug := strings.TrimSpace(c.Param("slug"))
	entity, ok := s.findSlideForAsset(slug, c)
	if !ok {
		return
	}
	if entity.CoverObjectPath == "" {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("cover not uploaded"))
		return
	}
	s.serveCOSObject(ctx, c, entity.CoverObjectPath)
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
	if err := validateSlideSlug(req.Slug); err != nil {
		return err
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
	entity.CoverObjectPath = strings.TrimSpace(req.CoverObjectPath)
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
	cover := slide.Cover
	if slide.CoverObjectPath != "" {
		cover = "/api/slides/" + slide.Slug + "/cover"
	}
	return &dto.SlideDTO{
		DatabaseID:      slide.ID,
		ID:              slide.Slug,
		Title:           slide.Title,
		TitleEn:         slide.TitleEn,
		Description:     slide.Description,
		DescriptionEn:   slide.DescriptionEn,
		Cover:           cover,
		CoverObjectPath: slide.CoverObjectPath,
		Entry:           slide.Entry,
		ObjectPrefix:    slide.ObjectPrefix,
		Tags:            decodeTags(slide.Tags),
		Protected:       slide.Protected,
		HasPassword:     slide.PasswordHash != "",
		CreatedAt:       slide.CreatedAt.Unix(),
		UpdatedAt:       slide.UpdatedAt.Unix(),
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

func validateSlideSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("幻灯片 id 不能为空")
	}
	if !slideSlugPattern.MatchString(slug) {
		return fmt.Errorf("幻灯片 id 只能包含字母、数字、下划线和短横线，并且必须以字母或数字开头")
	}
	return nil
}

func parseSlideID(id string) (uint, error) {
	value, err := strconv.ParseUint(id, 10, 64)
	return uint(value), err
}

type zipSlideEntry struct {
	file *zip.File
	path string
}

func (s *SlideService) uploadSlideZip(ctx context.Context, slug string, data []byte) (string, string, int, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", "", 0, err
	}
	entries := make([]zipSlideEntry, 0, len(reader.File))
	for _, file := range reader.File {
		cleaned, ok := cleanZipPath(file.Name)
		if !ok || file.FileInfo().IsDir() || ignoredSlideZipPath(cleaned) {
			continue
		}
		entries = append(entries, zipSlideEntry{file: file, path: cleaned})
	}
	entries = stripCommonZipRoot(entries)
	hasIndex := false
	for _, entry := range entries {
		if entry.path == "index.html" {
			hasIndex = true
			break
		}
	}
	if !hasIndex {
		return "", "", 0, fmt.Errorf("zip 根目录下没有 index.html，请上传 Slidev build 输出目录压缩包")
	}

	prefix := "slides/" + slug + "/"
	for _, entry := range entries {
		content, err := readZipFile(entry.file)
		if err != nil {
			return "", "", 0, err
		}
		if shouldRewriteSlideAssetRefs(entry.path) {
			content = rewriteSlideAssetRefs(content, slug)
		}
		objectPath := prefix + entry.path
		if err = s.putCOSObject(ctx, objectPath, content, contentTypeForPath(entry.path, content)); err != nil {
			return "", "", 0, err
		}
	}
	return "/api/slides/" + slug + "/assets/", prefix, len(entries), nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func cleanZipPath(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	cleaned := path.Clean("/" + name)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." || cleaned == "" || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "..") {
		return "", false
	}
	return cleaned, true
}

func cleanAssetPath(name string) (string, bool) {
	name = strings.TrimPrefix(name, "/")
	if strings.TrimSpace(name) == "" {
		return "", true
	}
	return cleanZipPath(name)
}

func ignoredSlideZipPath(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(name, "__MACOSX/") || base == ".DS_Store"
}

func stripCommonZipRoot(entries []zipSlideEntry) []zipSlideEntry {
	if len(entries) == 0 {
		return entries
	}
	roots := map[string]bool{}
	for _, entry := range entries {
		parts := strings.SplitN(entry.path, "/", 2)
		if len(parts) < 2 {
			return entries
		}
		roots[parts[0]] = true
	}
	if len(roots) != 1 {
		return entries
	}
	var root string
	for value := range roots {
		root = value
	}
	prefix := root + "/"
	hasIndex := false
	for _, entry := range entries {
		if entry.path == prefix+"index.html" {
			hasIndex = true
			break
		}
	}
	if !hasIndex {
		return entries
	}
	for i := range entries {
		entries[i].path = strings.TrimPrefix(entries[i].path, prefix)
	}
	return entries
}

func shouldRewriteSlideAssetRefs(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".html", ".js", ".mjs", ".css":
		return true
	default:
		return false
	}
}

func rewriteSlideAssetRefs(content []byte, slug string) []byte {
	return []byte(slideDeckBasePattern.ReplaceAllString(string(content), "/api/slides/"+slug+"/assets/"))
}

func contentTypeForPath(name string, content []byte) string {
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); value != "" {
		return value
	}
	if len(content) > 0 {
		return http.DetectContentType(content)
	}
	return "application/octet-stream"
}

func (s *SlideService) putCOSObject(ctx context.Context, objectPath string, data []byte, contentType string) error {
	client, err := s.cosClient()
	if err != nil {
		return err
	}
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	resp, err := client.Object.Put(ctx, objectPath, bytes.NewReader(data), opt)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return err
}

func (s *SlideService) serveCOSObject(ctx context.Context, c *app.RequestContext, objectPath string) {
	client, err := s.cosClient()
	if err != nil {
		c.Data(consts.StatusInternalServerError, "text/plain; charset=utf-8", []byte(err.Error()))
		return
	}
	resp, err := client.Object.Get(ctx, objectPath, nil)
	if err != nil {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("not found"))
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Data(consts.StatusInternalServerError, "text/plain; charset=utf-8", []byte(err.Error()))
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = contentTypeForPath(objectPath, data)
	}
	c.Response.Header.Set("Cache-Control", "public, max-age=300")
	c.Data(consts.StatusOK, contentType, data)
}

func (s *SlideService) findSlideForAsset(slug string, c *app.RequestContext) (*domain.SlideEntity, bool) {
	if slug == "" {
		c.Data(consts.StatusBadRequest, "text/plain; charset=utf-8", []byte("missing slide id"))
		return nil, false
	}
	entity, err := s.Repo.FindSlideBySlug(slug)
	if err != nil {
		c.Data(consts.StatusInternalServerError, "text/plain; charset=utf-8", []byte(err.Error()))
		return nil, false
	}
	if entity.ID == 0 {
		c.Data(consts.StatusNotFound, "text/plain; charset=utf-8", []byte("slide not found"))
		return nil, false
	}
	return entity, true
}

func (s *SlideService) cosClient() (*cos.Client, error) {
	endpoint := s.Config.EffectiveCOSEndpoint()
	ak := s.Config.EffectiveCOSAk()
	sk := s.Config.EffectiveCOSSk()
	if endpoint == "" || ak == "" || sk == "" {
		return nil, fmt.Errorf("COS config is incomplete")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ak,
			SecretKey: sk,
		},
	}), nil
}

func slideAccessCookieName(slug string) string {
	return "slide_access_" + slug
}

func (s *SlideService) issueSlideAccessToken(slug string) string {
	exp := time.Now().Add(time.Duration(slideAccessMaxAgeSeconds) * time.Second).Unix()
	payload := fmt.Sprintf("%s:%d", slug, exp)
	mac := hmac.New(sha256.New, []byte(s.Config.EffectiveJWTKey()))
	_, _ = mac.Write([]byte(payload))
	raw := payload + "." + hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func (s *SlideService) hasSlideAccess(slug string, c *app.RequestContext) bool {
	token := string(c.Cookie(slideAccessCookieName(slug)))
	if token == "" {
		return false
	}
	rawBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	raw := string(rawBytes)
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return false
	}
	payload := parts[0]
	sig := parts[1]
	payloadParts := strings.Split(payload, ":")
	if len(payloadParts) != 2 || payloadParts[0] != slug {
		return false
	}
	exp, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.Config.EffectiveJWTKey()))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}
