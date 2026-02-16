package entity

import "time"

// Report represents a report domain entity
type Report struct {
	ID        string
	AuthorID  string
	Title     string
	Text      string
	Photos    []interface{} // Photo data
	CreatedAt time.Time
	UpdatedAt time.Time
}
