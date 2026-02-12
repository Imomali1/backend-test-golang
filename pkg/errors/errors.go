package errors

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalidCacheEntry   = errors.New("invalid cache entry")
	ErrValidationFailed    = errors.New("validation failed")
	ErrUserNotFound        = errors.New("user not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type ErrRateLimitExceed struct {
	RetryAfter time.Duration
}

func NewRateLimitExceedErr(retryAfter time.Duration) *ErrRateLimitExceed {
	return &ErrRateLimitExceed{RetryAfter: retryAfter}
}

func (e *ErrRateLimitExceed) Error() string {
	return "rate limit exceed"
}

func (e *ErrRateLimitExceed) ErrMsgWithRetry() string {
	return fmt.Sprintf("rate limit exceed: please retry after %v", e.RetryAfter)
}
