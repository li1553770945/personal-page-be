package dto

type SaveSlideReq struct {
	Slug            string   `json:"id"`
	Title           string   `json:"title"`
	TitleEn         string   `json:"titleEn"`
	Description     string   `json:"description"`
	DescriptionEn   string   `json:"descriptionEn"`
	Cover           string   `json:"cover"`
	CoverObjectPath string   `json:"coverObjectPath"`
	Entry           string   `json:"entry"`
	ObjectPrefix    string   `json:"objectPrefix"`
	Tags            []string `json:"tags"`
	Protected       bool     `json:"protected"`
	Password        string   `json:"password"`
}

type UnlockSlideReq struct {
	Password string `json:"password"`
}

type SlideDTO struct {
	DatabaseID      uint     `json:"database_id"`
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	TitleEn         string   `json:"titleEn"`
	Description     string   `json:"description"`
	DescriptionEn   string   `json:"descriptionEn"`
	Cover           string   `json:"cover"`
	CoverObjectPath string   `json:"coverObjectPath,omitempty"`
	Entry           string   `json:"entry"`
	ObjectPrefix    string   `json:"objectPrefix"`
	Tags            []string `json:"tags"`
	Protected       bool     `json:"protected"`
	Password        string   `json:"password,omitempty"`
	HasPassword     bool     `json:"has_password"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
}

type SlideUploadDTO struct {
	ID              string `json:"id,omitempty"`
	Entry           string `json:"entry,omitempty"`
	ObjectPrefix    string `json:"objectPrefix,omitempty"`
	Cover           string `json:"cover,omitempty"`
	CoverObjectPath string `json:"coverObjectPath,omitempty"`
	FileCount       int    `json:"fileCount,omitempty"`
}
