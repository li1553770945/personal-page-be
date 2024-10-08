package domain

import (
	"personal-page-be/biz/internal/do"
	"time"
)

const (
	Completed int = iota + 1
	InProgress
	Discontinued
)

type ProjectEntity struct {
	do.BaseModel
	Name         string     `vd:"len($)>5" json:"name"`
	Desc         string     `vd:"len($)>5" json:"desc"`
	Link         string     `vd:"len($)>5" json:"link"`
	Status       int        `vd:"$>=1 && $<=3" json:"status"`
	VolumeOfWork int        `vd:"$>=1 && $<=5" json:"volume_of_work"`
	Difficulty   int        `vd:"$>=1 && $<=5" json:"difficulty"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
}
