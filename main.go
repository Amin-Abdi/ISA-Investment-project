package main

import (
	"context"
	"log"
	"os"

	"github.com/Amin-Abdi/ISA-Investment-project/api/server"
	"github.com/Amin-Abdi/ISA-Investment-project/internal/postgres"
	"github.com/jackc/pgx/v4"
)

func main() {

	//Load db from the environment
	os.Setenv("DB_URL", "postgres://myuser:mypassword@localhost:5432/my_database?sslmode=disable")
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

	store := postgres.NewStore(conn)
	s := server.NewServer(store)

	if err := s.Start(); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}
