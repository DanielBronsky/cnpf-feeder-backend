package repository

import (
	"context"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
)

// RegistrationRepository defines the interface for registration data operations
type RegistrationRepository interface {
	// Create creates a new registration
	Create(ctx context.Context, registration *entity.Registration) (string, error)
	
	// FindByID finds a registration by ID
	FindByID(ctx context.Context, id string) (*entity.Registration, error)
	
	// FindByCompetitionID finds all registrations for a competition
	FindByCompetitionID(ctx context.Context, competitionID string) ([]*entity.Registration, error)
	
	// FindByUserID finds all registrations by a user
	FindByUserID(ctx context.Context, userID string) ([]*entity.Registration, error)
	
	// FindByCompetitionAndUser finds registration for specific competition and user
	FindByCompetitionAndUser(ctx context.Context, competitionID, userID string) (*entity.Registration, error)
	
	// Update updates a registration
	Update(ctx context.Context, id string, registration *entity.Registration) error
	
	// Delete deletes a registration
	Delete(ctx context.Context, id string) error
}
