package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
)

type Store struct {
	db *pgx.Conn
}

func NewStore(db *pgx.Conn) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateIsa(ctx context.Context, isa ISA) (string, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()

	logger = logger.WithFields(logrus.Fields{
		"user_id": isa.UserID,
	})

	query := `INSERT INTO isas (id, user_id, fund_ids, cash_balance, investment_amount, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	args := []any{
		isa.ID,
		isa.UserID,
		pq.Array(isa.FundIDs), // Convert Go slice to PostgreSQL array
		isa.CashBalance,
		isa.InvestmentAmount,
		now,
		now,
	}

	var isaID string
	err := s.db.QueryRow(ctx, query, args...).Scan(&isaID)
	if err != nil {
		logger.WithError(err).Error("Failed to execute create isa query")
		return "", fmt.Errorf("execute create isa query: %w", err)
	}

	logger.Info("ISA succesfully created")
	return isaID, nil
}

func (s *Store) GetIsa(ctx context.Context, id string) (*ISA, error) {
	logger := logrus.New().WithContext(ctx)

	logger = logger.WithField("isa_id", id)

	query := `SELECT id, user_id, fund_ids, cash_balance, investment_amount, created_at, updated_at 
		FROM isas WHERE id = $1`

	var isa ISA
	err := s.db.QueryRow(ctx, query, id).
		Scan(
			&isa.ID,
			&isa.UserID,
			&isa.FundIDs,
			&isa.CashBalance,
			&isa.InvestmentAmount,
			&isa.CreatedAt,
			&isa.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("ISA not found")
			return nil, fmt.Errorf("get isa: %w", err)
		}
		logger.WithError(err).Error("Failed to execute query for get isa")
		return nil, fmt.Errorf("failed to execute query for get isa: %w", err)
	}

	return &isa, nil
}

func (s *Store) UpdateIsa(ctx context.Context, isa ISA) (*ISA, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()
	logger = logger.WithFields(logrus.Fields{
		"isa_id":  isa.ID,
		"user_id": isa.UserID,
	})

	query := `UPDATE isas
	SET fund_ids = $1, cash_balance = $2, investment_amount = $3, updated_at = $4
	WHERE id = $5
	RETURNING id, user_id, fund_ids, cash_balance, investment_amount, created_at, updated_at`

	args := []any{
		pq.Array(isa.FundIDs),
		isa.CashBalance,
		isa.InvestmentAmount,
		now, // Update the timestamp
		isa.ID,
	}

	var updatedISA ISA
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&updatedISA.ID,
		&updatedISA.UserID,
		&updatedISA.FundIDs,
		&updatedISA.CashBalance,
		&updatedISA.InvestmentAmount,
		&updatedISA.CreatedAt,
		&updatedISA.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("ISA not found for update")
			return nil, fmt.Errorf("isa not found: %w", err)
		}
		logger.WithError(err).Error("Failed to execute update isa query")
		return nil, fmt.Errorf("failed to execute update isa query: %w", err)
	}

	logger.Info("ISA successfully updated")
	return &updatedISA, nil
}

func (s *Store) CreateFund(ctx context.Context, fund Fund) (string, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()

	logger = logger.WithFields(logrus.Fields{
		"fund_id": fund.ID,
	})

	query := `INSERT INTO funds (id, name, description, type, risk_level, performance, total_amount, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	args := []any{
		fund.ID,
		fund.Name,
		fund.Description,
		fund.Type,
		fund.RiskLevel,
		fund.Performance,
		fund.TotalAmount,
		now,
		now,
	}

	var fundID string
	err := s.db.QueryRow(ctx, query, args...).Scan(&fundID)
	if err != nil {
		logger.WithError(err).Error("Failed to execute create fund query: %w", err)
		return "", fmt.Errorf("execute create fund query: %w", err)
	}

	logger.Info("Fund successfully created")
	return fundID, nil
}

func (s *Store) GetFund(ctx context.Context, id string) (*Fund, error) {
	logger := logrus.New().WithContext(ctx)

	logger = logger.WithField("fund_id", id)

	query := `SELECT id, name, description, type, risk_level, performance, total_amount, created_at, updated_at
		FROM funds WHERE id = $1`

	var fund Fund
	err := s.db.QueryRow(ctx, query, id).
		Scan(
			&fund.ID,
			&fund.Name,
			&fund.Description,
			&fund.Type,
			&fund.RiskLevel,
			&fund.Performance,
			&fund.TotalAmount,
			&fund.CreatedAt,
			&fund.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("Failed to find fund")
			return nil, fmt.Errorf("get fund: %w", err)
		}
		logger.WithError(err).Error("Failed to execute query for get fund")
		return nil, fmt.Errorf("failed to execute query for get fund: %w", err)
	}

	return &fund, nil
}

// UpdateFund updates the fund details i.e, name and description.
func (s *Store) UpdateFund(ctx context.Context, id, name, description string) (*Fund, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()

	logger = logger.WithFields(logrus.Fields{
		"fund_id": id,
	})

	query := `UPDATE funds
	SET name = $1, description = $2, updated_at = $3
	WHERE id = $4
	RETURNING id, name, description, type, risk_level, performance, total_amount, created_at, updated_at`

	args := []any{
		name,
		description,
		now,
		id,
	}

	var updatedFund Fund
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&updatedFund.ID,
		&updatedFund.Name,
		&updatedFund.Description,
		&updatedFund.Type,
		&updatedFund.RiskLevel,
		&updatedFund.Performance,
		&updatedFund.TotalAmount,
		&updatedFund.CreatedAt,
		&updatedFund.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("Fund not found for update")
			return nil, fmt.Errorf("fund not found: %w", err)
		}
		logger.WithError(err).Error("Failed to execute update fund query")
		return nil, fmt.Errorf("failed to execute update fund query: %w", err)
	}

	logger.Info("Fund successfully updated")
	return &updatedFund, nil
}

// UpdateFundTotalAmount updates the total amount in a fund
func (s *Store) UpdateFundTotalAmount(ctx context.Context, fundID string, totalAmount float64) (*Fund, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()

	logger = logger.WithFields(logrus.Fields{
		"fund_id":      fundID,
		"total_amount": totalAmount,
	})

	query := `UPDATE funds
	SET total_amount = $1, updated_at = $2
	WHERE id = $3
	RETURNING id, name, description, type, risk_level, performance, total_amount, created_at, updated_at`

	args := []any{
		totalAmount,
		now,
		fundID,
	}

	var updatedFund Fund
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&updatedFund.ID,
		&updatedFund.Name,
		&updatedFund.Description,
		&updatedFund.Type,
		&updatedFund.RiskLevel,
		&updatedFund.Performance,
		&updatedFund.TotalAmount,
		&updatedFund.CreatedAt,
		&updatedFund.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("Fund not found for update")
			return nil, fmt.Errorf("fund not found: %w", err)
		}
		logger.WithError(err).Error("Failed to execute update fund total amount query")
		return nil, fmt.Errorf("failed to execute update fund total amount query: %w", err)
	}

	logger.Info("Fund total amount successfully updated")
	return &updatedFund, nil
}

func (s *Store) ListFunds(ctx context.Context) ([]Fund, error) {
	logger := logrus.New().WithContext(ctx)

	query := `SELECT id, name, description, type, risk_level, performance, total_amount, created_at, updated_at FROM funds`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Failed to execute query for list funds")
		return nil, fmt.Errorf("failed to execute query for list funds: %w", err)
	}

	var funds []Fund
	for rows.Next() {
		var fund Fund
		if err := rows.Scan(
			&fund.ID,
			&fund.Name,
			&fund.Description,
			&fund.Type,
			&fund.RiskLevel,
			&fund.Performance,
			&fund.TotalAmount,
			&fund.CreatedAt,
			&fund.UpdatedAt,
		); err != nil {
			logger.WithError(err).Error("Failed to scan fund row")
			return nil, fmt.Errorf("failed to scan fund row: %w", err)
		}
		funds = append(funds, fund)
	}

	if err := rows.Err(); err != nil {
		logger.WithError(err).Error("Error iterating over fund rows")
		return nil, fmt.Errorf("error iterating over fund rows: %w", err)
	}

	return funds, nil

}

func (s *Store) CreateInvestment(ctx context.Context, investment Investment) (string, error) {
	logger := logrus.New().WithContext(ctx)
	now := time.Now()

	logger = logger.WithFields(logrus.Fields{
		"isa_id":  investment.ISAID,
		"fund_id": investment.FundID,
		"amount":  investment.Amount,
	})

	fmt.Println("ISA ID:", investment.ISAID, " Fund ID:", investment.FundID)

	query := `INSERT INTO investments (id, isa_id, fund_id, amount, invested_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	args := []any{
		investment.ID,
		investment.ISAID,
		investment.FundID,
		investment.Amount,
		now,
		now,
	}

	var investmentID string
	err := s.db.QueryRow(ctx, query, args...).Scan(&investmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to execute create investment query")
		return "", fmt.Errorf("execute create investment query: %w", err)
	}

	logger.Info("Investment successfully created")
	return investmentID, nil
}

func (s *Store) GetInvestment(ctx context.Context, investmentID string) (*Investment, error) {
	logger := logrus.New().WithContext(ctx)
	logger = logger.WithField("investment_id", investmentID)

	query := `SELECT id, isa_id, fund_id, amount, invested_at, created_at 
			  FROM investments WHERE id = $1`

	var investment Investment
	err := s.db.QueryRow(ctx, query, investmentID).Scan(
		&investment.ID,
		&investment.ISAID,
		&investment.FundID,
		&investment.Amount,
		&investment.InvestedAt,
		&investment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WithError(err).Error("Investment not found")
			return nil, fmt.Errorf("investment not found: %w", err)
		}
		logger.WithError(err).Error("Failed to execute get investment query")
		return nil, fmt.Errorf("execute get investment query: %w", err)
	}

	return &investment, nil
}
