package repository

import (
	"context"
	"time"
)

// PasswordResetRepository defines interface for password reset tokens
type PasswordResetRepository interface {
	Save(ctx context.Context, userID, token string, expiresAt time.Time) error
	GetByToken(ctx context.Context, token string) (userID string, err error)
	DeleteByToken(ctx context.Context, token string) error
}
