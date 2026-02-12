package models

import (
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	ID             int64           `json:"id"`
	IdempotencyKey string          `json:"idempotency_key"`
	UserID         int64           `json:"user_id"`
	BalanceBefore  decimal.Decimal `json:"balance_before"`
	BalanceAfter   decimal.Decimal `json:"balance_after"`
	Amount         decimal.Decimal `json:"amount"`
	CreatedAt      time.Time       `json:"created_at"`
}

type WithdrawResponse struct {
	HTTPCode    int          `json:"-"`
	Success     bool         `json:"success"`
	Message     string       `json:"message,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
}

type WithdrawRequest struct {
	IdempotencyKey string          `json:"-"`
	Amount         decimal.Decimal `json:"amount"`
	UserID         int64           `json:"user_id"`
}

func (r WithdrawRequest) Validate() error {
	if !r.Amount.IsPositive() {
		return errors.New("amount must be greater than zero")
	}

	if r.UserID <= 0 {
		return errors.New("invalid user id")
	}

	return nil
}

type Balance struct {
	UserID  int64           `json:"user_id"`
	Balance decimal.Decimal `json:"balance"`
}

type GetBalanceResponse struct {
	HTTPCode int             `json:"-"`
	Success  bool            `json:"success"`
	Message  string          `json:"message,omitempty"`
	Balance  decimal.Decimal `json:"balance,omitempty"`
	UserID   int64           `json:"user_id"`
}

type GetTransactionsResponse struct {
	HTTPCode     int            `json:"-"`
	Success      bool           `json:"success"`
	Message      string         `json:"message,omitempty"`
	Transactions []*Transaction `json:"transactions"`
}
