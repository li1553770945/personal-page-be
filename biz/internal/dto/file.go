package dto

type UploadFileReq struct {
	FileKey string `json:"file_key"`
	Count   int    `json:"count"`
}

type SignedUploadFileReq struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	MaxDownload int32  `json:"maxDownload"`
}

type DeleteFileReq struct {
	Key string `json:"key"`
}

type FileDTO struct {
	ID            uint   `json:"id"`
	UserID        int    `json:"user_id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Name          string `json:"name"`
	Key           string `json:"key"`
	Kind          string `json:"kind"`
	Count         int    `json:"count"`
	MaxDownload   int32  `json:"max_download"`
	DownloadCount int32  `json:"download_count"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}
