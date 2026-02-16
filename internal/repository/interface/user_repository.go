package repository

import (
	"context"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
)

// UserRepository defines the interface for user data operations
// This interface belongs to the domain layer and doesn't depend on infrastructure
type UserRepository interface {
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (*entity.User, error)
	
	// FindByEmailOrUsername finds a user by email or username
	FindByEmailOrUsername(ctx context.Context, email, username string) (*entity.User, error)
	
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) (string, error)
	
	// Update updates user fields
	Update(ctx context.Context, id string, user *entity.User) error
	
	// Delete deletes a user
	Delete(ctx context.Context, id string) error
	
	// FindAll finds all users
	FindAll(ctx context.Context) ([]*entity.User, error)
	
	// CountUsers counts total number of users
	CountUsers(ctx context.Context) (int64, error)
	
	// CountAdmins counts number of admin users
	CountAdmins(ctx context.Context) (int64, error)
}
