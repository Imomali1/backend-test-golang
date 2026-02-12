package handlers

import (
	errs "backend-test-golang/pkg/errors"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend-test-golang/internal/models"
	"backend-test-golang/internal/services"

	"github.com/google/uuid"
)

type Handler struct {
	defaultTimeout time.Duration
	svc            *services.Service
}

func New(svc *services.Service) *Handler {
	return &Handler{
		svc:            svc,
		defaultTimeout: 10 * time.Second,
	}
}

func (h *Handler) GetItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respond(w, http.StatusMethodNotAllowed, models.Response{Message: "method not allowed"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	items, err := h.svc.GetItems(ctx)
	if err != nil {
		var e *errs.ErrRateLimitExceed
		if errors.As(err, &e) {
			w.Header().Set("Retry-After", strconv.FormatUint(uint64(e.RetryAfter.Seconds()), 10))
			respond(w, http.StatusTooManyRequests, models.Response{Message: err.Error()})
			return
		}

		respond(w, http.StatusInternalServerError, models.Response{Message: "internal server error"})
		return
	}

	respond(w, http.StatusOK, models.Response{
		Success: true,
		Payload: items,
	})
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respond(w, http.StatusMethodNotAllowed, models.Response{Message: "method not allowed"})
		return
	}

	idempotencyKey := r.Header.Get("X-Idempotency-Key")
	if idempotencyKey == "" {
		respond(w, http.StatusBadRequest, models.Response{Message: "no idempotency key provided"})
		return
	}

	if uuid.Validate(idempotencyKey) != nil {
		respond(w, http.StatusBadRequest, models.Response{Message: "invalid idempotency key provided"})
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, models.Response{Message: err.Error()})
		return
	}

	req.IdempotencyKey = idempotencyKey

	ctx, cancel := context.WithTimeout(r.Context(), h.defaultTimeout)
	defer cancel()

	withdrawal, err := h.svc.Withdraw(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUserNotFound):
			respond(w, http.StatusNotFound, models.Response{Message: "user not found"})
		case errors.Is(err, errs.ErrInsufficientBalance):
			respond(w, http.StatusBadRequest, models.Response{Message: "insufficient balance"})
		default:
			respond(w, http.StatusInternalServerError, models.Response{Message: "internal server error"})
		}
		return
	}

	respond(w, http.StatusOK, models.Response{
		Success: true,
		Payload: withdrawal,
	})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respond(w, http.StatusMethodNotAllowed, models.Response{Message: "method not allowed"})
		return
	}

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		respond(w, http.StatusBadRequest, models.Response{Message: "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.defaultTimeout)
	defer cancel()

	balance, err := h.svc.GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			respond(w, http.StatusNotFound, models.Response{Message: "user not found"})
			return
		}

		respond(w, http.StatusInternalServerError, models.Response{Message: err.Error()})
		return
	}

	respond(w, http.StatusOK, models.Response{
		Success: true,
		Payload: balance,
	})
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respond(w, http.StatusMethodNotAllowed, models.Response{Message: "method not allowed"})
		return
	}

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		respond(w, http.StatusBadRequest, models.Response{Message: "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.defaultTimeout)
	defer cancel()

	txs, err := h.svc.GetTransactions(ctx, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			respond(w, http.StatusRequestTimeout, models.Response{Message: "context deadline exceeded"})
			return
		}

		if errors.Is(err, errs.ErrNotFound) {
			respond(w, http.StatusNotFound, models.Response{Message: "transactions not found"})
			return
		}

		respond(w, http.StatusInternalServerError, models.Response{Message: "internal server error"})
		return
	}

	respond(w, http.StatusOK, models.Response{
		Success: true,
		Payload: txs,
	})
}

func respond(w http.ResponseWriter, httpCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(v)
}
