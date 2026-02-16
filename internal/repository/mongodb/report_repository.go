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

// ReportRepository handles report database operations
// Implements repository.ReportRepository interface
type ReportRepository struct {
	db *mongo.Database
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *mongo.Database) repository.ReportRepository {
	return &ReportRepository{db: db}
}

// Ensure ReportRepository implements repository.ReportRepository interface
var _ repository.ReportRepository = (*ReportRepository)(nil)

// ReportDocument represents a report document in MongoDB (internal to this package)
type ReportDocument struct {
	ID        primitive.ObjectID `bson:"_id"`
	AuthorID  primitive.ObjectID `bson:"authorId"`
	Title     string             `bson:"title"`
	Text      string             `bson:"text"`
	Photos    bson.A             `bson:"photos"`
	CreatedAt primitive.DateTime `bson:"createdAt"`
	UpdatedAt primitive.DateTime `bson:"updatedAt"`
}

// toEntity converts MongoDB document to domain entity
func (doc *ReportDocument) toEntity() *entity.Report {
	photos := make([]interface{}, len(doc.Photos))
	for i, photo := range doc.Photos {
		photos[i] = photo
	}
	
	return &entity.Report{
		ID:        doc.ID.Hex(),
		AuthorID:  doc.AuthorID.Hex(),
		Title:     doc.Title,
		Text:      doc.Text,
		Photos:    photos,
		CreatedAt: doc.CreatedAt.Time(),
		UpdatedAt: doc.UpdatedAt.Time(),
	}
}

// fromEntity converts domain entity to MongoDB document
func reportFromEntity(report *entity.Report) (*ReportDocument, error) {
	reportID := primitive.NilObjectID
	if report.ID != "" {
		var err error
		reportID, err = primitive.ObjectIDFromHex(report.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid report ID: %w", err)
		}
	}
	
	authorID, err := primitive.ObjectIDFromHex(report.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author ID: %w", err)
	}
	
	photos := bson.A{}
	for _, photo := range report.Photos {
		photos = append(photos, photo)
	}
	
	now := primitive.NewDateTimeFromTime(time.Now())
	createdAt := now
	if !report.CreatedAt.IsZero() {
		createdAt = primitive.NewDateTimeFromTime(report.CreatedAt)
	}
	updatedAt := now
	if !report.UpdatedAt.IsZero() {
		updatedAt = primitive.NewDateTimeFromTime(report.UpdatedAt)
	}
	
	return &ReportDocument{
		ID:        reportID,
		AuthorID:  authorID,
		Title:     report.Title,
		Text:      report.Text,
		Photos:    photos,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// Create creates a new report
func (r *ReportRepository) Create(ctx context.Context, report *entity.Report) (string, error) {
	doc, err := reportFromEntity(report)
	if err != nil {
		return "", err
	}
	
	result, err := r.db.Collection("reports").InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// FindByID finds a report by ID
func (r *ReportRepository) FindByID(ctx context.Context, id string) (*entity.Report, error) {
	reportID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID: %w", err)
	}
	
	var doc ReportDocument
	err = r.db.Collection("reports").FindOne(ctx, bson.M{"_id": reportID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.toEntity(), nil
}

// FindAll finds all reports with limit
func (r *ReportRepository) FindAll(ctx context.Context, limit int) ([]*entity.Report, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "authorId"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "author"},
		}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$author"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
	}
	
	cursor, err := r.db.Collection("reports").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var docs []ReportDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	
	reports := make([]*entity.Report, len(docs))
	for i, doc := range docs {
		reports[i] = doc.toEntity()
	}
	return reports, nil
}

// Update updates a report
func (r *ReportRepository) Update(ctx context.Context, id string, report *entity.Report) error {
	reportID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}
	
	doc, err := reportFromEntity(report)
	if err != nil {
		return err
	}
	
	update := bson.M{
		"title":     doc.Title,
		"text":      doc.Text,
		"photos":    doc.Photos,
		"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
	}
	
	_, err = r.db.Collection("reports").UpdateOne(ctx, bson.M{"_id": reportID}, bson.M{"$set": update})
	return err
}

// Delete deletes a report
func (r *ReportRepository) Delete(ctx context.Context, id string) error {
	reportID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid report ID: %w", err)
	}
	
	_, err = r.db.Collection("reports").DeleteOne(ctx, bson.M{"_id": reportID})
	return err
}

// GetAuthorID gets author ID of a report
func (r *ReportRepository) GetAuthorID(ctx context.Context, id string) (string, error) {
	reportID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", fmt.Errorf("invalid report ID: %w", err)
	}
	
	var result struct {
		AuthorID primitive.ObjectID `bson:"authorId"`
	}
	err = r.db.Collection("reports").FindOne(ctx, bson.M{"_id": reportID}, options.FindOne().SetProjection(bson.M{"authorId": 1})).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.AuthorID.Hex(), nil
}
