package entity

import "time"

// User represents a user domain entity
// This is the core domain entity - it doesn't depend on anything
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	IsAdmin      bool
	HasAvatar    bool
	Avatar       map[string]interface{} // Avatar data (can be nil)
	CreatedAt    time.Time
}
