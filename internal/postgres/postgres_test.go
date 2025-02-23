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

func TestCreateFund(t *testing.T) {
	ctx := context.Background()

	// Set up the test database
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	tests := map[string]struct {
		initialFund  postgres.Fund
		expectedFund postgres.Fund
	}{
		"success: Create a Fund": {
			initialFund: postgres.Fund{
				ID:          "4b24808e-4114-4076-ac8d-031532ef8576",
				Name:        "Fund One",
				Description: "A sample fund",
				Type:        postgres.FundTypeEquity,
				RiskLevel:   postgres.RiskLevelHigh,
				Performance: 12.5,
				TotalAmount: 1000000,
			},
			expectedFund: postgres.Fund{
				ID:          "4b24808e-4114-4076-ac8d-031532ef8576",
				Name:        "Fund One",
				Description: "A sample fund",
				Type:        postgres.FundTypeEquity,
				RiskLevel:   postgres.RiskLevelHigh,
				Performance: 12.5,
				TotalAmount: 1000000,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Create the fund
			fundID, err := store.CreateFund(ctx, test.initialFund)
			require.NoError(t, err)

			// Get the fund back
			createdFund, err := store.GetFund(ctx, fundID)
			require.NoError(t, err)

			assert.Equal(t, test.expectedFund.ID, createdFund.ID)
			assert.Equal(t, test.expectedFund.Name, createdFund.Name)
			assert.Equal(t, test.expectedFund.Description, createdFund.Description)
			assert.Equal(t, test.expectedFund.Type, createdFund.Type)
			assert.Equal(t, test.expectedFund.RiskLevel, createdFund.RiskLevel)
			assert.Equal(t, test.expectedFund.Performance, createdFund.Performance)
			assert.Equal(t, test.expectedFund.TotalAmount, createdFund.TotalAmount)
		})
	}

}

func TestGetFundNotFound(t *testing.T) {
	ctx := context.Background()

	// Set up the test database
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	// Test for a Fund ID that doesn't exist
	invalidID := "non-existent-id"
	fund, err := store.GetFund(ctx, invalidID)

	require.Error(t, err)
	assert.Nil(t, fund)
	fmt.Println("HERE:", err.Error())
	assert.Contains(t, err.Error(), "get fund: ")
}

func TestUpdateFund(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	tests := map[string]struct {
		initialFund   postgres.Fund
		updateFund    postgres.Fund
		expectedFund  postgres.Fund
		errorContains string
	}{
		"success: Update an existing Fund": {
			initialFund: postgres.Fund{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Name:        "Tech Growth Fund",
				Description: "A fund focused on technology stocks",
				Type:        postgres.FundTypeEquity,
				RiskLevel:   postgres.RiskLevelHigh,
				Performance: 15.4,
				TotalAmount: 1000000,
			},
			updateFund: postgres.Fund{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Name:        "Tech Growth Fund Plus",
				Description: "An updated description for the tech fund",
			},
			expectedFund: postgres.Fund{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Name:        "Tech Growth Fund Plus",
				Description: "An updated description for the tech fund",
				Type:        postgres.FundTypeEquity,
				RiskLevel:   postgres.RiskLevelHigh,
				Performance: 15.4,
				TotalAmount: 1000000,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := store.CreateFund(ctx, test.initialFund)
			require.NoError(t, err)

			updatedFund, err := store.UpdateFund(ctx, test.updateFund.ID, test.updateFund.Name, test.updateFund.Description)
			require.NoError(t, err)

			assert.Equal(t, test.expectedFund.ID, updatedFund.ID)
			assert.Equal(t, test.expectedFund.Name, updatedFund.Name)
			assert.Equal(t, test.expectedFund.Description, updatedFund.Description)
			assert.Equal(t, test.expectedFund.TotalAmount, updatedFund.TotalAmount)
			assert.Equal(t, test.expectedFund.Type, updatedFund.Type)
			assert.Equal(t, test.expectedFund.RiskLevel, updatedFund.RiskLevel)
			assert.Equal(t, test.expectedFund.Performance, updatedFund.Performance)
			assert.WithinDuration(t, now, updatedFund.UpdatedAt, time.Millisecond*100)
		})
	}

}

func TestUpdateFundTotalAmount(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	initialFund := postgres.Fund{
		ID:          "a8364471-0a6c-4537-a7e3-dc2a18d9f4b6",
		Name:        "Growth Fund",
		Description: "A high-risk, high-reward fund",
		Type:        postgres.FundTypeEquity,
		RiskLevel:   postgres.RiskLevelHigh,
		Performance: 12.5,
		TotalAmount: 50000,
	}

	_, err = store.CreateFund(ctx, initialFund)
	require.NoError(t, err)

	// Update total amount
	newTotalAmount := 75000.0
	updatedFund, err := store.UpdateFundTotalAmount(ctx, initialFund.ID, newTotalAmount)
	require.NoError(t, err)

	// Ensure only TotalAmount changed
	assert.Equal(t, initialFund.ID, updatedFund.ID)
	assert.Equal(t, initialFund.Name, updatedFund.Name)
	assert.Equal(t, initialFund.Description, updatedFund.Description)
	assert.Equal(t, initialFund.Type, updatedFund.Type)
	assert.Equal(t, initialFund.RiskLevel, updatedFund.RiskLevel)
	assert.Equal(t, initialFund.Performance, updatedFund.Performance)
	assert.Equal(t, newTotalAmount, updatedFund.TotalAmount) // Ensure total amount is updated
	assert.WithinDuration(t, now, updatedFund.UpdatedAt, time.Millisecond*100)
}

func TestListFunds(t *testing.T) {
	ctx := context.Background()
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	fund1 := postgres.Fund{
		ID:          "4b24808e-4114-4076-ac8d-031532ef8576",
		Name:        "Fund One",
		Description: "A sample fund",
		Type:        postgres.FundTypeEquity,
		RiskLevel:   postgres.RiskLevelHigh,
		Performance: 12.5,
		TotalAmount: 1000000,
	}
	fund2 := postgres.Fund{
		ID:          "7c9b02c8-2924-48b4-9223-2e6471bc1939",
		Name:        "Fund Two",
		Description: "Another sample fund",
		Type:        postgres.FundTypeBond,
		RiskLevel:   postgres.RiskLevelLow,
		Performance: 8.7,
		TotalAmount: 500000,
	}

	_, err = store.CreateFund(ctx, fund1)
	require.NoError(t, err)
	_, err = store.CreateFund(ctx, fund2)
	require.NoError(t, err)

	gotFunds, err := store.ListFunds(ctx)
	require.NoError(t, err)

	assert.Len(t, gotFunds, 2)
	expectedFunds := []postgres.Fund{fund1, fund2}

	for i := range gotFunds {
		assert.Equal(t, expectedFunds[i].ID, gotFunds[i].ID)
		assert.Equal(t, expectedFunds[i].Name, gotFunds[i].Name)
		assert.Equal(t, expectedFunds[i].Description, gotFunds[i].Description)
		assert.Equal(t, expectedFunds[i].Type, gotFunds[i].Type)
		assert.Equal(t, expectedFunds[i].RiskLevel, gotFunds[i].RiskLevel)
		assert.Equal(t, expectedFunds[i].Performance, gotFunds[i].Performance)
		assert.Equal(t, expectedFunds[i].TotalAmount, gotFunds[i].TotalAmount)
	}
}

func TestCreateInvestment(t *testing.T) {
	ctx := context.Background()
	conn, cleanup, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer cleanup()

	store := postgres.NewStore(conn)

	isa := postgres.ISA{
		ID:               "ccba7538-a706-4816-b85a-2424f64df11a",
		UserID:           "6343b120-b611-4288-a8ff-9c79dec043f1",
		FundIDs:          []string{},
		CashBalance:      50000,
		InvestmentAmount: 0,
	}
	_, err = store.CreateIsa(ctx, isa)
	require.NoError(t, err)

	fund := postgres.Fund{
		ID:          "4b24808e-4114-4076-ac8d-031532ef8576",
		Name:        "Fund One",
		Description: "A sample fund",
		Type:        postgres.FundTypeEquity,
		RiskLevel:   postgres.RiskLevelHigh,
		Performance: 12.5,
		TotalAmount: 1000000,
	}
	_, err = store.CreateFund(ctx, fund)
	require.NoError(t, err)

	tests := map[string]struct {
		initialInvestment postgres.Investment
		errorContains     string
	}{
		"success: Create an Investment": {
			initialInvestment: postgres.Investment{
				ID:     "ad3e30a4-761b-4e0d-a6f5-4fc8a1b4f299",
				ISAID:  isa.ID,
				FundID: fund.ID,
				Amount: 10000,
			},
		},
		"failure: Create Investment with non-existent ISA": {
			initialInvestment: postgres.Investment{
				ID:     "0f71a84e-8cd3-4a87-b4c4-7ef582aa5f31",
				ISAID:  "2ba4eb3d-68f6-475c-9164-a5717eab1acc", //non-existent ISA
				FundID: fund.ID,
				Amount: 5000,
			},
			errorContains: "foreign key constraint",
		},
		"failure: Create Investment with non-existent Fund": {
			initialInvestment: postgres.Investment{
				ID:     "1d41e3e2-f841-4f6b-ae2c-2bdf1ad0ecbd",
				ISAID:  isa.ID,
				FundID: "2ba4eb3d-68f6-475c-9164-a5717eab1acc", //non-existent Fund
				Amount: 5000,
			},
			errorContains: "foreign key constraint",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			investmentID, err := store.CreateInvestment(ctx, test.initialInvestment)
			if test.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
				return
			}

			require.NoError(t, err)

			createdInvestment, err := store.GetInvestment(ctx, investmentID)
			require.NoError(t, err)

			assert.Equal(t, test.initialInvestment.ID, createdInvestment.ID)
			assert.Equal(t, test.initialInvestment.ISAID, createdInvestment.ISAID)
			assert.Equal(t, test.initialInvestment.FundID, createdInvestment.FundID)
			assert.Equal(t, test.initialInvestment.Amount, createdInvestment.Amount)
			assert.WithinDuration(t, time.Now(), createdInvestment.InvestedAt, time.Millisecond*100)
			assert.WithinDuration(t, time.Now(), createdInvestment.CreatedAt, time.Millisecond*100)

		})
	}
}
