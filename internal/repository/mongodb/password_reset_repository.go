package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cnpf/feeder-backend/internal/repository/interface"
)

const passwordResetCollection = "password_reset_tokens"

// PasswordResetRepository implements repository.PasswordResetRepository
type PasswordResetRepository struct {
	db *mongo.Database
}

// NewPasswordResetRepository creates a new password reset repository
func NewPasswordResetRepository(db *mongo.Database) repository.PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

var _ repository.PasswordResetRepository = (*PasswordResetRepository)(nil)

// Save stores a reset token for a user
func (r *PasswordResetRepository) Save(ctx context.Context, userID, token string, expiresAt time.Time) error {
	doc := bson.M{
		"token":     token,
		"userId":    userID,
		"expiresAt": primitive.NewDateTimeFromTime(expiresAt),
	}
	_, err := r.db.Collection(passwordResetCollection).InsertOne(ctx, doc)
	return err
}

// GetByToken returns userID if token is valid and not expired
func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (string, error) {
	var doc struct {
		UserID    string             `bson:"userId"`
		ExpiresAt primitive.DateTime `bson:"expiresAt"`
	}
	err := r.db.Collection(passwordResetCollection).FindOne(ctx, bson.M{
		"token":     token,
		"expiresAt": bson.M{"$gt": primitive.NewDateTimeFromTime(time.Now())},
	}).Decode(&doc)
	if err != nil {
		return "", err
	}
	return doc.UserID, nil
}

// DeleteByToken removes a used token
func (r *PasswordResetRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Collection(passwordResetCollection).DeleteOne(ctx, bson.M{"token": token})
	return err
}
