package main

import (
	"context"
	"log"
	"os"
	"time"

	"smart-forms/internal/analytics"
	"smart-forms/internal/auth"
	"smart-forms/internal/cache"
	"smart-forms/internal/flows"
	"smart-forms/internal/forms"
	"smart-forms/internal/links"
	"smart-forms/internal/migrations"
	"smart-forms/internal/plans"
	"smart-forms/internal/questions"
	"smart-forms/internal/responses"
	"smart-forms/internal/responses/buffer"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var db *pgxpool.Pool

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Run database migrations
	if err := migrations.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
		// Don't fatal here - let the app try to start anyway
	}

	// Connect to DB (POOL)
	db = connectDB()
	defer db.Close()

	// Initialize cache
	formCache, err := cache.NewCache(cache.Config{
		MaxCost:     100 * 1024 * 1024, // 100MB
		NumCounters: 10_000_000,        // 10M counters
		BufferItems: 64,                // Ring buffer size
		DefaultTTL:  5 * time.Minute,   // 5 minute TTL
	})
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}
	defer formCache.Close()
	log.Println("Cache initialized successfully (100MB limit, 5min TTL)")

	// Initialize response buffer for batch inserts
	responseBuffer := buffer.NewResponseBuffer(
		db,
		500,              // Queue size
		50,               // Batch size
		300*time.Millisecond, // Flush interval
	)
	defer responseBuffer.Close()

	app := fiber.New()

	// Logger middleware
	app.Use(logger.New())

	// CORS middleware (from ENV)
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("CORS_ORIGINS"),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, ngrok-skip-browser-warning",
	}))

	// Custom ENV middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-App-Name", os.Getenv("APP_NAME"))
		return c.Next()
	})

	// Routes
	app.Get("/", helloHandler)
	app.Get("/status", statusHandler)

	// Auth setup
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	// Auth routes
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/refresh", authHandler.Refresh)
	app.Post("/auth/register", authHandler.Register)

	formsRepo := forms.NewFormsRepository(db)
	formsService := forms.NewFormsService(formsRepo, formCache)
	formsHandler := forms.NewFormsHandler(formsService)

	questionRepo := questions.NewQuestionRepository(db)
	questionService := questions.NewQuestionService(questionRepo)
	questionHandler := questions.NewQuestionHandler(questionService)

	flowRepo := flows.NewFlowRepository(db)
	flowService := flows.NewFlowService(flowRepo, formCache)
	flowHandler := flows.NewFlowHandler(flowService)

	linksRepo := links.NewLinksRepository(db)
	linksService := links.NewLinksService(linksRepo, formCache)
	linksHandler := links.NewLinksHandler(linksService)

	responsesRepo := responses.NewResponsesRepository(db)
	responsesService := responses.NewResponsesService(responsesRepo, responseBuffer)
	responsesHandler := responses.NewResponsesHandler(responsesService)

	analyticsRepo := analytics.NewAnalyticsRepository(db)
	analyticsService := analytics.NewAnalyticsService(analyticsRepo)
	analyticsHandler := analytics.NewAnalyticsHandler(analyticsService)

	plansRepo := plans.NewPlansRepository(db)
	plansService := plans.NewPlansService(plansRepo)
	plansHandler := plans.NewPlansHandler(plansService)

	// Public routes (no auth) - MUST be before protected group
	app.Get("/f/:slug", linksHandler.GetPublicForm)
	app.Post("/f/:slug/responses", responsesHandler.SubmitResponse)
	app.Get("/plans", plansHandler.ListActivePlans) // Public pricing page

	// Protect routes
	api := app.Group("/", auth.JWTAuthMiddleware())

	// Forms routes
	api.Post("/forms", formsHandler.Create)
	api.Get("/forms", formsHandler.List)
	api.Get("/forms/:id", formsHandler.GetByID)
	api.Patch("/forms/:id", formsHandler.Update)
	api.Patch("/forms/:id/delete", formsHandler.SoftDelete)

	// Questions routes
	api.Post("/questions", questionHandler.Create)
	api.Get("/questions", questionHandler.List)
	api.Get("/questions/:id", questionHandler.GetByID)
	api.Patch("/questions/:id", questionHandler.Update)
	api.Delete("/questions/:id", questionHandler.Delete)

	// Flow routes
	api.Patch("/forms/:form_id/flow", flowHandler.UpdateFlow)
	api.Get("/forms/:form_id/flow", flowHandler.GetFlow)

	// Links routes (protected)
	api.Patch("/forms/:form_id/publish", linksHandler.PublishForm)
	api.Patch("/forms/:form_id/accepting-responses", linksHandler.ToggleAcceptingResponses)

	// Responses routes (protected)
	api.Get("/forms/:form_id/responses", responsesHandler.GetFormResponses)
	api.Get("/responses/:response_id", responsesHandler.GetResponseDetails)

	// Analytics routes (protected)
	api.Get("/forms/:form_id/analytics/status", analyticsHandler.GetAnalyticsStatus)
	api.Get("/forms/:form_id/analytics/nodes", analyticsHandler.GetNodeAnalytics)
	api.Get("/forms/:form_id/analytics/flow", analyticsHandler.GetFlowAnalytics)

	// Super Admin routes (requires super_admin role)
	admin := api.Group("/admin", auth.RequireSuperAdmin())

	// Plans management (super admin only)
	admin.Get("/plans", plansHandler.ListAllPlans)
	admin.Post("/plans", plansHandler.CreatePlan)
	admin.Get("/plans/:id", plansHandler.GetPlan)
	admin.Patch("/plans/:id", plansHandler.UpdatePlan)
	admin.Delete("/plans/:id", plansHandler.DeletePlan)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Server running on port", port)
	log.Fatal(app.Listen(":" + port))
}

// ---------------- HANDLERS ----------------

func helloHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hello World 🚀",
		"app":     os.Getenv("APP_NAME"),
	})
}

func statusHandler(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := db.Ping(ctx)
	dbStatus := "connected"
	if err != nil {
		dbStatus = "disconnected"
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"db":     dbStatus,
	})
}

// ---------------- DB ----------------

func connectDB() *pgxpool.Pool {
	connStr := os.Getenv("DATABASE_URL")

	// Parse connection string and configure pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal("Failed to parse database URL:", err)
	}

	// Connection pool configuration optimized for 916MB RAM server
	config.MaxConns = 15                           // Maximum connections (balance between throughput and memory)
	config.MinConns = 3                            // Minimum idle connections (keep connections warm)
	config.MaxConnLifetime = 1 * time.Hour         // Recycle connections after 1 hour
	config.MaxConnIdleTime = 15 * time.Minute      // Close idle connections after 15 minutes
	config.HealthCheckPeriod = 1 * time.Minute     // Health check every minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second // Connection timeout

	// Create pool with configuration
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Verify connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Printf("Database pool created successfully (max: %d, min: %d)",
		config.MaxConns, config.MinConns)
	return pool
}

// TODO: OPTIMIZE - Add Cache Metrics Endpoint
// Add GET /metrics/cache endpoint to monitor cache performance
// Should return:
// - Hit rate (hits / (hits + misses))
// - Total hits, misses
// - Current cache size
// - Memory usage
// - Most accessed keys
// Implementation: Use formCache.GetMetrics() from Ristretto
// Example: app.Get("/metrics/cache", cacheMetricsHandler)
