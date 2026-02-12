package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"backend-test-golang/internal/models"
	"backend-test-golang/pkg/database"
	errs "backend-test-golang/pkg/errors"

	"github.com/shopspring/decimal"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Withdraw(ctx context.Context, in models.WithdrawRequest) (*models.Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, "SELECT balance FROM users WHERE id = $1 FOR UPDATE", in.UserID).Scan(&currentBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get current balance for user(%d): %w", in.UserID, err)
	}

	newBalance := currentBalance.Sub(in.Amount)
	if newBalance.IsNegative() {
		return nil, fmt.Errorf("amount(%s) is greater than current balance(%s): %w", in.Amount.String(), currentBalance.String(), errs.ErrInsufficientBalance)
	}

	_, err = tx.Exec("UPDATE users SET balance = $1 WHERE id = $2", newBalance, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user(%d): %w", in.UserID, err)
	}

	row := tx.QueryRowContext(ctx,
		`
			INSERT INTO transactions (idempotency_key, user_id, balance_before, balance_after, amount, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			RETURNING id, idempotency_key, user_id, balance_before, balance_after, amount, created_at
		`, in.IdempotencyKey, in.UserID, currentBalance, newBalance, in.Amount)

	var txRecord models.Transaction
	err = row.Scan(
		&txRecord.ID,
		&txRecord.IdempotencyKey,
		&txRecord.UserID,
		&txRecord.BalanceBefore,
		&txRecord.BalanceAfter,
		&txRecord.Amount,
		&txRecord.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &txRecord, nil
}

func (r *Repository) GetUser(ctx context.Context, userID int64) (models.User, error) {
	var balance decimal.Decimal
	err := r.db.QueryRowContext(ctx, `select balance from users where id = $1`, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to get user(%d): %w", userID, err)
	}

	return models.User{
		ID:      userID,
		Balance: balance,
	}, nil
}

func (r *Repository) GetTransactions(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, idempotency_key, user_id, balance_before, balance_after, amount, created_at
		from transactions 
		where user_id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err = rows.Scan(&tx.ID,
			&tx.IdempotencyKey,
			&tx.UserID,
			&tx.BalanceBefore,
			&tx.BalanceAfter,
			&tx.Amount,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, &tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get transactions: rows.Err: %w", err)
	}

	return transactions, nil
}

func (r *Repository) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error) {
	var tx models.Transaction

	err := r.db.QueryRowContext(ctx, `
		select id, idempotency_key, user_id, balance_before, balance_after, amount, created_at
		from transactions 
		where idempotency_key = $1
	`, idempotencyKey).Scan(&tx.ID,
		&tx.IdempotencyKey,
		&tx.UserID,
		&tx.BalanceBefore,
		&tx.BalanceAfter,
		&tx.Amount,
		&tx.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return &tx, nil
}
