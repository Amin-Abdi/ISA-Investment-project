package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Amin-Abdi/ISA-Investment-project/api/server"
	"github.com/Amin-Abdi/ISA-Investment-project/api/server/mocks"
	"github.com/Amin-Abdi/ISA-Investment-project/internal/postgres"
)

//go:generate moq -out ./mocks/store.mock.go -skip-ensure -pkg mocks . Store

func setupTestServer(store *mocks.StoreMock) *gin.Engine {
	s := &server.Server{Store: store}
	r := gin.Default()
	r.POST("/isa/:id/invest", s.InvestIntoFund)

	return r
}

func TestInvestInFund(t *testing.T) {
	now := time.Now()
	tests := map[string]struct {
		reqBody interface{}

		isaID       string
		getIsa      postgres.ISA
		getIsaError error

		fundID       string
		getFund      postgres.Fund
		getFundError error

		moneyToInvest float64
		investmentID  string

		errorReturned    bool
		expectedStatus   int
		expectedResponse interface{}
	}{
		"failure: Missing fund_id and amount": {
			isaID:            "non-existent-id",
			errorReturned:    true,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Invalid request. Fund ID and amount are required.",
		},

		"failure: isa not found": {
			isaID: "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
			reqBody: map[string]interface{}{
				"fund_id": "fund-123",
				"amount":  10000.0,
			},
			errorReturned:    true,
			getIsaError:      postgres.ErrNotFound,
			expectedStatus:   http.StatusNotFound,
			expectedResponse: "Isa not found. Please check the id and try again.",
		},

		"failure: amount to invest is greater than the balance": {
			isaID: "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
			reqBody: map[string]interface{}{
				"fund_id": "fund-123",
				"amount":  10000.0,
			},
			getIsa: postgres.ISA{
				ID:               "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
				UserID:           "123e4567-e89b-12d3-a456-426614174000",
				FundIDs:          []string{"373e51ae-f6b9-4a29-a219-5816aa3d68e0"},
				CashBalance:      500.00,
				InvestmentAmount: 0.00,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			errorReturned:    true,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Insufficient balance for this investment. Please add funds to your account and try again",
		},

		"failure: fund is not related to the isa": {
			isaID: "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
			reqBody: map[string]interface{}{
				"fund_id": "fund-123",
				"amount":  1000.0,
			},
			getIsa: postgres.ISA{
				ID:               "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
				UserID:           "123e4567-e89b-12d3-a456-426614174000",
				FundIDs:          []string{"373e51ae-f6b9-4a29-a219-5816aa3d68e0"},
				CashBalance:      10000.00,
				InvestmentAmount: 0.00,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			errorReturned:    true,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Fund not found in your ISA. Please add it before investing.",
		},

		"failure: fund not found": {
			isaID: "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
			reqBody: map[string]interface{}{
				"fund_id": "373e51ae-f6b9-4a29-a219-5816aa3d68e0",
				"amount":  1000.0,
			},
			getIsa: postgres.ISA{
				ID:               "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
				UserID:           "123e4567-e89b-12d3-a456-426614174000",
				FundIDs:          []string{"373e51ae-f6b9-4a29-a219-5816aa3d68e0"},
				CashBalance:      10000.00,
				InvestmentAmount: 500.00,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			moneyToInvest:    1000,
			fundID:           "373e51ae-f6b9-4a29-a219-5816aa3d68e0",
			getFundError:     postgres.ErrNotFound,
			errorReturned:    true,
			expectedStatus:   http.StatusNotFound,
			expectedResponse: "Fund not found. Please check the id and try again.",
		},

		"success: invest 25,000 into a fund": {
			isaID: "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
			reqBody: map[string]interface{}{
				"fund_id": "373e51ae-f6b9-4a29-a219-5816aa3d68e0",
				"amount":  25000.0,
			},
			getIsa: postgres.ISA{
				ID:               "62ad0fef-9bdc-43a1-85ca-05b60f39cf8f",
				UserID:           "123e4567-e89b-12d3-a456-426614174000",
				FundIDs:          []string{"373e51ae-f6b9-4a29-a219-5816aa3d68e0"},
				CashBalance:      25000.00,
				InvestmentAmount: 0.00,
				CreatedAt:        now,
				UpdatedAt:        now,
			},

			fundID: "373e51ae-f6b9-4a29-a219-5816aa3d68e0",
			getFund: postgres.Fund{
				ID:          "373e51ae-f6b9-4a29-a219-5816aa3d68e0",
				Name:        "Growth Fund",
				Description: "A high-performance equity fund",
				Type:        postgres.FundTypeEquity,
				RiskLevel:   postgres.RiskLevelHigh,
				Performance: 12.5,
				TotalAmount: 10000,
				CreatedAt:   now,
				UpdatedAt:   now,
			},

			moneyToInvest: 25000,
			investmentID:  "bde2702d-b189-4a57-8a0f-1abdad9f50fe",

			expectedStatus: http.StatusOK,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockStore := &mocks.StoreMock{
				GetIsaFunc: func(ctx context.Context, id string) (*postgres.ISA, error) {
					assert.Equal(t, test.isaID, id)
					if test.getIsaError != nil {
						return nil, test.getIsaError
					}
					return &test.getIsa, nil
				},
				GetFundFunc: func(ctx context.Context, id string) (*postgres.Fund, error) {
					assert.Equal(t, test.fundID, id)
					if test.getFundError != nil {
						return nil, test.getFundError
					}
					return &test.getFund, nil
				},
				UpdateIsaFunc: func(ctx context.Context, isaID string, cashBalance, investmentAmount float64) (*postgres.ISA, error) {
					newCashBalance := test.getIsa.CashBalance - test.moneyToInvest
					newInvestmentAmount := test.getIsa.InvestmentAmount + test.moneyToInvest
					assert.Equal(t, test.isaID, isaID)
					assert.Equal(t, newCashBalance, cashBalance)
					assert.Equal(t, newInvestmentAmount, investmentAmount)
					return nil, nil
				},
				UpdateFundTotalAmountFunc: func(ctx context.Context, fundID string, totalAmount float64) (*postgres.Fund, error) {
					newTotalAmount := test.getFund.TotalAmount + test.moneyToInvest
					assert.Equal(t, test.fundID, fundID)
					assert.Equal(t, newTotalAmount, totalAmount)
					return nil, nil
				},
				CreateInvestmentFunc: func(ctx context.Context, investment postgres.Investment) (string, error) {
					assert.Equal(t, test.fundID, investment.FundID)
					assert.Equal(t, test.isaID, investment.ISAID)
					assert.Equal(t, test.moneyToInvest, investment.Amount)
					return test.investmentID, nil
				},
			}

			r := setupTestServer(mockStore)

			jsonBody, err := json.Marshal(test.reqBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/isa/"+test.isaID+"/invest", bytes.NewReader(jsonBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if test.errorReturned {
				assert.Equal(t, test.expectedResponse, response["error"])
			}
		})
	}
}
