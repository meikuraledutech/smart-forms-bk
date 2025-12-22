package main

import (
	"context"
	"log"
	"os"
	"smart-forms/internal/auth"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var db *pgx.Conn

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to DB
	db = connectDB()

	app := fiber.New()

	// Logger middleware
	app.Use(logger.New())

	// CORS middleware (from ENV)
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("CORS_ORIGINS"),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Custom ENV middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-App-Name", os.Getenv("APP_NAME"))
		return c.Next()
	})

	// Routes
	app.Get("/", helloHandler)
	app.Get("/status", statusHandler)

	// auth setup
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	// routes
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/refresh", authHandler.Refresh)
	app.Post("/auth/register", authHandler.Register)



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

func connectDB() *pgx.Conn {
	connStr := os.Getenv("DATABASE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Println("Database connected successfully")
	return conn
}
