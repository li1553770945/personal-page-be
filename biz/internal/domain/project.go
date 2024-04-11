package domain

import (
	"personal-page-be/biz/internal/do"
)

type ProjectEntity struct {
	do.BaseModel
	Name string `json:"file_name"`
	Desc string `json:"desc"`
	Link string `json:"link"`
	Date string `json:"date"`
}
