package postgres

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

func SetupTestDB() (*pgx.Conn, func(), error) {
	dbURL := os.Getenv("DB_URL")
	fmt.Println("My test url", dbURL)

	if dbURL == "" {
		return nil, nil, fmt.Errorf("DB_URL is not set")
	}

	ctx := context.Background()

	// Set up the connection
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, nil, err
	}

	// Create cleanup function to disconnect and remove test data
	cleanup := func() {
		//Delete all the test data inserted into the DB
		_, err := conn.Exec(context.Background(), "DELETE FROM isas")
		if err != nil {
			log.Fatalf("Failed to cleanup isas table: %v", err)
		}

		_, err = conn.Exec(context.Background(), "DELETE FROM funds")
		if err != nil {
			log.Fatalf("Failed to cleanup isas table: %v", err)
		}

		_, err = conn.Exec(context.Background(), "DELETE FROM investments")
		if err != nil {
			log.Fatalf("Failed to cleanup investments table: %v", err)
		}
		conn.Close(ctx)
	}

	return conn, cleanup, nil
}
