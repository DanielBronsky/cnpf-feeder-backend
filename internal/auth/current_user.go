package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/cnpf/feeder-backend/internal/repository/mongodb"
)

// CurrentUser represents the current authenticated user
type CurrentUser struct {
	ID        string
	Email     string
	Username  string
	IsAdmin   bool
	HasAvatar bool
}

// GetCurrentUser extracts current user from request (cookie or Authorization header)
func GetCurrentUser(c *gin.Context) (*CurrentUser, error) {
	// Try to get token from cookie
	token, err := c.Cookie(AuthCookieName)
	if err != nil {
		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			return nil, nil // No token found
		}
	}

	if token == "" {
		return nil, nil
	}

	// Verify token
	claims, err := VerifyToken(token)
	if err != nil {
		return nil, nil // Invalid token
	}

	// Get user from database
	database, err := mongodb.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	userID, err := primitive.ObjectIDFromHex(claims.Sub)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user struct {
		ID        primitive.ObjectID `bson:"_id"`
		Email     string             `bson:"email"`
		Username  string             `bson:"username"`
		IsAdmin   bool               `bson:"isAdmin"`
		HasAvatar bool              `bson:"hasAvatar"`
	}

	err = database.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &CurrentUser{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		HasAvatar: user.HasAvatar,
	}, nil
}
