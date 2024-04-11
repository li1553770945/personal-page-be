package domain

import (
	"personal-page-be/biz/internal/do"
)

type ProjectEntity struct {
	do.BaseModel
	Name   string `vd:"len($)>5" json:"name"`
	Desc   string `vd:"len($)>5" json:"desc"`
	Status string `vd:"len($)>2" json:"status"`
	Link   string `vd:"len($)>5" json:"link"`
	Period string `vd:"len($)>5" json:"period"`
}
