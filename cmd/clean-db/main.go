package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	sqlBytes, err := os.ReadFile("migrations/001_create_users.down.sql")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	fmt.Println("⚠️  Cleaning database...")
	_, err = conn.Exec(ctx, string(sqlBytes))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Database cleaned")
}
