package services

import (
	"backend-test-golang/internal/models"
	errs "backend-test-golang/pkg/errors"
	"context"
	"errors"
	"log"
)

func (s *Service) Withdraw(ctx context.Context, in models.WithdrawRequest) (*models.Transaction, error) {
	if err := in.Validate(); err != nil {
		err = errors.Join(errs.ErrValidationFailed, err)
		log.Printf("validation error in withdraw: %v", err)
		return nil, err
	}

	tx, err := s.repo.GetTransactionByIdempotencyKey(ctx, in.IdempotencyKey)
	if err == nil {
		return tx, nil
	}

	withdrawal, err := s.repo.Withdraw(ctx, in)
	if err != nil {
		log.Printf("failed to withdraw: %v", err)
		return nil, err
	}

	return withdrawal, nil
}

func (s *Service) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		log.Printf("failed to get user: %v", err)
		return models.Balance{}, err
	}

	return models.Balance{
		UserID:  userID,
		Balance: user.Balance,
	}, nil
}

func (s *Service) GetTransactions(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	txs, err := s.repo.GetTransactions(ctx, userID)
	if err != nil {
		log.Printf("failed to get transactions: %v", err)
		return nil, err
	}

	return txs, nil
}
