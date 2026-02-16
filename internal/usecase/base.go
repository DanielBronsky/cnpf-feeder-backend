package usecase

import (
	"context"
	"io"

	"github.com/cnpf/feeder-backend/graph/model"
	"github.com/cnpf/feeder-backend/internal/auth"
)

// UseCase defines all business operations
type UseCase interface {
	// Auth
	Register(ctx context.Context, email, username, password, passwordConfirm string, avatar *PhotoUpload) (*model.AuthResult, error)
	Login(ctx context.Context, login, password string) (*model.AuthResult, error)
	Logout(ctx context.Context) (bool, error)
	
	// User
	GetCurrentUser(ctx context.Context, userID string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID string, username *string, removeAvatar *bool, avatar io.Reader, avatarSize int64, avatarContentType string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID string, oldPassword, newPassword string) (bool, error)
	
	// Reports
	GetReports(ctx context.Context, currentUserID string, limit *int) ([]*model.Report, error)
	GetReport(ctx context.Context, currentUserID string, id string) (*model.Report, error)
	CreateReport(ctx context.Context, userID string, title, text string, photos []*PhotoUpload) (*model.Report, error)
	UpdateReport(ctx context.Context, userID string, id string, title, text *string, removePhoto []int, removeAllPhotos *bool, photos []*PhotoUpload) (*model.Report, error)
	DeleteReport(ctx context.Context, userID string, id string) (bool, error)
	
	// Competitions
	GetCompetitions(ctx context.Context) ([]*model.Competition, error)
	GetCompetition(ctx context.Context, id string) (*model.Competition, error)
	CreateCompetition(ctx context.Context, input *model.CompetitionInput) (*model.Competition, error)
	UpdateCompetition(ctx context.Context, id string, input *model.CompetitionInput) (*model.Competition, error)
	DeleteCompetition(ctx context.Context, id string) (bool, error)
	
	// Admin
	GetAdminUsers(ctx context.Context) ([]*model.User, error)
	GetAdminUser(ctx context.Context, id string) (*model.User, error)
	AdminUpdateUser(ctx context.Context, id string, isAdmin *bool) (*model.User, error)
	AdminDeleteUser(ctx context.Context, id string) (bool, error)
	
	// Registrations
	CreateRegistration(ctx context.Context, userID string, competitionID string, registrationType string, teamName *string, participants []ParticipantInput, coach *CoachInput) (*model.Registration, error)
	GetRegistrationsByCompetition(ctx context.Context, competitionID string, currentUserID string) ([]*model.Registration, error)
	UpdateRegistration(ctx context.Context, userID string, registrationID string, teamName *string, participants []ParticipantInput, coach *CoachInput) (*model.Registration, error)
	DeleteRegistration(ctx context.Context, userID string, registrationID string) (bool, error)
}

// ParticipantInput represents participant input for registration
type ParticipantInput struct {
	FirstName string
	LastName  string
}

// CoachInput represents coach input for registration
type CoachInput struct {
	FirstName string
	LastName  string
}

// PhotoUpload represents an uploaded photo
type PhotoUpload struct {
	File        io.Reader
	Size        int64
	ContentType string
}

// GetCurrentUserFromContext extracts current user from context
func GetCurrentUserFromContext(ctx context.Context) (*auth.CurrentUser, error) {
	// This will be implemented in resolver layer
	return nil, nil
}
