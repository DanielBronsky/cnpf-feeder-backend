package repository

import (
	"context"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
)

// ReportRepository defines the interface for report data operations
type ReportRepository interface {
	// Create creates a new report
	Create(ctx context.Context, report *entity.Report) (string, error)
	
	// FindByID finds a report by ID
	FindByID(ctx context.Context, id string) (*entity.Report, error)
	
	// FindAll finds all reports with limit
	FindAll(ctx context.Context, limit int) ([]*entity.Report, error)
	
	// Update updates a report
	Update(ctx context.Context, id string, report *entity.Report) error
	
	// Delete deletes a report
	Delete(ctx context.Context, id string) error
	
	// GetAuthorID gets author ID of a report
	GetAuthorID(ctx context.Context, id string) (string, error)
}
