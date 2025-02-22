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

var (
	fundsTable   = "funds"
	fundsColumns = []string{
		"id",
		"name",
		"description",
		"type",
		"risk_level",
		"performance",
		"total_amount",
	}
	investmentsTable   = "investments"
	investmentsColumns = []string{
		"id",
		"isa_id",
		"fund_id",
		"amount",
	}
	isasTable   = "isas"
	isasColumns = []string{
		"id",
		"user_id",
		"fund_ids",
		"cash_balance",
		"investment_amount",
	}
	usersTable   = "users"
	usersColumns = []string{
		"id",
		"first_name",
		"last_name",
		"email",
		"password",
	}
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
		"user_id":           isa.UserID,
		"fund_ids":          isa.FundIDs,
		"cash_balance":      isa.CashBalance,
		"investment_amount": isa.InvestmentAmount,
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
		logger.WithError(err).Error("Failed to execute create isa query: %w", err)
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
		"isa_id":            isa.ID,
		"user_id":           isa.UserID,
		"fund_ids":          isa.FundIDs,
		"cash_balance":      isa.CashBalance,
		"investment_amount": isa.InvestmentAmount,
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
