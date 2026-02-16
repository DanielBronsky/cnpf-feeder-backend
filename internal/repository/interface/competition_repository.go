package repository

import (
	"context"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
)

// CompetitionRepository defines the interface for competition data operations
type CompetitionRepository interface {
	// Create creates a new competition
	Create(ctx context.Context, competition *entity.Competition) (string, error)
	
	// FindByID finds a competition by ID
	FindByID(ctx context.Context, id string) (*entity.Competition, error)
	
	// FindAll finds all competitions
	FindAll(ctx context.Context) ([]*entity.Competition, error)
	
	// Update updates a competition
	Update(ctx context.Context, id string, competition *entity.Competition) error
	
	// Delete deletes a competition
	Delete(ctx context.Context, id string) error
}
