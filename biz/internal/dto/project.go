package dto

type AddProjectReq struct {
	Name string `json:"file_name"`
	Desc string `json:"desc"`
	Link string `json:"link"`
	Date string `json:"date"`
}
