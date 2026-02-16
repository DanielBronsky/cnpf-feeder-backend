package entity

import "time"

// Tour represents a tour in a competition
type Tour struct {
	Date time.Time
	Time string
}

// Competition represents a competition domain entity
type Competition struct {
	ID               string
	Title            string
	StartDate        *time.Time
	EndDate          *time.Time
	Location         string
	Tours            []Tour
	OpeningDate      *time.Time
	OpeningTime      *string
	IndividualFormat bool
	TeamFormat       bool
	Fee              *float64
	TeamLimit        *int
	Regulations      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
