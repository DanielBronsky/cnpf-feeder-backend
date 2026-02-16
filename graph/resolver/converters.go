package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/cnpf/feeder-backend/graph/model"
	"github.com/cnpf/feeder-backend/graph/scalars"
	"github.com/cnpf/feeder-backend/internal/auth"
	"github.com/cnpf/feeder-backend/internal/domain/entity"
	"github.com/cnpf/feeder-backend/internal/repository/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// formatUser converts MongoDB user document to GraphQL User model
func formatUser(userDoc *mongodb.UserDocument) *model.User {
	userID := userDoc.ID.Hex()
	hasAvatar := userDoc.HasAvatar
	var avatarURL *string
	if hasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", userID)
		avatarURL = &url
	}
	username := userDoc.Username
	if username == "" {
		username = userDoc.Email
	}
	return &model.User{
		ID:        userID,
		Email:     userDoc.Email,
		Username:  username,
		IsAdmin:   userDoc.IsAdmin,
		HasAvatar: hasAvatar,
		AvatarURL: avatarURL,
	}
}

// getBsonString safely extracts string from bson.M
func getBsonString(doc bson.M, key string) string {
	if v, ok := doc[key].(string); ok {
		return v
	}
	return ""
}

// getBsonBool safely extracts bool from bson.M
func getBsonBool(doc bson.M, key string) bool {
	if v, ok := doc[key].(bool); ok {
		return v
	}
	return false
}

// formatUserFromBSON converts BSON document to GraphQL User model
func formatUserFromBSON(userDoc bson.M) *model.User {
	objID, ok := userDoc["_id"].(primitive.ObjectID)
	if !ok {
		return nil
	}
	userID := objID.Hex()
	hasAvatar := getBsonBool(userDoc, "hasAvatar")
	var avatarURL *string
	if hasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", userID)
		avatarURL = &url
	}
	username := getBsonString(userDoc, "username")
	if username == "" {
		username = getBsonString(userDoc, "email")
	}
	return &model.User{
		ID:        userID,
		Email:     getBsonString(userDoc, "email"),
		Username:  username,
		IsAdmin:   getBsonBool(userDoc, "isAdmin"),
		HasAvatar: hasAvatar,
		AvatarURL: avatarURL,
	}
}

// formatAuthor converts MongoDB user document to GraphQL Author model
func formatAuthor(userDoc bson.M, authorID string) *model.Author {
	authorUsername := getBsonString(userDoc, "username")
	if authorUsername == "" {
		authorUsername = getBsonString(userDoc, "email")
	}
	authorHasAvatar := getBsonBool(userDoc, "hasAvatar")
	var authorAvatarURL *string
	if authorHasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", authorID)
		authorAvatarURL = &url
	}
	return &model.Author{
		ID:        authorID,
		Username:  authorUsername,
		HasAvatar: authorHasAvatar,
		AvatarURL: authorAvatarURL,
	}
}

// formatReport converts MongoDB report document to GraphQL Report model
func (r *Resolver) formatReport(ctx context.Context, reportDoc bson.M, currentUser *auth.CurrentUser) (*model.Report, error) {
	reportOID, ok := reportDoc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid report _id")
	}
	reportID := reportOID.Hex()
	authorOID, ok := reportDoc["authorId"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid report authorId")
	}
	authorID := authorOID.Hex()

	// Get author
	db, err := mongodb.GetDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить доступ к базе данных: %w", err)
	}

	var authorDoc bson.M
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": reportDoc["authorId"]}).Decode(&authorDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to find author: %w", err)
	}

	var author *model.Author
	if err == nil {
		author = formatAuthor(authorDoc, authorID)
	} else {
		author = &model.Author{
			ID:        authorID,
			Username:  "unknown",
			HasAvatar: false,
		}
	}

	// Format photos
	photos := []*model.Photo{}
	if photosArray, ok := reportDoc["photos"].(bson.A); ok {
		for i := range photosArray {
			photos = append(photos, &model.Photo{
				URL: fmt.Sprintf("/api/reports/%s/photos/%d", reportID, i),
			})
		}
	}

	// Format dates
	var createdAt *scalars.Time
	if ct, ok := reportDoc["createdAt"].(primitive.DateTime); ok {
		t := scalars.Time(time.Unix(int64(ct)/1000, 0))
		createdAt = &t
	}

	var updatedAt *scalars.Time
	if ut, ok := reportDoc["updatedAt"].(primitive.DateTime); ok {
		t := scalars.Time(time.Unix(int64(ut)/1000, 0))
		updatedAt = &t
	}

	canEdit := isAllowed(currentUser, authorID)

	return &model.Report{
		ID:        reportID,
		Title:     getBsonString(reportDoc, "title"),
		Text:      getBsonString(reportDoc, "text"),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		AuthorID:  authorID,
		Author:    author,
		Photos:    photos,
		CanEdit:   canEdit,
	}, nil
}

// formatCompetition converts MongoDB competition document to GraphQL Competition model
func formatCompetition(competitionDoc bson.M) (*model.Competition, error) {
	compOID, ok := competitionDoc["_id"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid competition _id")
	}
	competitionID := compOID.Hex()

	// Format tours
	tours := []*model.Tour{}
	if toursArray, ok := competitionDoc["tours"].(bson.A); ok {
		for _, tourRaw := range toursArray {
			tourDoc, ok := tourRaw.(bson.M)
			if !ok {
				continue
			}
			tourDateVal, okDate := tourDoc["date"].(primitive.DateTime)
			tourTimeVal, okTime := tourDoc["time"].(string)
			if !okDate || !okTime {
				continue
			}
			tours = append(tours, &model.Tour{
				Date: scalars.Time(time.Unix(int64(tourDateVal)/1000, 0)),
				Time: tourTimeVal,
			})
		}
	}

	// Format dates
	sd, ok := competitionDoc["startDate"].(primitive.DateTime)
	if !ok {
		return nil, fmt.Errorf("invalid competition startDate")
	}
	ed, ok := competitionDoc["endDate"].(primitive.DateTime)
	if !ok {
		return nil, fmt.Errorf("invalid competition endDate")
	}
	startDate := scalars.Time(time.Unix(int64(sd)/1000, 0))
	endDate := scalars.Time(time.Unix(int64(ed)/1000, 0))

	var openingDate *scalars.Time
	if od, ok := competitionDoc["openingDate"].(primitive.DateTime); ok {
		t := scalars.Time(time.Unix(int64(od)/1000, 0))
		openingDate = &t
	}

	var openingTime *string
	if ot, ok := competitionDoc["openingTime"].(string); ok && ot != "" {
		openingTime = &ot
	}

	var fee *float64
	if f, ok := competitionDoc["fee"].(float64); ok {
		fee = &f
	}

	var teamLimit *int
	if tl, ok := competitionDoc["teamLimit"].(int32); ok {
		t := int(tl)
		teamLimit = &t
	}

	var regulations *string
	if reg, ok := competitionDoc["regulations"].(string); ok && reg != "" {
		regulations = &reg
	}

	var createdAt *scalars.Time
	if ct, ok := competitionDoc["createdAt"].(primitive.DateTime); ok {
		t := scalars.Time(time.Unix(int64(ct)/1000, 0))
		createdAt = &t
	}

	var updatedAt *scalars.Time
	if ut, ok := competitionDoc["updatedAt"].(primitive.DateTime); ok {
		t := scalars.Time(time.Unix(int64(ut)/1000, 0))
		updatedAt = &t
	}

	return &model.Competition{
		ID:               competitionID,
		Title:            getBsonString(competitionDoc, "title"),
		StartDate:        startDate,
		EndDate:          endDate,
		Location:         getBsonString(competitionDoc, "location"),
		Tours:            tours,
		OpeningDate:      openingDate,
		OpeningTime:      openingTime,
		IndividualFormat: getBsonBool(competitionDoc, "individualFormat"),
		TeamFormat:       getBsonBool(competitionDoc, "teamFormat"),
		Fee:              fee,
		TeamLimit:        teamLimit,
		Regulations:      regulations,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// formatUserFromEntity converts domain entity User to GraphQL User model
func formatUserFromEntity(user *entity.User) *model.User {
	hasAvatar := user.HasAvatar
	var avatarURL *string
	if hasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", user.ID)
		avatarURL = &url
	}
	username := user.Username
	if username == "" {
		username = user.Email
	}
	return &model.User{
		ID:        user.ID,
		Email:     user.Email,
		Username:  username,
		IsAdmin:   user.IsAdmin,
		HasAvatar: hasAvatar,
		AvatarURL: avatarURL,
	}
}

// formatReportFromEntity converts domain entity Report to GraphQL Report model
func (r *Resolver) formatReportFromEntity(ctx context.Context, report *entity.Report, currentUser *auth.CurrentUser) (*model.Report, error) {
	// Get author
	db, err := mongodb.GetDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить доступ к базе данных: %w", err)
	}

	authorIDObj, err := primitive.ObjectIDFromHex(report.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author ID: %w", err)
	}

	var authorDoc bson.M
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": authorIDObj}).Decode(&authorDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to find author: %w", err)
	}

	var author *model.Author
	if err == nil {
		author = formatAuthor(authorDoc, report.AuthorID)
	} else {
		author = &model.Author{
			ID:        report.AuthorID,
			Username:  "unknown",
			HasAvatar: false,
		}
	}

	// Format photos
	photos := []*model.Photo{}
	for i := range report.Photos {
		photos = append(photos, &model.Photo{
			URL: fmt.Sprintf("/api/reports/%s/photos/%d", report.ID, i),
		})
	}

	// Format dates
	var createdAt *scalars.Time
	if !report.CreatedAt.IsZero() {
		t := scalars.Time(report.CreatedAt)
		createdAt = &t
	}

	var updatedAt *scalars.Time
	if !report.UpdatedAt.IsZero() {
		t := scalars.Time(report.UpdatedAt)
		updatedAt = &t
	}

	canEdit := isAllowed(currentUser, report.AuthorID)

	return &model.Report{
		ID:        report.ID,
		Title:     report.Title,
		Text:      report.Text,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		AuthorID:  report.AuthorID,
		Author:    author,
		Photos:    photos,
		CanEdit:   canEdit,
	}, nil
}

// formatCompetitionFromEntity converts domain entity Competition to GraphQL Competition model
// Since entity.Competition doesn't have all fields, we fetch from MongoDB directly
func formatCompetitionFromEntity(competition *entity.Competition) (*model.Competition, error) {
	// Fetch full competition document from MongoDB since entity doesn't have all fields
	db, err := mongodb.GetDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить доступ к базе данных: %w", err)
	}

	competitionID, err := primitive.ObjectIDFromHex(competition.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid competition ID: %w", err)
	}

	var competitionDoc bson.M
	err = db.Collection("competitions").FindOne(context.Background(), bson.M{"_id": competitionID}).Decode(&competitionDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to find competition: %w", err)
	}

	// Use existing formatCompetition function which works with bson.M
	return formatCompetition(competitionDoc)
}
