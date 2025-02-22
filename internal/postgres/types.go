package postgres

import "time"

// FundType represents the type of a fund (e.g., "Equity", "Bond", etc.)
type FundType string

const (
	FundTypeEquity FundType = "Equity"
	FundTypeBond   FundType = "Bond"
	FundTypeIndex  FundType = "Index"
	FundTypeMixed  FundType = "Mixed"
)

// RiskLevel represents the risk level of a fund (e.g., "Low", "Medium", "High")
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "Low"
	RiskLevelMedium RiskLevel = "Medium"
	RiskLevelHigh   RiskLevel = "High"
)

type ISA struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	FundIDs          []string  `json:"fund_ids" db:"fund_ids"`
	CashBalance      float64   `json:"cash_balance" db:"cash_balance"`
	InvestmentAmount float64   `json:"investment_amount" db:"investment_amount"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Fund struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        FundType  `json:"type" db:"type"`
	RiskLevel   RiskLevel `json:"risk_level" db:"risk_level"`
	Performance float64   `json:"performance" db:"performance"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Investment struct {
	ID         string    `json:"id" db:"id"`
	ISAID      string    `json:"isa_id" db:"isa_id"`
	FundID     string    `json:"fund_id" db:"fund_id"`
	Amount     float64   `json:"amount" db:"amount"`
	InvestedAt time.Time `json:"invested_at" db:"invested_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type User struct {
	ID        string    `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
