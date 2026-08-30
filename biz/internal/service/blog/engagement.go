package blog

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"personal-page-be/biz/internal/domain"
	"personal-page-be/biz/internal/dto"
	"personal-page-be/biz/internal/response"
)

const (
	maxBlogCommentRunes = 2000
	maxBlogCommentBytes = 8 * 1024
)

func (s *BlogService) TrackView(ctx context.Context, c *app.RequestContext) {
	post, ok := s.publishedPost(c)
	if !ok {
		return
	}
	if err := s.Repo.IncrementBlogPostView(post.ID); err != nil {
		response.Error(c, 5001, "记录浏览量失败")
		return
	}
	response.OK(c, nil, "ok")
}

func (s *BlogService) GetLikeState(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireUser(ctx, c)
	if !ok {
		return
	}
	post, ok := s.publishedPost(c)
	if !ok {
		return
	}
	like, err := s.Repo.GetBlogPostLike(post.ID, user.ID)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.OK(c, &dto.BlogLikeStateDTO{Liked: like.ID != 0, LikeCount: post.LikeCount}, "ok")
}

func (s *BlogService) ToggleLike(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireUser(ctx, c)
	if !ok {
		return
	}
	post, ok := s.publishedPost(c)
	if !ok {
		return
	}
	liked, count, err := s.Repo.ToggleBlogPostLike(post.ID, user.ID)
	if err != nil {
		response.Error(c, 5001, "更新点赞失败: "+err.Error())
		return
	}
	response.OK(c, &dto.BlogLikeStateDTO{Liked: liked, LikeCount: count}, "ok")
}

func (s *BlogService) ListPublicComments(ctx context.Context, c *app.RequestContext) {
	post, ok := s.publishedPost(c)
	if !ok {
		return
	}
	page := positiveQueryInt(c, "page", 1)
	pageSize := positiveQueryInt(c, "pageSize", 30)
	if pageSize > 100 {
		pageSize = 100
	}
	comments, total, err := s.Repo.ListApprovedBlogComments(post.ID, (page-1)*pageSize, pageSize)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	items := make([]*dto.BlogCommentDTO, 0, len(*comments))
	for i := range *comments {
		items = append(items, blogCommentToDTO(&(*comments)[i], false))
	}
	response.OK(c, &dto.BlogCommentListDTO{Items: items, Total: total, Page: page, PageSize: pageSize}, "ok")
}

func (s *BlogService) CreateComment(ctx context.Context, c *app.RequestContext) {
	user, ok := s.requireUser(ctx, c)
	if !ok {
		return
	}
	post, ok := s.publishedPost(c)
	if !ok {
		return
	}
	var req dto.CreateBlogCommentReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	content, err := validateBlogCommentContent(req.Content)
	if err != nil {
		response.Error(c, 2001, err.Error())
		return
	}
	revision, err := s.Repo.FindBlogRevisionByID(post.PublishedRevisionID)
	if err != nil || revision.ID == 0 {
		response.Error(c, 5001, "文章版本不存在")
		return
	}
	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = user.Username
	}
	comment := &domain.BlogCommentEntity{
		PostID: post.ID, PostSlug: post.Slug, PostTitle: revision.Title,
		UserID: user.ID, AuthorUsername: user.Username, AuthorNickname: nickname,
		AuthorAvatar: strings.TrimSpace(user.Avatar), Content: content,
		Status: domain.BlogCommentStatusPending,
	}
	if err = s.Repo.CreateBlogComment(comment); err != nil {
		response.Error(c, 5001, "提交评论失败")
		return
	}
	response.OK(c, blogCommentToDTO(comment, true), "评论已提交，审核通过后会公开显示")
}

func (s *BlogService) ListAdminComments(ctx context.Context, c *app.RequestContext) {
	if _, ok := s.requireSuperAdmin(ctx, c); !ok {
		return
	}
	status := strings.ToLower(strings.TrimSpace(c.DefaultQuery("status", domain.BlogCommentStatusPending)))
	if status == "all" {
		status = ""
	}
	if status != "" && !isBlogCommentStatus(status) {
		response.Error(c, 2001, "评论状态无效")
		return
	}
	page := positiveQueryInt(c, "page", 1)
	pageSize := positiveQueryInt(c, "pageSize", 50)
	if pageSize > 100 {
		pageSize = 100
	}
	comments, total, err := s.Repo.ListBlogComments(status, (page-1)*pageSize, pageSize)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	items := make([]*dto.BlogCommentDTO, 0, len(*comments))
	for i := range *comments {
		items = append(items, blogCommentToDTO(&(*comments)[i], true))
	}
	response.OK(c, &dto.BlogCommentListDTO{Items: items, Total: total, Page: page, PageSize: pageSize}, "ok")
}

func (s *BlogService) ReviewComment(ctx context.Context, c *app.RequestContext) {
	reviewer, ok := s.requireSuperAdmin(ctx, c)
	if !ok {
		return
	}
	commentID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ReviewBlogCommentReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, 2001, "参数错误")
		return
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != domain.BlogCommentStatusApproved && status != domain.BlogCommentStatusRejected {
		response.Error(c, 2001, "审核状态只能是 approved 或 rejected")
		return
	}
	comment, err := s.Repo.ReviewBlogComment(commentID, status, reviewer.ID, reviewer.Username, time.Now())
	if err != nil {
		response.Error(c, 5001, "审核评论失败: "+err.Error())
		return
	}
	response.OK(c, blogCommentToDTO(comment, true), "审核结果已保存")
}

func (s *BlogService) publishedPost(c *app.RequestContext) (*domain.BlogPostEntity, bool) {
	value := strings.TrimSpace(c.Param("slug"))
	post, err := s.Repo.FindPublishedBlogPostBySlugOrLegacy(value)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return nil, false
	}
	if post.ID == 0 || post.PublishedRevisionID == 0 {
		response.Error(c, 4004, "文章不存在")
		return nil, false
	}
	return post, true
}

func (s *BlogService) requireUser(ctx context.Context, c *app.RequestContext) (*domain.UserEntity, bool) {
	username, _ := ctx.Value("username").(string)
	user, err := s.Repo.FindUser(username)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return nil, false
	}
	if user.ID == 0 || !user.CanUse {
		response.Error(c, 4003, "账号不可用")
		return nil, false
	}
	return user, true
}

func validateBlogCommentContent(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("评论内容不能为空")
	}
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("评论内容编码无效")
	}
	if utf8.RuneCountInString(value) > maxBlogCommentRunes || len([]byte(value)) > maxBlogCommentBytes {
		return "", fmt.Errorf("评论不能超过 %d 个字符", maxBlogCommentRunes)
	}
	return value, nil
}

func isBlogCommentStatus(value string) bool {
	return value == domain.BlogCommentStatusPending ||
		value == domain.BlogCommentStatusApproved ||
		value == domain.BlogCommentStatusRejected
}

func blogCommentToDTO(comment *domain.BlogCommentEntity, admin bool) *dto.BlogCommentDTO {
	result := &dto.BlogCommentDTO{
		ID: comment.ID, PostID: comment.PostID, PostSlug: comment.PostSlug, PostTitle: comment.PostTitle,
		Content: comment.Content, AuthorUsername: comment.AuthorUsername,
		AuthorNickname: comment.AuthorNickname, AuthorAvatar: comment.AuthorAvatar,
		CreatedAt: comment.CreatedAt.Unix(),
	}
	if admin {
		result.Status = comment.Status
		result.ReviewerUsername = comment.ReviewerUsername
		if comment.ReviewedAt != nil {
			result.ReviewedAt = comment.ReviewedAt.Unix()
		}
	}
	return result
}
