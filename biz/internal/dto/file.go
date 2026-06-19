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
