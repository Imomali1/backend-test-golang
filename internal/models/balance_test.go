package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestWithdrawRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     WithdrawRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid positive amount",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(100.50),
				UserID: 1,
			},
			wantErr: false,
		},
		{
			name: "valid small amount",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(0.01),
				UserID: 1,
			},
			wantErr: false,
		},
		{
			name: "zero amount should fail",
			req: WithdrawRequest{
				Amount: decimal.Zero,
				UserID: 1,
			},
			wantErr: true,
			errMsg:  "amount must be greater than zero",
		},
		{
			name: "negative amount should fail",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(-10.50),
				UserID: 1,
			},
			wantErr: true,
			errMsg:  "amount must be greater than zero",
		},
		{
			name: "zero user_id should fail",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(100),
				UserID: 0,
			},
			wantErr: true,
			errMsg:  "invalid user id",
		},
		{
			name: "negative user_id should fail",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(100),
				UserID: -1,
			},
			wantErr: true,
			errMsg:  "invalid user id",
		},
		{
			name: "large amount should pass",
			req: WithdrawRequest{
				Amount: decimal.NewFromFloat(999999.99),
				UserID: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected validation error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validation error: %v, want: %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("got unexpected error = %v", err)
				}
			}
		})
	}
}

func TestWithdrawRequest_Validate_EdgeCases(t *testing.T) {
	t.Run("very large decimal precision", func(t *testing.T) {
		req := WithdrawRequest{
			Amount: decimal.NewFromFloat(0.0000001),
			UserID: 1,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected very small positive amount to pass, got error: %v", err)
		}
	})

	t.Run("very large integer user_id", func(t *testing.T) {
		req := WithdrawRequest{
			Amount: decimal.NewFromFloat(100),
			UserID: 9999999999999999,
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("Expected large user_id to pass, got error: %v", err)
		}
	})
}
