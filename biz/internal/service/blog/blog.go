package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
)

var blogSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,190}$`)

func (s *BlogService) ListPublicPosts(ctx context.Context, c *app.RequestContext) {
	s.listPosts(c, false)
}

func (s *BlogService) ListAdminPosts(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	s.listPosts(c, true)
}

func (s *BlogService) listPosts(c *app.RequestContext, admin bool) {
	page := positiveQueryInt(c, "page", 1)
	pageSize := positiveQueryInt(c, "pageSize", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	posts, total, err := s.Repo.ListBlogPosts(
		admin,
		(page-1)*pageSize,
		pageSize,
		strings.TrimSpace(c.DefaultQuery("q", "")),
		strings.TrimSpace(c.DefaultQuery("category", "")),
		strings.TrimSpace(c.DefaultQuery("tag", "")),
	)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	items := make([]*dto.BlogPostDTO, 0, len(*posts))
	for i := range *posts {
		post := &(*posts)[i]
		revisionID := post.PublishedRevisionID
		if admin && post.DraftRevisionID != 0 {
			revisionID = post.DraftRevisionID
		}
		revision, err := s.Repo.FindBlogRevisionByID(revisionID)
		if err != nil || revision.ID == 0 {
			continue
		}
		items = append(items, blogPostToDTO(post, revision, false, admin))
	}
	response.OK(c, &dto.BlogPostListDTO{Items: items, Total: total, Page: page, PageSize: pageSize}, "ok")
}

func (s *BlogService) GetPublicPost(ctx context.Context, c *app.RequestContext) {
	value := strings.TrimSpace(c.Param("slug"))
	if value == "" {
		value = strings.TrimSpace(c.Param("legacy"))
	}
	post, err := s.Repo.FindPublishedBlogPostBySlugOrLegacy(value)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	if post.ID == 0 || post.PublishedRevisionID == 0 {
		response.Error(c, 4004, "文章不存在")
		return
	}
	revision, err := s.Repo.FindBlogRevisionByID(post.PublishedRevisionID)
	if err != nil || revision.ID == 0 {
		response.Error(c, 5001, "文章版本不存在")
		return
	}
	response.OK(c, blogPostToDTO(post, revision, true, false), "ok")
}

func (s *BlogService) CreatePost(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}
	var req dto.SaveBlogPostReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if err := s.normalizeAndValidatePostReq(&req, 0); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	post := &domain.BlogPostEntity{
		Slug:            req.Slug,
		LegacyPermalink: normalizeLegacyPermalink(req.LegacyPermalink),
		Status:          domain.BlogStatusDraft,
	}
	revision := revisionFromRequest(&req, user)
	if err := s.Repo.CreateBlogPost(post, revision, req.PublishAfterSave); err != nil {
		response.Error(c, 5001, "保存文章失败: "+err.Error())
		return
	}
	response.OK(c, blogPostToDTO(post, revision, true, true), "ok")
}

func (s *BlogService) SaveDraft(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	post, err := s.Repo.FindBlogPostByID(postID)
	if err != nil || post.ID == 0 {
		response.Error(c, 4004, "文章不存在")
		return
	}
	var req dto.SaveBlogPostReq
	if err = c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误: "+err.Error())
		return
	}
	if req.Slug == "" {
		req.Slug = post.Slug
	}
	if req.LegacyPermalink == "" {
		req.LegacyPermalink = post.LegacyPermalink
	}
	if err = s.normalizeAndValidatePostReq(&req, postID); err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	revision := revisionFromRequest(&req, user)
	post, err = s.Repo.SaveBlogDraft(postID, req.BaseRevisionID, revision)
	if err != nil {
		if err.Error() == "draft revision conflict" {
			response.Error(c, 4090, "草稿已在其他页面更新，请刷新后再保存")
		} else {
			response.Error(c, 5001, "保存草稿失败: "+err.Error())
		}
		return
	}
	if err = s.Repo.UpdateBlogPostIdentity(postID, req.Slug, normalizeLegacyPermalink(req.LegacyPermalink)); err != nil {
		response.Error(c, 5001, "更新文章地址失败: "+err.Error())
		return
	}
	post.Slug = req.Slug
	post.LegacyPermalink = normalizeLegacyPermalink(req.LegacyPermalink)
	if req.PublishAfterSave {
		post, err = s.Repo.PublishBlogPost(postID)
		if err != nil {
			response.Error(c, 5001, "发布文章失败: "+err.Error())
			return
		}
	}
	response.OK(c, blogPostToDTO(post, revision, true, true), "ok")
}

func (s *BlogService) PublishPost(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	post, err := s.Repo.PublishBlogPost(postID)
	if err != nil {
		response.Error(c, 5001, "发布失败: "+err.Error())
		return
	}
	revision, err := s.Repo.FindBlogRevisionByID(post.PublishedRevisionID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, blogPostToDTO(post, revision, true, true), "发布成功")
}

func (s *BlogService) UnpublishPost(ctx context.Context, c *app.RequestContext) {
	s.setStatus(ctx, c, domain.BlogStatusDraft, "已撤回")
}

func (s *BlogService) ArchivePost(ctx context.Context, c *app.RequestContext) {
	s.setStatus(ctx, c, domain.BlogStatusArchived, "已归档")
}

func (s *BlogService) setStatus(ctx context.Context, c *app.RequestContext, status, message string) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := s.Repo.SetBlogPostStatus(postID, status); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, nil, message)
}

func (s *BlogService) DeletePost(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := s.Repo.RemoveBlogPost(postID); err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, nil, "删除成功")
}

func (s *BlogService) ListRevisions(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	revisions, err := s.Repo.ListBlogRevisions(postID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	items := make([]*dto.BlogRevisionDTO, 0, len(*revisions))
	for i := range *revisions {
		items = append(items, blogRevisionToDTO(&(*revisions)[i]))
	}
	response.OK(c, items, "ok")
}

func (s *BlogService) GetRevision(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	revisionID, ok := parseIDParam(c, "revisionId")
	if !ok {
		return
	}
	revision, err := s.Repo.FindBlogRevisionByID(revisionID)
	if err != nil || revision.ID == 0 {
		response.Error(c, 4004, "版本不存在")
		return
	}
	response.OK(c, blogRevisionToDTO(revision), "ok")
}

func (s *BlogService) RestoreRevision(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}
	postID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	revisionID, ok := parseIDParam(c, "revisionId")
	if !ok {
		return
	}
	post, err := s.Repo.FindBlogPostByID(postID)
	if err != nil || post.ID == 0 {
		response.Error(c, 4004, "文章不存在")
		return
	}
	source, err := s.Repo.FindBlogRevisionByID(revisionID)
	if err != nil || source.ID == 0 || source.PostID != postID {
		response.Error(c, 4004, "版本不存在")
		return
	}
	restored := *source
	restored.ID = 0
	restored.CreatedAt = time.Time{}
	restored.UpdatedAt = time.Time{}
	restored.DeletedAt.Valid = false
	restored.AuthorID = user.ID
	restored.AuthorUsername = user.Username
	restored.ChangeSummary = fmt.Sprintf("恢复自版本 %d", source.Version)
	post, err = s.Repo.SaveBlogDraft(postID, post.DraftRevisionID, &restored)
	if err != nil {
		response.Error(c, 5001, "恢复版本失败: "+err.Error())
		return
	}
	response.OK(c, blogPostToDTO(post, &restored, true, true), "已恢复为新草稿")
}

func (s *BlogService) normalizeAndValidatePostReq(req *dto.SaveBlogPostReq, currentID uint) error {
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if req.Slug == "" {
		req.Slug = strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
	}
	if !blogSlugPattern.MatchString(req.Slug) {
		return fmt.Errorf("文章地址只能包含小写字母、数字和连字符")
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if len([]rune(req.Title)) > 300 {
		return fmt.Errorf("标题过长")
	}
	if len([]byte(req.ContentMarkdown)) > 4*1024*1024 {
		return fmt.Errorf("正文不能超过 4 MiB")
	}
	existing, err := s.Repo.FindBlogPostBySlug(req.Slug)
	if err != nil {
		return err
	}
	if existing.ID != 0 && existing.ID != currentID {
		return fmt.Errorf("文章地址已存在")
	}
	req.Categories = normalizeList(req.Categories)
	req.Tags = normalizeList(req.Tags)
	req.Description = strings.TrimSpace(req.Description)
	req.ChangeSummary = strings.TrimSpace(req.ChangeSummary)
	return nil
}

func (s *BlogService) requireSuperAdmin(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
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

func revisionFromRequest(req *dto.SaveBlogPostReq, user *domain.UserEntity) *domain.BlogRevisionEntity {
	return &domain.BlogRevisionEntity{
		Title:           req.Title,
		Description:     req.Description,
		ContentMarkdown: req.ContentMarkdown,
		Cover:           strings.TrimSpace(req.Cover),
		CoverObjectPath: strings.TrimSpace(req.CoverObjectPath),
		Categories:      encodeList(req.Categories),
		Tags:            encodeList(req.Tags),
		ChangeSummary:   req.ChangeSummary,
		AuthorID:        user.ID,
		AuthorUsername:  user.Username,
	}
}

func blogPostToDTO(post *domain.BlogPostEntity, revision *domain.BlogRevisionEntity, includeContent, admin bool) *dto.BlogPostDTO {
	result := &dto.BlogPostDTO{
		DatabaseID: post.ID, Slug: post.Slug, LegacyPermalink: post.LegacyPermalink,
		Status: post.Status, Title: revision.Title, Description: revision.Description,
		Cover: revision.Cover, Categories: decodeList(revision.Categories), Tags: decodeList(revision.Tags),
		RevisionID: revision.ID, Version: revision.Version, ChangeSummary: revision.ChangeSummary,
		AuthorUsername: revision.AuthorUsername, CreatedAt: post.CreatedAt.Unix(), UpdatedAt: post.UpdatedAt.Unix(),
		LikeCount: post.LikeCount, CommentCount: post.ApprovedCommentCount,
	}
	if includeContent {
		result.ContentMarkdown = revision.ContentMarkdown
	}
	if admin {
		result.DraftRevisionID = post.DraftRevisionID
		result.PublishedRevisionID = post.PublishedRevisionID
		result.CoverObjectPath = revision.CoverObjectPath
		viewCount := post.ViewCount
		result.ViewCount = &viewCount
	}
	if post.PublishedAt != nil {
		result.PublishedAt = post.PublishedAt.Unix()
	}
	return result
}

func blogRevisionToDTO(revision *domain.BlogRevisionEntity) *dto.BlogRevisionDTO {
	return &dto.BlogRevisionDTO{
		RevisionID: revision.ID, PostID: revision.PostID, Version: revision.Version,
		Title: revision.Title, Description: revision.Description, ContentMarkdown: revision.ContentMarkdown,
		Cover: revision.Cover, CoverObjectPath: revision.CoverObjectPath,
		Categories: decodeList(revision.Categories), Tags: decodeList(revision.Tags),
		ChangeSummary: revision.ChangeSummary, AuthorUsername: revision.AuthorUsername,
		CreatedAt: revision.CreatedAt.Unix(),
	}
}

func normalizeList(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func encodeList(values []string) string {
	data, _ := json.Marshal(normalizeList(values))
	return string(data)
}

func decodeList(value string) []string {
	var result []string
	_ = json.Unmarshal([]byte(value), &result)
	if result == nil {
		return []string{}
	}
	return result
}

func normalizeLegacyPermalink(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = "/" + strings.Trim(value, "/") + "/"
	return value
}

func positiveQueryInt(c *app.RequestContext, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseIDParam(c *app.RequestContext, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, 2001, "参数错误")
		return 0, false
	}
	return uint(value), true
}
