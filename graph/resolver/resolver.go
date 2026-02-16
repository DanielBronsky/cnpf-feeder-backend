package resolver

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cnpf/feeder-backend/internal/gemini"
	"github.com/cnpf/feeder-backend/internal/repository/interface"
	"github.com/cnpf/feeder-backend/internal/usecase"
)

// Resolver is the root resolver
// 
// TEMPORARY: Currently contains both useCase (correct) and repositories (temporary)
// During migration to Onion Architecture, we need repositories for backward compatibility
// TODO: Remove repositories after all schema.resolvers.go methods are migrated to use useCase
type Resolver struct {
	useCase usecase.UseCase
	
	// TEMPORARY: These will be removed after migration
	// All methods in schema.resolvers.go should use r.useCase instead
	userRepo         repository.UserRepository
	reportRepo       repository.ReportRepository
	competitionRepo  repository.CompetitionRepository
	registrationRepo repository.RegistrationRepository
	
	// Chat dependencies
	db          *mongo.Database
	geminiClient *gemini.Client
}

// NewResolver creates a new resolver
// TEMPORARY: Accepts repositories for backward compatibility during migration
// TODO: Remove repository parameters after migration
func NewResolver(useCase usecase.UseCase, userRepo repository.UserRepository, reportRepo repository.ReportRepository, competitionRepo repository.CompetitionRepository, registrationRepo repository.RegistrationRepository, db *mongo.Database) *Resolver {
	geminiClient, err := gemini.NewClient()
	if err != nil {
		geminiClient = nil // Fallback: chat features disabled if Gemini unavailable
	}

	return &Resolver{
		useCase:         useCase,
		userRepo:        userRepo,
		reportRepo:      reportRepo,
		competitionRepo: competitionRepo,
		registrationRepo: registrationRepo,
		db:              db,
		geminiClient:    geminiClient,
	}
}
