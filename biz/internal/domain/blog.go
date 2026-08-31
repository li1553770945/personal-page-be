package domain

import (
	"time"

	"personal-page-be/biz/internal/do"
)

const (
	BlogStatusDraft     = "draft"
	BlogStatusPublished = "published"
	BlogStatusArchived  = "archived"

	BlogCommentStatusPending  = "pending"
	BlogCommentStatusApproved = "approved"
	BlogCommentStatusRejected = "rejected"
)

type BlogPostEntity struct {
	do.BaseModel
	Slug                 string     `gorm:"uniqueIndex;size:191" json:"slug"`
	LegacyPermalink      string     `gorm:"index;size:191" json:"legacy_permalink"`
	Status               string     `gorm:"size:32;index" json:"status"`
	DraftRevisionID      uint       `gorm:"index" json:"draft_revision_id"`
	PublishedRevisionID  uint       `gorm:"index" json:"published_revision_id"`
	PublishedAt          *time.Time `gorm:"index" json:"published_at"`
	Pinned               bool       `gorm:"not null;default:false;index" json:"pinned"`
	PinnedAt             *time.Time `gorm:"index" json:"pinned_at"`
	ViewCount            int64      `gorm:"not null;default:0" json:"-"`
	LikeCount            int64      `gorm:"not null;default:0" json:"-"`
	ApprovedCommentCount int64      `gorm:"not null;default:0" json:"-"`
}

func (BlogPostEntity) TableName() string { return "blog_posts" }

type BlogRevisionEntity struct {
	do.BaseModel
	PostID          uint   `gorm:"uniqueIndex:idx_blog_revision_version,priority:1;index" json:"post_id"`
	Version         int    `gorm:"uniqueIndex:idx_blog_revision_version,priority:2" json:"version"`
	Title           string `gorm:"size:512" json:"title"`
	Description     string `gorm:"type:text" json:"description"`
	ContentMarkdown string `gorm:"type:text" json:"content_markdown"`
	Cover           string `gorm:"type:text" json:"cover"`
	CoverObjectPath string `gorm:"type:text" json:"-"`
	Categories      string `gorm:"type:text" json:"categories"`
	Tags            string `gorm:"type:text" json:"tags"`
	ChangeSummary   string `gorm:"size:512" json:"change_summary"`
	AuthorID        uint   `gorm:"index" json:"author_id"`
	AuthorUsername  string `gorm:"size:191" json:"author_username"`
}

func (BlogRevisionEntity) TableName() string { return "blog_post_revisions" }

type BlogAssetEntity struct {
	do.BaseModel
	PostID     uint   `gorm:"index" json:"post_id"`
	ObjectPath string `gorm:"uniqueIndex;size:512" json:"object_path"`
	FileName   string `gorm:"size:512" json:"file_name"`
	MimeType   string `gorm:"size:128" json:"mime_type"`
	Size       int64  `json:"size"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Alt        string `gorm:"size:512" json:"alt"`
	CreatedBy  uint   `gorm:"index" json:"created_by"`
	Ready      bool   `gorm:"index" json:"ready"`
}

func (BlogAssetEntity) TableName() string { return "blog_assets" }

type BlogLikeEntity struct {
	do.BaseModel
	PostID uint `gorm:"uniqueIndex:idx_blog_like_post_user,priority:1;index" json:"post_id"`
	UserID uint `gorm:"uniqueIndex:idx_blog_like_post_user,priority:2;index" json:"user_id"`
}

func (BlogLikeEntity) TableName() string { return "blog_likes" }

type BlogCommentEntity struct {
	do.BaseModel
	PostID           uint       `gorm:"index" json:"post_id"`
	PostSlug         string     `gorm:"size:191;index" json:"post_slug"`
	PostTitle        string     `gorm:"size:512" json:"post_title"`
	UserID           uint       `gorm:"index" json:"user_id"`
	AuthorUsername   string     `gorm:"size:191" json:"author_username"`
	AuthorNickname   string     `gorm:"size:191" json:"author_nickname"`
	AuthorAvatar     string     `gorm:"type:text" json:"author_avatar"`
	Content          string     `gorm:"type:text" json:"content"`
	Status           string     `gorm:"size:32;index" json:"status"`
	ReviewerID       uint       `gorm:"index" json:"reviewer_id"`
	ReviewerUsername string     `gorm:"size:191" json:"reviewer_username"`
	ReviewedAt       *time.Time `gorm:"index" json:"reviewed_at"`
}

func (BlogCommentEntity) TableName() string { return "blog_comments" }

type BlogMigrationEntity struct {
	do.BaseModel
	MigrationKey string `gorm:"column:migration_key;uniqueIndex;size:191" json:"migration_key"`
}

func (BlogMigrationEntity) TableName() string { return "blog_migrations" }
