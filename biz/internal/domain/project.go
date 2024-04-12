package domain

import (
	"personal-page-be/biz/internal/do"
	"time"
)

type ProjectStatus int

const (
	Completed ProjectStatus = iota + 1
	InProgress
	Discontinued
)

type ProjectEntity struct {
	do.BaseModel
	Name      string        `vd:"len($)>5" json:"name"`
	Desc      string        `vd:"len($)>5" json:"desc"`
	Link      string        `vd:"len($)>5" json:"link"`
	Status    ProjectStatus `json:"status"`
	StartDate time.Time     `json:"start_date"`
	EndDate   *time.Time    `json:"end_date,omitempty"`
}
