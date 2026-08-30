package dto

type SaveBlogPostReq struct {
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	ContentMarkdown  string   `json:"contentMarkdown"`
	Cover            string   `json:"cover"`
	CoverObjectPath  string   `json:"coverObjectPath"`
	Categories       []string `json:"categories"`
	Tags             []string `json:"tags"`
	ChangeSummary    string   `json:"changeSummary"`
	BaseRevisionID   uint     `json:"baseRevisionId"`
	LegacyPermalink  string   `json:"legacyPermalink"`
	PublishAfterSave bool     `json:"publishAfterSave"`
}

type BlogPostDTO struct {
	DatabaseID          uint     `json:"databaseId"`
	Slug                string   `json:"slug"`
	LegacyPermalink     string   `json:"legacyPermalink,omitempty"`
	Status              string   `json:"status"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	ContentMarkdown     string   `json:"contentMarkdown,omitempty"`
	Cover               string   `json:"cover,omitempty"`
	CoverObjectPath     string   `json:"coverObjectPath,omitempty"`
	Categories          []string `json:"categories"`
	Tags                []string `json:"tags"`
	DraftRevisionID     uint     `json:"draftRevisionId,omitempty"`
	PublishedRevisionID uint     `json:"publishedRevisionId,omitempty"`
	RevisionID          uint     `json:"revisionId,omitempty"`
	Version             int      `json:"version,omitempty"`
	ChangeSummary       string   `json:"changeSummary,omitempty"`
	AuthorUsername      string   `json:"authorUsername,omitempty"`
	CreatedAt           int64    `json:"createdAt"`
	UpdatedAt           int64    `json:"updatedAt"`
	PublishedAt         int64    `json:"publishedAt,omitempty"`
}

type BlogPostListDTO struct {
	Items    []*BlogPostDTO `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type BlogRevisionDTO struct {
	RevisionID      uint     `json:"revisionId"`
	PostID          uint     `json:"postId"`
	Version         int      `json:"version"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	ContentMarkdown string   `json:"contentMarkdown"`
	Cover           string   `json:"cover"`
	CoverObjectPath string   `json:"coverObjectPath,omitempty"`
	Categories      []string `json:"categories"`
	Tags            []string `json:"tags"`
	ChangeSummary   string   `json:"changeSummary"`
	AuthorUsername  string   `json:"authorUsername"`
	CreatedAt       int64    `json:"createdAt"`
}

type SignBlogAssetReq struct {
	PostID   uint   `json:"postId"`
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type ConfirmBlogAssetReq struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Alt    string `json:"alt"`
}

type BlogAssetDTO struct {
	ID         uint   `json:"id"`
	PostID     uint   `json:"postId"`
	ObjectPath string `json:"objectPath"`
	FileName   string `json:"fileName"`
	MimeType   string `json:"mimeType"`
	Size       int64  `json:"size"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Alt        string `json:"alt"`
	URL        string `json:"url"`
	SignedURL  string `json:"signedUrl,omitempty"`
	Ready      bool   `json:"ready"`
}
