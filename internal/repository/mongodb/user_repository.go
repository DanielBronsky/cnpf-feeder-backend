package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
	"github.com/cnpf/feeder-backend/internal/repository/interface"
)

// UserRepository handles user database operations
// Implements repository.UserRepository interface
type UserRepository struct {
	db *mongo.Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database) repository.UserRepository {
	return &UserRepository{db: db}
}

// Ensure UserRepository implements repository.UserRepository interface
var _ repository.UserRepository = (*UserRepository)(nil)

// UserDocument represents a user document in MongoDB (internal to this package)
type UserDocument struct {
	ID           primitive.ObjectID `bson:"_id"`
	Email        string             `bson:"email"`
	Username     string             `bson:"username"`
	PasswordHash string             `bson:"passwordHash"`
	IsAdmin      bool               `bson:"isAdmin"`
	HasAvatar    bool               `bson:"hasAvatar"`
	Avatar       bson.M              `bson:"avatar,omitempty"`
	CreatedAt    primitive.DateTime  `bson:"createdAt"`
}

// toEntity converts MongoDB document to domain entity
func (doc *UserDocument) toEntity() *entity.User {
	avatar := make(map[string]interface{})
	if doc.Avatar != nil {
		for k, v := range doc.Avatar {
			avatar[k] = v
		}
	}
	
	return &entity.User{
		ID:           doc.ID.Hex(),
		Email:        doc.Email,
		Username:     doc.Username,
		PasswordHash: doc.PasswordHash,
		IsAdmin:      doc.IsAdmin,
		HasAvatar:    doc.HasAvatar,
		Avatar:       avatar,
		CreatedAt:    doc.CreatedAt.Time(),
	}
}

// fromEntity converts domain entity to MongoDB document
func fromEntity(user *entity.User) (*UserDocument, error) {
	userID := primitive.NilObjectID
	if user.ID != "" {
		var err error
		userID, err = primitive.ObjectIDFromHex(user.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
	}
	
	var avatar bson.M
	if len(user.Avatar) > 0 {
		avatar = bson.M{}
		for k, v := range user.Avatar {
			avatar[k] = v
		}
	}
	
	return &UserDocument{
		ID:           userID,
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		IsAdmin:      user.IsAdmin,
		HasAvatar:    user.HasAvatar,
		Avatar:       avatar,
		CreatedAt:    primitive.NewDateTimeFromTime(user.CreatedAt),
	}, nil
}

// FindByEmailOrUsername finds a user by email or username
func (r *UserRepository) FindByEmailOrUsername(ctx context.Context, email, username string) (*entity.User, error) {
	var doc UserDocument
	err := r.db.Collection("users").FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": username},
		},
	}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.toEntity(), nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	var doc UserDocument
	err = r.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.toEntity(), nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) (string, error) {
	doc, err := fromEntity(user)
	if err != nil {
		return "", err
	}
	
	// Set createdAt if not set
	if doc.CreatedAt == 0 {
		doc.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	}
	
	result, err := r.db.Collection("users").InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Update updates user fields
func (r *UserRepository) Update(ctx context.Context, id string, user *entity.User) error {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	
	doc, err := fromEntity(user)
	if err != nil {
		return err
	}
	
	update := bson.M{
		"email":        doc.Email,
		"username":     doc.Username,
		"passwordHash": doc.PasswordHash,
		"isAdmin":      doc.IsAdmin,
		"hasAvatar":    doc.HasAvatar,
	}
	
	if doc.Avatar != nil {
		update["avatar"] = doc.Avatar
	}
	
	_, err = r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": update})
	return err
}

// FindAll finds all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	cursor, err := r.db.Collection("users").Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var docs []UserDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	
	users := make([]*entity.User, len(docs))
	for i, doc := range docs {
		users[i] = doc.toEntity()
	}
	return users, nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	
	_, err = r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": userID})
	return err
}

// CountUsers counts total number of users
func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.db.Collection("users").EstimatedDocumentCount(ctx)
}

// CountAdmins counts number of admin users
func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	return r.db.Collection("users").CountDocuments(ctx, bson.M{"isAdmin": true})
}

// TEMPORARY: Legacy methods for backward compatibility during migration
// TODO: Remove after migrating all resolvers to use entity-based methods

// CreateWithAvatar creates a new user with optional avatar (legacy method)
func (r *UserRepository) CreateWithAvatar(ctx context.Context, email, username, passwordHash string, isAdmin bool, avatar bson.M, hasAvatar bool) (string, error) {
	user := &entity.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
		HasAvatar:    hasAvatar,
		CreatedAt:    time.Now(),
	}
	
	if hasAvatar && avatar != nil {
		avatarMap := make(map[string]interface{})
		for k, v := range avatar {
			avatarMap[k] = v
		}
		user.Avatar = avatarMap
	}
	
	return r.Create(ctx, user)
}

// UpdateLegacy updates user fields using bson.M (legacy method)
func (r *UserRepository) UpdateLegacy(ctx context.Context, id string, update bson.M) error {
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	
	_, err = r.db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": update})
	return err
}
