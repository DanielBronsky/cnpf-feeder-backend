package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/cnpf/feeder-backend/graph/generated"
	"github.com/cnpf/feeder-backend/graph/resolver"
	"github.com/cnpf/feeder-backend/internal/auth"
	"github.com/cnpf/feeder-backend/internal/domain"
	"github.com/cnpf/feeder-backend/internal/repository/mongodb"
	"github.com/cnpf/feeder-backend/internal/usecase"
)

const defaultPort = "4000"

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := domain.LoadConfig()

	// Initialize JWT
	if err := auth.InitJWT(); err != nil {
		log.Fatalf("Failed to initialize JWT: %v", err)
	}

	// Initialize MongoDB
	db, err := mongodb.GetDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Initialize repositories (infrastructure layer)
	userRepo := mongodb.NewUserRepository(db)
	reportRepo := mongodb.NewReportRepository(db)
	competitionRepo := mongodb.NewCompetitionRepository(db)
	registrationRepo := mongodb.NewRegistrationRepository(db)

	// Initialize use case (application layer) - uses repository interfaces
	useCase := usecase.NewUseCase(userRepo, reportRepo, competitionRepo, registrationRepo)

	// Initialize resolver (presentation layer) - uses use case
	// TEMPORARY: Passing repositories for backward compatibility during migration
	// TODO: Remove repository parameters after migration to Onion Architecture
	resolver := resolver.NewResolver(useCase, userRepo, reportRepo, competitionRepo, registrationRepo, db)

	// Set Gin mode
	ginMode := cfg.GinMode
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// Setup router
	router := gin.Default()

	// CORS configuration
	corsOrigin := cfg.CORSOrigin
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// GraphQL playground
	// Available in both debug and release modes for easier development/testing
	router.GET("/", playgroundHandler())

	// GraphQL endpoint
	router.POST("/graphql", graphqlHandler(resolver))

	// Get port
	port := cfg.Port
	if port == "" {
		port = defaultPort
	}

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 GraphQL Server ready at http://localhost:%s/graphql", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL Playground", "/graphql")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func graphqlHandler(resolver *resolver.Resolver) gin.HandlerFunc {
	// Create GraphQL handler
	h := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: resolver,
		}),
	)

	return func(c *gin.Context) {
		// Pass Gin context to GraphQL handler
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "ginContext", c)
		c.Request = c.Request.WithContext(ctx)
		
		h.ServeHTTP(c.Writer, c.Request)
	}
}
