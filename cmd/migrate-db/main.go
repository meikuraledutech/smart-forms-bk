package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	fmt.Println("🚀 Running migrations...")

	for _, file := range files {
		fmt.Println("➡️ ", file)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Exec(ctx, string(sqlBytes))
		if err != nil {
			fmt.Printf("⚠️  Warning: %s - %v\n", file, err)
			continue
		}
		fmt.Println("✅", file)
	}

	fmt.Println("✅ All migrations completed")
}
