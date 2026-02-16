package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	clientOnce sync.Once
)

// GetMongoURI returns MongoDB connection URI from environment
func GetMongoURI() string {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		panic("Missing MONGODB_URI in environment variables")
	}
	return uri
}

// GetClient returns MongoDB client (singleton)
func GetClient() (*mongo.Client, error) {
	var err error
	clientOnce.Do(func() {
		uri := GetMongoURI()
		ctx := context.Background()
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			return
		}
		// Test connection
		err = client.Ping(ctx, nil)
	})
	return client, err
}

// GetDBName extracts database name from MongoDB URI
func GetDBName() string {
	uri := GetMongoURI()
	
	// Try to parse as URL first
	if parsed, err := url.Parse(uri); err == nil {
		// Remove leading slash from path
		dbName := strings.TrimPrefix(parsed.Path, "/")
		// Remove query parameters if any
		if idx := strings.Index(dbName, "?"); idx != -1 {
			dbName = dbName[:idx]
		}
		if dbName != "" {
			return dbName
		}
	}
	
	// Fallback: try to extract from URI string directly
	// Format: mongodb://host:port/dbname
	parts := strings.Split(uri, "/")
	if len(parts) > 1 {
		dbPart := parts[len(parts)-1]
		// Remove query parameters
		if idx := strings.Index(dbPart, "?"); idx != -1 {
			dbPart = dbPart[:idx]
		}
		if dbPart != "" {
			return dbPart
		}
	}
	
	// Default database name
	return "cnpf_feeder"
}

// GetDB returns MongoDB database instance
func GetDB() (*mongo.Database, error) {
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к MongoDB: %w", err)
	}
	dbName := GetDBName()
	if dbName == "" {
		return nil, fmt.Errorf("имя базы данных не может быть пустым. Проверьте MONGODB_URI")
	}
	return client.Database(dbName), nil
}
