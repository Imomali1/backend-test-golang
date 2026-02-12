package models

import "github.com/shopspring/decimal"

type User struct {
	ID      int64           `json:"id"`
	Balance decimal.Decimal `json:"balance"`
}
