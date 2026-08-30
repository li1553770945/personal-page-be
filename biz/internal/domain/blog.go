package domain

import (
	"time"

	"personal-page-be/biz/internal/do"
)

const (
	BlogStatusDraft     = "draft"
	BlogStatusPublished = "published"
	BlogStatusArchived  = "archived"
)

type BlogPostEntity struct {
	do.BaseModel
	Slug                string     `gorm:"uniqueIndex;size:191" json:"slug"`
	LegacyPermalink     string     `gorm:"index;size:191" json:"legacy_permalink"`
	Status              string     `gorm:"size:32;index" json:"status"`
	DraftRevisionID     uint       `gorm:"index" json:"draft_revision_id"`
	PublishedRevisionID uint       `gorm:"index" json:"published_revision_id"`
	PublishedAt         *time.Time `gorm:"index" json:"published_at"`
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

type BlogMigrationEntity struct {
	do.BaseModel
	MigrationKey string `gorm:"column:migration_key;uniqueIndex;size:191" json:"migration_key"`
}

func (BlogMigrationEntity) TableName() string { return "blog_migrations" }
