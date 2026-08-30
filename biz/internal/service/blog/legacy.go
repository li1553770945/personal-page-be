package blog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"personal-page-be/biz/internal/do"
	"personal-page-be/biz/internal/domain"
)

//go:embed legacy_posts.json
var legacyPostsJSON []byte

const legacyVuePressMigrationKey = "vuepress-2026-08-30-v1"

type legacyPostSeed struct {
	Slug            string   `json:"slug"`
	LegacyPermalink string   `json:"legacyPermalink"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	ContentMarkdown string   `json:"contentMarkdown"`
	Categories      []string `json:"categories"`
	Tags            []string `json:"tags"`
	PublishedAt     string   `json:"publishedAt"`
	SourcePath      string   `json:"sourcePath"`
}

// importLegacyPosts is an idempotent, startup-time data migration. Keeping the
// source snapshot embedded in the binary means the first deployment can create
// the new tables and migrate the old VuePress content in one rollout.
func (s *BlogService) importLegacyPosts() {
	var seeds []legacyPostSeed
	if err := json.Unmarshal(legacyPostsJSON, &seeds); err != nil {
		panic("decode embedded legacy blog posts failed: " + err.Error())
	}

	completed, err := s.Repo.HasBlogMigration(legacyVuePressMigrationKey)
	if err != nil {
		panic("check legacy blog migration failed: " + err.Error())
	}
	if completed {
		return
	}

	author, err := s.legacyImportAuthor()
	if err != nil {
		panic("resolve legacy blog author failed: " + err.Error())
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*60*60)
	}
	for _, seed := range seeds {
		existing, findErr := s.Repo.FindBlogPostBySlug(seed.Slug)
		if findErr != nil {
			panic("find legacy blog post failed: " + findErr.Error())
		}
		if existing.ID != 0 {
			continue
		}

		publishedAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", seed.PublishedAt, location)
		if parseErr != nil {
			panic(fmt.Sprintf("parse publish time for legacy post %s failed: %v", seed.Slug, parseErr))
		}
		post := &domain.BlogPostEntity{
			BaseModel:       do.BaseModel{CreatedAt: publishedAt, UpdatedAt: publishedAt},
			Slug:            seed.Slug,
			LegacyPermalink: normalizeLegacyPermalink(seed.LegacyPermalink),
			Status:          domain.BlogStatusPublished,
		}
		revision := &domain.BlogRevisionEntity{
			BaseModel:       do.BaseModel{CreatedAt: publishedAt, UpdatedAt: publishedAt},
			Title:           seed.Title,
			Description:     seed.Description,
			ContentMarkdown: seed.ContentMarkdown,
			Categories:      encodeList(seed.Categories),
			Tags:            encodeList(seed.Tags),
			ChangeSummary:   "从旧博客迁移：" + seed.SourcePath,
			AuthorID:        author.ID,
			AuthorUsername:  author.Username,
		}
		if createErr := s.Repo.CreateBlogPost(post, revision, true); createErr != nil {
			panic(fmt.Sprintf("import legacy blog post %s failed: %v", seed.Slug, createErr))
		}
	}
	if err = s.Repo.SaveBlogMigration(legacyVuePressMigrationKey); err != nil {
		panic("save legacy blog migration marker failed: " + err.Error())
	}
}

func (s *BlogService) legacyImportAuthor() (*domain.UserEntity, error) {
	if username := s.Config.EffectiveSuperAdminUsername(); username != "" {
		user, err := s.Repo.FindUser(username)
		if err != nil {
			return nil, err
		}
		if user.ID != 0 {
			return user, nil
		}
	}
	user, err := s.Repo.FindFirstUserByRole(domain.RoleSuperAdmin)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("no super administrator exists")
	}
	return user, nil
}
