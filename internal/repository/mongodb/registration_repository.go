package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cnpf/feeder-backend/internal/domain/entity"
	"github.com/cnpf/feeder-backend/internal/repository/interface"
)

// RegistrationRepository handles registration database operations
// Implements repository.RegistrationRepository interface
type RegistrationRepository struct {
	db *mongo.Database
}

// NewRegistrationRepository creates a new registration repository
func NewRegistrationRepository(db *mongo.Database) repository.RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// Ensure RegistrationRepository implements repository.RegistrationRepository interface
var _ repository.RegistrationRepository = (*RegistrationRepository)(nil)

// RegistrationDocument represents a registration document in MongoDB
type RegistrationDocument struct {
	ID            primitive.ObjectID `bson:"_id"`
	CompetitionID primitive.ObjectID `bson:"competitionId"`
	UserID        primitive.ObjectID `bson:"userId"`
	Type          string             `bson:"type"`
	TeamName      *string            `bson:"teamName,omitempty"`
	Participants  []ParticipantDoc   `bson:"participants"`
	Coach         *CoachDoc           `bson:"coach,omitempty"`
	CreatedAt     primitive.DateTime `bson:"createdAt"`
	UpdatedAt     primitive.DateTime `bson:"updatedAt"`
}

type ParticipantDoc struct {
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}

type CoachDoc struct {
	FirstName string `bson:"firstName"`
	LastName  string `bson:"lastName"`
}

// toEntity converts MongoDB document to domain entity
func (doc *RegistrationDocument) toEntity() *entity.Registration {
	participants := make([]entity.Participant, len(doc.Participants))
	for i, p := range doc.Participants {
		participants[i] = entity.Participant{
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}

	var coach *entity.Coach
	if doc.Coach != nil {
		coach = &entity.Coach{
			FirstName: doc.Coach.FirstName,
			LastName:  doc.Coach.LastName,
		}
	}

	return &entity.Registration{
		ID:            doc.ID.Hex(),
		CompetitionID: doc.CompetitionID.Hex(),
		UserID:        doc.UserID.Hex(),
		Type:          entity.RegistrationType(doc.Type),
		TeamName:      doc.TeamName,
		Participants:  participants,
		Coach:         coach,
		CreatedAt:     doc.CreatedAt.Time(),
		UpdatedAt:     doc.UpdatedAt.Time(),
	}
}

// Create creates a new registration
func (r *RegistrationRepository) Create(ctx context.Context, reg *entity.Registration) (string, error) {
	competitionID, err := primitive.ObjectIDFromHex(reg.CompetitionID)
	if err != nil {
		return "", fmt.Errorf("invalid competition ID: %w", err)
	}

	userID, err := primitive.ObjectIDFromHex(reg.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	participants := make([]ParticipantDoc, len(reg.Participants))
	for i, p := range reg.Participants {
		participants[i] = ParticipantDoc{
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}

	var coach *CoachDoc
	if reg.Coach != nil {
		coach = &CoachDoc{
			FirstName: reg.Coach.FirstName,
			LastName:  reg.Coach.LastName,
		}
	}

	doc := RegistrationDocument{
		ID:            primitive.NewObjectID(),
		CompetitionID: competitionID,
		UserID:        userID,
		Type:          string(reg.Type),
		TeamName:      reg.TeamName,
		Participants:  participants,
		Coach:         coach,
		CreatedAt:     primitive.NewDateTimeFromTime(reg.CreatedAt),
		UpdatedAt:     primitive.NewDateTimeFromTime(reg.UpdatedAt),
	}

	result, err := r.db.Collection("registrations").InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to create registration: %w", err)
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("unexpected InsertedID type: %T", result.InsertedID)
	}
	return oid.Hex(), nil
}

// FindByID finds a registration by ID
func (r *RegistrationRepository) FindByID(ctx context.Context, id string) (*entity.Registration, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	var doc RegistrationDocument
	err = r.db.Collection("registrations").FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("registration not found")
		}
		return nil, fmt.Errorf("failed to find registration: %w", err)
	}

	return doc.toEntity(), nil
}

// FindByCompetitionID finds all registrations for a competition
func (r *RegistrationRepository) FindByCompetitionID(ctx context.Context, competitionID string) ([]*entity.Registration, error) {
	objID, err := primitive.ObjectIDFromHex(competitionID)
	if err != nil {
		return nil, fmt.Errorf("invalid competition ID: %w", err)
	}

	cursor, err := r.db.Collection("registrations").Find(ctx, bson.M{"competitionId": objID})
	if err != nil {
		return nil, fmt.Errorf("failed to find registrations: %w", err)
	}
	defer cursor.Close(ctx)

	var registrations []*entity.Registration
	for cursor.Next(ctx) {
		var doc RegistrationDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		registrations = append(registrations, doc.toEntity())
	}

	return registrations, nil
}

// FindByUserID finds all registrations by a user
func (r *RegistrationRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Registration, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	cursor, err := r.db.Collection("registrations").Find(ctx, bson.M{"userId": objID})
	if err != nil {
		return nil, fmt.Errorf("failed to find registrations: %w", err)
	}
	defer cursor.Close(ctx)

	var registrations []*entity.Registration
	for cursor.Next(ctx) {
		var doc RegistrationDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		registrations = append(registrations, doc.toEntity())
	}

	return registrations, nil
}

// FindByCompetitionAndUser finds registration for specific competition and user
func (r *RegistrationRepository) FindByCompetitionAndUser(ctx context.Context, competitionID, userID string) (*entity.Registration, error) {
	compID, err := primitive.ObjectIDFromHex(competitionID)
	if err != nil {
		return nil, fmt.Errorf("invalid competition ID: %w", err)
	}

	usrID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var doc RegistrationDocument
	err = r.db.Collection("registrations").FindOne(ctx, bson.M{
		"competitionId": compID,
		"userId":        usrID,
	}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to find registration: %w", err)
	}

	return doc.toEntity(), nil
}

// Update updates a registration
func (r *RegistrationRepository) Update(ctx context.Context, id string, reg *entity.Registration) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	participants := make([]ParticipantDoc, len(reg.Participants))
	for i, p := range reg.Participants {
		participants[i] = ParticipantDoc{
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}

	var coach *CoachDoc
	if reg.Coach != nil {
		coach = &CoachDoc{
			FirstName: reg.Coach.FirstName,
			LastName:  reg.Coach.LastName,
		}
	}

	update := bson.M{
		"$set": bson.M{
			"type":        string(reg.Type),
			"teamName":    reg.TeamName,
			"participants": participants,
			"coach":       coach,
			"updatedAt":   primitive.NewDateTimeFromTime(reg.UpdatedAt),
		},
	}

	result, err := r.db.Collection("registrations").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update registration: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("registration not found")
	}

	return nil
}

// Delete deletes a registration
func (r *RegistrationRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	result, err := r.db.Collection("registrations").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete registration: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("registration not found")
	}

	return nil
}
