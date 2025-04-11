package service

import "time"

// TaskLog in C#
type taskLog struct {
	Id            string    `json:"id"`
	Location      string    `json:"location"`
	IndexLocation string    `json:"index_location"`
	Path          string    `json:"path"`
	LineCount     int64     `json:"line_count"`
	CreatedOn     time.Time `json:"created_on"`
	LastChangedOn time.Time `json:"last_changed_on"`
}
