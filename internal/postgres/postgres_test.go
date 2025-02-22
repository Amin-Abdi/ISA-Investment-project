package postgres_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Amin-Abdi/ISA-Investment-project/internal/postgres"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB() (*pgx.Conn, func(), error) {
	os.Setenv("DB_URL", "postgres://myuser:mypassword@localhost:5432/my_database?sslmode=disable")
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
		_, err := conn.Exec(context.Background(), "DELETE FROM isas")
		if err != nil {
			log.Fatalf("Failed to cleanup isas table: %v", err)
		}
		conn.Close(ctx)
	}

	return conn, cleanup, nil
}

func TestCreateISA(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	//Set up the test database
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	tests := map[string]struct {
		initialISA    postgres.ISA
		isaID         string
		errorContains string
	}{
		"success: Create an ISA": {
			initialISA: postgres.ISA{
				ID:               "ccba7538-a706-4816-b85a-2424f64df11a",
				UserID:           "6343b120-b611-4288-a8ff-9c79dec043f1",
				FundIDs:          []string{"be5fef5a-4637-47d2-a804-6308f95552c4"},
				CashBalance:      10000,
				InvestmentAmount: 25000,
			},
			isaID: "ccba7538-a706-4816-b85a-2424f64df11a",
		},
		"failure: Create an ISA with an existing uuid": {
			initialISA: postgres.ISA{
				ID:               "ccba7538-a706-4816-b85a-2424f64df11a",
				UserID:           "6343b120-b611-4288-a8ff-9c79dec043f1",
				FundIDs:          []string{"be5fef5a-4637-47d2-a804-6308f95552c4"},
				CashBalance:      10000,
				InvestmentAmount: 25000,
			},
			errorContains: "duplicate key value violates unique constraint",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			//create the isa
			isaID, err := store.CreateIsa(ctx, test.initialISA)

			if test.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
				return
			} else {
				require.NoError(t, err)
			}

			time.Sleep(100 * time.Millisecond) // Add a small delay
			createdISA, err := store.GetIsa(ctx, isaID)
			require.NoError(t, err)

			assert.Equal(t, test.initialISA.ID, createdISA.ID)
			assert.Equal(t, test.initialISA.UserID, createdISA.UserID)
			assert.Equal(t, test.initialISA.CashBalance, createdISA.CashBalance)
			assert.Equal(t, test.initialISA.InvestmentAmount, createdISA.InvestmentAmount)
			for i, fundID := range createdISA.FundIDs {
				assert.Equal(t, fundID, test.initialISA.FundIDs[i])
			}
			assert.WithinDuration(t, now, createdISA.CreatedAt, time.Millisecond*100)
			assert.WithinDuration(t, now, createdISA.UpdatedAt, time.Millisecond*100)
		})
	}

}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	//Set up the test database
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	tests := map[string]struct {
		initialISA    postgres.ISA
		updateISA     postgres.ISA
		expectedISA   postgres.ISA
		errorContains string
	}{
		"success: Update an existing ISA": {
			initialISA: postgres.ISA{
				ID:               "d9e89726-46f7-4f36-99ff-c9f45fd58fb3",
				UserID:           "6343b120-b611-4288-a8ff-9c79dec043f1",
				FundIDs:          []string{"a8364471-0a6c-4537-a7e3-dc2a18d9f4b6"},
				CashBalance:      10000,
				InvestmentAmount: 25000,
			},
			updateISA: postgres.ISA{
				ID:               "d9e89726-46f7-4f36-99ff-c9f45fd58fb3",
				FundIDs:          []string{"fe865bc7-5f3e-4523-895d-67b705041e5b", "289a532f-d7c3-423a-b5cb-af4a3f26ae48"},
				CashBalance:      15000,
				InvestmentAmount: 35000,
			},
			expectedISA: postgres.ISA{
				ID:               "d9e89726-46f7-4f36-99ff-c9f45fd58fb3",
				UserID:           "6343b120-b611-4288-a8ff-9c79dec043f1", // UserID should remain the same
				FundIDs:          []string{"fe865bc7-5f3e-4523-895d-67b705041e5b", "289a532f-d7c3-423a-b5cb-af4a3f26ae48"},
				CashBalance:      15000,
				InvestmentAmount: 35000,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			_, err := store.CreateIsa(ctx, test.initialISA)
			require.NoError(t, err)

			updatedISA, err := store.UpdateIsa(ctx, test.updateISA)
			require.NoError(t, err)

			assert.Equal(t, test.expectedISA.ID, updatedISA.ID)
			assert.Equal(t, test.expectedISA.UserID, updatedISA.UserID)
			assert.Equal(t, test.expectedISA.CashBalance, updatedISA.CashBalance)
			assert.Equal(t, test.expectedISA.InvestmentAmount, updatedISA.InvestmentAmount)
			assert.WithinDuration(t, now, updatedISA.UpdatedAt, time.Millisecond*100)
			for i, fundID := range updatedISA.FundIDs {
				assert.Equal(t, test.expectedISA.FundIDs[i], fundID)
			}
		})
	}

}
