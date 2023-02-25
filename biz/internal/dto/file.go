package dto

type UploadFileReq struct {
	FileKey string `json:"file_key"`
	Count   int    `json:"count"`
}
