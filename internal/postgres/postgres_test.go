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

	dbURL := os.Getenv("DB_URL")

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
