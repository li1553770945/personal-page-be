package repo

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"personal-page-be/biz/internal/domain"
)

func (Repo *Repository) CountBlogPosts() (int64, error) {
	var count int64
	err := Repo.DB.Model(&domain.BlogPostEntity{}).Count(&count).Error
	return count, err
}

func (Repo *Repository) HasBlogMigration(key string) (bool, error) {
	var count int64
	err := Repo.DB.Model(&domain.BlogMigrationEntity{}).Where("migration_key = ?", key).Count(&count).Error
	return count > 0, err
}

func (Repo *Repository) SaveBlogMigration(key string) error {
	migration := &domain.BlogMigrationEntity{MigrationKey: key}
	return Repo.DB.Where("migration_key = ?", key).FirstOrCreate(migration).Error
}

func (Repo *Repository) FindBlogPostByID(postID uint) (*domain.BlogPostEntity, error) {
	var post domain.BlogPostEntity
	err := Repo.DB.Where("id = ?", postID).Limit(1).Find(&post).Error
	return &post, err
}

func (Repo *Repository) FindBlogPostBySlug(slug string) (*domain.BlogPostEntity, error) {
	var post domain.BlogPostEntity
	err := Repo.DB.Where("slug = ?", slug).Limit(1).Find(&post).Error
	return &post, err
}

func (Repo *Repository) FindPublishedBlogPostBySlugOrLegacy(value string) (*domain.BlogPostEntity, error) {
	var post domain.BlogPostEntity
	legacy := strings.Trim(value, "/")
	if strings.HasPrefix(legacy, "pages/") {
		legacy = strings.TrimPrefix(legacy, "pages/")
	}
	err := Repo.DB.Where("status = ? AND (slug = ? OR legacy_permalink = ? OR legacy_permalink = ?)",
		domain.BlogStatusPublished, value, value, "/pages/"+legacy+"/").Limit(1).Find(&post).Error
	return &post, err
}

func (Repo *Repository) ListBlogPosts(admin bool, offset, limit int, query, category, tag string) (*[]domain.BlogPostEntity, int64, error) {
	var posts []domain.BlogPostEntity
	var count int64
	db := Repo.DB.Model(&domain.BlogPostEntity{})
	if !admin {
		db = db.Where("status = ?", domain.BlogStatusPublished)
	}
	if query != "" || category != "" || tag != "" {
		db = db.Joins("JOIN blog_post_revisions ON blog_post_revisions.id = CASE WHEN blog_posts.status = ? THEN blog_posts.published_revision_id ELSE blog_posts.draft_revision_id END", domain.BlogStatusPublished)
	}
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("blog_post_revisions.title LIKE ? OR blog_post_revisions.description LIKE ?", like, like)
	}
	if category != "" {
		db = db.Where("blog_post_revisions.categories LIKE ?", "%\""+category+"\"%")
	}
	if tag != "" {
		db = db.Where("blog_post_revisions.tags LIKE ?", "%\""+tag+"\"%")
	}
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("COALESCE(blog_posts.published_at, blog_posts.created_at) DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return &posts, count, err
}

func (Repo *Repository) FindBlogRevisionByID(revisionID uint) (*domain.BlogRevisionEntity, error) {
	var revision domain.BlogRevisionEntity
	err := Repo.DB.Where("id = ?", revisionID).Limit(1).Find(&revision).Error
	return &revision, err
}

func (Repo *Repository) ListBlogRevisions(postID uint) (*[]domain.BlogRevisionEntity, error) {
	var revisions []domain.BlogRevisionEntity
	err := Repo.DB.Where("post_id = ?", postID).Order("version DESC").Find(&revisions).Error
	return &revisions, err
}

func (Repo *Repository) CreateBlogPost(post *domain.BlogPostEntity, revision *domain.BlogRevisionEntity, publish bool) error {
	return Repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		revision.PostID = post.ID
		revision.Version = 1
		if err := tx.Create(revision).Error; err != nil {
			return err
		}
		post.DraftRevisionID = revision.ID
		if publish {
			now := revision.CreatedAt
			if now.IsZero() {
				now = time.Now()
			}
			post.Status = domain.BlogStatusPublished
			post.PublishedRevisionID = revision.ID
			post.PublishedAt = &now
		}
		return tx.Save(post).Error
	})
}

func (Repo *Repository) SaveBlogDraft(postID, baseRevisionID uint, revision *domain.BlogRevisionEntity) (*domain.BlogPostEntity, error) {
	var result domain.BlogPostEntity
	err := Repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&result, postID).Error; err != nil {
			return err
		}
		if baseRevisionID != 0 && result.DraftRevisionID != baseRevisionID {
			return errors.New("draft revision conflict")
		}
		var maxVersion int
		if err := tx.Model(&domain.BlogRevisionEntity{}).Where("post_id = ?", postID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error; err != nil {
			return err
		}
		revision.PostID = postID
		revision.Version = maxVersion + 1
		if err := tx.Create(revision).Error; err != nil {
			return err
		}
		result.DraftRevisionID = revision.ID
		return tx.Save(&result).Error
	})
	return &result, err
}

func (Repo *Repository) UpdateBlogPostIdentity(postID uint, slug, legacyPermalink string) error {
	return Repo.DB.Model(&domain.BlogPostEntity{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"slug": slug, "legacy_permalink": legacyPermalink,
	}).Error
}

func (Repo *Repository) PublishBlogPost(postID uint) (*domain.BlogPostEntity, error) {
	var post domain.BlogPostEntity
	err := Repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&post, postID).Error; err != nil {
			return err
		}
		if post.DraftRevisionID == 0 {
			return errors.New("missing draft revision")
		}
		now := time.Now()
		post.Status = domain.BlogStatusPublished
		post.PublishedRevisionID = post.DraftRevisionID
		post.PublishedAt = &now
		return tx.Save(&post).Error
	})
	return &post, err
}

func (Repo *Repository) SetBlogPostStatus(postID uint, status string) error {
	return Repo.DB.Model(&domain.BlogPostEntity{}).Where("id = ?", postID).Update("status", status).Error
}

func (Repo *Repository) RemoveBlogPost(postID uint) error {
	return Repo.DB.Delete(&domain.BlogPostEntity{}, postID).Error
}

func (Repo *Repository) SaveBlogAsset(asset *domain.BlogAssetEntity) error {
	if asset.ID == 0 {
		return Repo.DB.Create(asset).Error
	}
	return Repo.DB.Save(asset).Error
}

func (Repo *Repository) FindBlogAssetByID(assetID uint) (*domain.BlogAssetEntity, error) {
	var asset domain.BlogAssetEntity
	err := Repo.DB.Where("id = ?", assetID).Limit(1).Find(&asset).Error
	return &asset, err
}

func (Repo *Repository) RemoveBlogAsset(assetID uint) error {
	return Repo.DB.Delete(&domain.BlogAssetEntity{}, assetID).Error
}
