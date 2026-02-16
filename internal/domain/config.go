package domain

import (
	"os"
)

// Config represents application configuration
type Config struct {
	// Server
	Port         string
	GinMode      string
	CORSOrigin   string
	
	// Database
	MongoDBURI   string
	MongoDBName  string
	
	// Auth
	AuthSecret   string
	
	// Logging
	LogLevel     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "4000"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "http://localhost:3000"),
		MongoDBURI:  getEnv("MONGODB_URI", ""),
		MongoDBName: getEnv("MONGODB_NAME", ""),
		AuthSecret:  getEnv("AUTH_SECRET", ""),
		LogLevel:    getEnv("LOGLEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
