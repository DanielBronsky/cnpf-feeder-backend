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

// CompetitionRepository handles competition database operations
// Implements repository.CompetitionRepository interface
type CompetitionRepository struct {
	db *mongo.Database
}

// NewCompetitionRepository creates a new competition repository
func NewCompetitionRepository(db *mongo.Database) repository.CompetitionRepository {
	return &CompetitionRepository{db: db}
}

// Ensure CompetitionRepository implements repository.CompetitionRepository interface
var _ repository.CompetitionRepository = (*CompetitionRepository)(nil)

// CompetitionDocument represents a competition document in MongoDB (internal to this package)
type CompetitionDocument struct {
	ID               primitive.ObjectID   `bson:"_id"`
	Title            string               `bson:"title"`
	StartDate        primitive.DateTime   `bson:"startDate"`
	EndDate          primitive.DateTime   `bson:"endDate"`
	Location         string               `bson:"location"`
	Tours            bson.A               `bson:"tours"`
	OpeningDate      *primitive.DateTime  `bson:"openingDate,omitempty"`
	OpeningTime      *string              `bson:"openingTime,omitempty"`
	IndividualFormat bool                 `bson:"individualFormat"`
	TeamFormat       bool                 `bson:"teamFormat"`
	Fee              *float64             `bson:"fee,omitempty"`
	TeamLimit        *int32               `bson:"teamLimit,omitempty"`
	Regulations      *string              `bson:"regulations,omitempty"`
	CreatedAt        primitive.DateTime   `bson:"createdAt"`
	UpdatedAt        primitive.DateTime   `bson:"updatedAt"`
}

// toEntity converts MongoDB document to domain entity
func (doc *CompetitionDocument) toEntity() *entity.Competition {
	startDate := doc.StartDate.Time()
	endDate := doc.EndDate.Time()
	
	tours := make([]entity.Tour, len(doc.Tours))
	for i, tourRaw := range doc.Tours {
		if tourDoc, ok := tourRaw.(bson.M); ok {
			tourDate := tourDoc["date"].(primitive.DateTime).Time()
			tourTime := tourDoc["time"].(string)
			tours[i] = entity.Tour{
				Date: tourDate,
				Time: tourTime,
			}
		}
	}
	
	var openingDate *time.Time
	if doc.OpeningDate != nil {
		t := doc.OpeningDate.Time()
		openingDate = &t
	}
	
	return &entity.Competition{
		ID:               doc.ID.Hex(),
		Title:            doc.Title,
		StartDate:        &startDate,
		EndDate:          &endDate,
		Location:         doc.Location,
		Tours:            tours,
		OpeningDate:      openingDate,
		OpeningTime:      doc.OpeningTime,
		IndividualFormat: doc.IndividualFormat,
		TeamFormat:       doc.TeamFormat,
		Fee:              doc.Fee,
		TeamLimit:        func() *int { if doc.TeamLimit != nil { v := int(*doc.TeamLimit); return &v }; return nil }(),
		Regulations:      doc.Regulations,
		CreatedAt:        doc.CreatedAt.Time(),
		UpdatedAt:        doc.UpdatedAt.Time(),
	}
}

// fromEntity converts domain entity to MongoDB document
func competitionFromEntity(competition *entity.Competition) (*CompetitionDocument, error) {
	competitionID := primitive.NilObjectID
	if competition.ID != "" {
		var err error
		competitionID, err = primitive.ObjectIDFromHex(competition.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid competition ID: %w", err)
		}
	}
	
	if competition.StartDate == nil {
		return nil, fmt.Errorf("startDate is required")
	}
	if competition.EndDate == nil {
		return nil, fmt.Errorf("endDate is required")
	}
	
	startDate := primitive.NewDateTimeFromTime(*competition.StartDate)
	endDate := primitive.NewDateTimeFromTime(*competition.EndDate)
	
	tours := bson.A{}
	for _, tour := range competition.Tours {
		tours = append(tours, bson.M{
			"date": primitive.NewDateTimeFromTime(tour.Date),
			"time": tour.Time,
		})
	}
	
	var openingDate *primitive.DateTime
	if competition.OpeningDate != nil {
		dt := primitive.NewDateTimeFromTime(*competition.OpeningDate)
		openingDate = &dt
	}
	
	var teamLimit *int32
	if competition.TeamLimit != nil {
		t := int32(*competition.TeamLimit)
		teamLimit = &t
	}
	
	now := primitive.NewDateTimeFromTime(time.Now())
	createdAt := now
	if !competition.CreatedAt.IsZero() {
		createdAt = primitive.NewDateTimeFromTime(competition.CreatedAt)
	}
	updatedAt := now
	if !competition.UpdatedAt.IsZero() {
		updatedAt = primitive.NewDateTimeFromTime(competition.UpdatedAt)
	}
	
	return &CompetitionDocument{
		ID:               competitionID,
		Title:            competition.Title,
		StartDate:        startDate,
		EndDate:          endDate,
		Location:         competition.Location,
		Tours:            tours,
		OpeningDate:      openingDate,
		OpeningTime:      competition.OpeningTime,
		IndividualFormat: competition.IndividualFormat,
		TeamFormat:       competition.TeamFormat,
		Fee:              competition.Fee,
		TeamLimit:        teamLimit,
		Regulations:      competition.Regulations,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// Create creates a new competition
func (r *CompetitionRepository) Create(ctx context.Context, competition *entity.Competition) (string, error) {
	doc, err := competitionFromEntity(competition)
	if err != nil {
		return "", err
	}
	
	result, err := r.db.Collection("competitions").InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// FindByID finds a competition by ID
func (r *CompetitionRepository) FindByID(ctx context.Context, id string) (*entity.Competition, error) {
	competitionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid competition ID: %w", err)
	}
	
	var doc CompetitionDocument
	err = r.db.Collection("competitions").FindOne(ctx, bson.M{"_id": competitionID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.toEntity(), nil
}

// FindAll finds all competitions
func (r *CompetitionRepository) FindAll(ctx context.Context) ([]*entity.Competition, error) {
	cursor, err := r.db.Collection("competitions").Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{"createdAt", -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var docs []CompetitionDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	
	competitions := make([]*entity.Competition, len(docs))
	for i, doc := range docs {
		competitions[i] = doc.toEntity()
	}
	return competitions, nil
}

// Update updates a competition
func (r *CompetitionRepository) Update(ctx context.Context, id string, competition *entity.Competition) error {
	competitionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid competition ID: %w", err)
	}
	
	doc, err := competitionFromEntity(competition)
	if err != nil {
		return err
	}
	
	update := bson.M{
		"title":            doc.Title,
		"startDate":        doc.StartDate,
		"endDate":          doc.EndDate,
		"location":         doc.Location,
		"tours":            doc.Tours,
		"individualFormat": doc.IndividualFormat,
		"teamFormat":       doc.TeamFormat,
		"updatedAt":        doc.UpdatedAt,
	}
	
	if doc.OpeningDate != nil {
		update["openingDate"] = doc.OpeningDate
	} else {
		update["openingDate"] = nil
	}
	if doc.OpeningTime != nil {
		update["openingTime"] = doc.OpeningTime
	} else {
		update["openingTime"] = nil
	}
	if doc.Fee != nil {
		update["fee"] = doc.Fee
	} else {
		update["fee"] = nil
	}
	if doc.TeamLimit != nil {
		update["teamLimit"] = doc.TeamLimit
	} else {
		update["teamLimit"] = nil
	}
	if doc.Regulations != nil {
		update["regulations"] = doc.Regulations
	} else {
		update["regulations"] = nil
	}
	
	_, err = r.db.Collection("competitions").UpdateOne(ctx, bson.M{"_id": competitionID}, bson.M{"$set": update})
	return err
}

// Delete deletes a competition
func (r *CompetitionRepository) Delete(ctx context.Context, id string) error {
	competitionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid competition ID: %w", err)
	}
	
	result, err := r.db.Collection("competitions").DeleteOne(ctx, bson.M{"_id": competitionID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("competition not found")
	}
	return nil
}
