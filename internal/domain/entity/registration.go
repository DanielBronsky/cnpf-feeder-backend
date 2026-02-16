package entity

import "time"

// RegistrationType represents the type of registration
type RegistrationType string

const (
	RegistrationTypeIndividual RegistrationType = "individual"
	RegistrationTypeTeam       RegistrationType = "team"
)

// Registration represents a competition registration domain entity
type Registration struct {
	ID              string
	CompetitionID   string
	UserID          string // User who created the registration
	Type            RegistrationType
	TeamName        *string // Only for team registrations
	Participants    []Participant
	Coach           *Coach // Optional coach for team registrations
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Participant represents a participant in a registration
type Participant struct {
	FirstName string
	LastName  string
}

// Coach represents a coach for team registration
type Coach struct {
	FirstName string
	LastName  string
}
