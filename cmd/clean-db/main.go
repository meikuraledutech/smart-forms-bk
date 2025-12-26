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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	fmt.Println("⚠️  Cleaning database...")

	// Drop all tables in correct order (reverse of creation)
	dropSQL := `
		DROP TABLE IF EXISTS questions CASCADE;
		DROP TABLE IF EXISTS forms CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`

	_, err = conn.Exec(ctx, dropSQL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Database cleaned")
}
