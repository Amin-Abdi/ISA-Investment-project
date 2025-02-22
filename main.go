package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

func main() {

	//Load db from the environment
	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}

	ctx := context.Background()
	//connect to db
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v\n", err)
	}
	defer conn.Close(ctx)

}
